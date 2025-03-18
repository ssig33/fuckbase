# 第2章: アーキテクチャ概要

この章では、FuckBaseの全体的なアーキテクチャと主要なコンポーネントについて説明します。

## 全体構造

FuckBaseは、以下の主要コンポーネントで構成されています：

```mermaid
graph TD
    A[HTTPサーバー] --> B[データベースマネージャ]
    A --> C[認証マネージャ]
    A --> D[S3コネクタ]
    B --> E[データベース1]
    B --> F[データベース2]
    E --> G[Set1]
    E --> H[Set2]
    E --> I[インデックス]
    F --> J[Set3]
    F --> K[インデックス]
```

この構造は階層的になっており、上位のコンポーネントが下位のコンポーネントを管理する形になっています。例えば、データベースマネージャは複数のデータベースを管理し、各データベースは複数のSetとインデックスを管理します。

## コアコンポーネント

### 1. HTTPサーバー

HTTPサーバーは、FuckBaseの外部インターフェースとなるコンポーネントです。クライアントからのHTTPリクエストを受け付け、適切なコンポーネントに転送します。

```mermaid
sequenceDiagram
    participant Client as クライアント
    participant Server as HTTPサーバー
    participant Router as ルーター
    participant Handler as ハンドラ
    
    Client->>Server: HTTPリクエスト
    Server->>Router: リクエスト転送
    Router->>Handler: 適切なハンドラにルーティング
    Handler->>Handler: リクエスト処理
    Handler-->>Client: HTTPレスポンス
```

HTTPサーバーの実装は [../internal/server/server.go](../internal/server/server.go) で確認できます。

### 2. データベースマネージャ

データベースマネージャは、複数のデータベースインスタンスを管理するコンポーネントです。データベースの作成、削除、アクセス制御などを担当します。

```mermaid
classDiagram
    class Manager {
        +Databases map[string]*Database
        +mu sync.RWMutex
        +CreateDatabase(name string, auth *AuthConfig) *Database
        +GetDatabase(name string) *Database
        +DeleteDatabase(name string) error
        +ListDatabases() []string
    }
```

データベースマネージャの実装は [../internal/database/manager.go](../internal/database/manager.go) で確認できます。

### 3. データベース

データベースは、複数のSetとインデックスを管理するコンポーネントです。各データベースは独立しており、他のデータベースとデータを共有することはありません。

```mermaid
classDiagram
    class Database {
        +Name string
        +Sets map[string]*Set
        +Indexes map[string]*Index
        +Auth *AuthConfig
        +mu sync.RWMutex
        +CreateSet(name string) *Set
        +GetSet(name string) *Set
        +CreateIndex(name string, setName string, field string) *Index
        +GetIndex(name string) *Index
    }
```

データベースの実装は [../internal/database/database.go](../internal/database/database.go) で確認できます。

### 4. Set

Setは、キーバリューペアを保存するコンポーネントです。内部的には、Goのマップ（`map[string][]byte`）を使用しています。値はMessagePackでエンコードされたバイト配列として保存されます。

```mermaid
classDiagram
    class Set {
        +Name string
        +Data map[string][]byte
        +mu sync.RWMutex
        +Put(key string, value interface{}) error
        +Get(key string, dest interface{}) error
        +Delete(key string) error
        +Keys() []string
    }
```

Setの実装は [../internal/database/set.go](../internal/database/set.go) で確認できます。

### 5. インデックス

インデックスは、Setのデータに対する二次インデックスを提供するコンポーネントです。特定のフィールドの値からキーのリストへのマッピングを管理します。

```mermaid
classDiagram
    class Index {
        +Name string
        +SetName string
        +Field string
        +Values map[string][]string
        +mu sync.RWMutex
        +Build(set *Set) error
        +Query(value string, sortOrder string, limit, offset int) ([]string, error)
        +AddEntry(key string, value []byte) error
        +RemoveEntry(key string, value []byte) error
    }
```

インデックスの実装は [../internal/database/index.go](../internal/database/index.go) で確認できます。

### 6. S3コネクタ

S3コネクタは、S3互換ストレージとの連携を担当するコンポーネントです。バックアップと復元の処理を行います。

```mermaid
classDiagram
    class Client {
        +s3Client *s3.Client
        +bucket string
        +UploadFile(objectName string, data []byte, contentType string) error
        +DownloadFile(objectName string) ([]byte, error)
        +ListFiles(prefix string) ([]string, error)
    }
    
    class BackupManager {
        +s3Client *Client
        +dbManager *database.Manager
        +BackupDatabase(dbName string) error
        +RestoreDatabase(objectName string) error
    }
```

S3コネクタの実装は [../internal/s3/client.go](../internal/s3/client.go) と [../internal/s3/backup.go](../internal/s3/backup.go) で確認できます。

### 7. 認証マネージャ

認証マネージャは、ユーザー認証とアクセス制御を担当するコンポーネントです。管理者認証と一般ユーザー認証を区別します。

```mermaid
classDiagram
    class AdminAuth {
        +Config *config.AdminAuthConfig
        +RequireAdminAuth(handler http.HandlerFunc) http.HandlerFunc
    }
```

認証マネージャの実装は [../internal/server/auth.go](../internal/server/auth.go) で確認できます。

## データフロー

FuckBaseにおけるデータフローを理解することは、システム全体の動作を理解する上で重要です。以下に、主要な操作におけるデータフローを示します。

### データ保存フロー

```mermaid
sequenceDiagram
    participant Client as クライアント
    participant Server as HTTPサーバー
    participant DBManager as データベースマネージャ
    participant DB as データベース
    participant Set as Set
    participant Index as インデックス
    
    Client->>Server: PUT リクエスト
    Server->>DBManager: GetDatabase(dbName)
    DBManager-->>Server: データベース
    Server->>DB: GetSet(setName)
    DB-->>Server: Set
    Server->>Set: Put(key, value)
    Set->>Set: MessagePackエンコード
    Set-->>Server: 成功
    
    alt インデックスが存在する場合
        Server->>DB: GetIndex(indexName)
        DB-->>Server: インデックス
        Server->>Set: GetRaw(key)
        Set-->>Server: エンコードされた値
        Server->>Index: AddEntry(key, rawValue)
        Index-->>Server: 成功
    end
    
    Server-->>Client: 成功レスポンス
```

### データ取得フロー

```mermaid
sequenceDiagram
    participant Client as クライアント
    participant Server as HTTPサーバー
    participant DBManager as データベースマネージャ
    participant DB as データベース
    participant Set as Set
    
    Client->>Server: GET リクエスト
    Server->>DBManager: GetDatabase(dbName)
    DBManager-->>Server: データベース
    Server->>DB: GetSet(setName)
    DB-->>Server: Set
    Server->>Set: Get(key, &value)
    Set->>Set: MessagePackデコード
    Set-->>Server: デコードされた値
    Server-->>Client: 値を含むレスポンス
```

### インデックスクエリフロー

```mermaid
sequenceDiagram
    participant Client as クライアント
    participant Server as HTTPサーバー
    participant DBManager as データベースマネージャ
    participant DB as データベース
    participant Index as インデックス
    participant Set as Set
    
    Client->>Server: QUERY リクエスト
    Server->>DBManager: GetDatabase(dbName)
    DBManager-->>Server: データベース
    Server->>DB: GetIndex(indexName)
    DB-->>Server: インデックス
    Server->>Index: Query(value, sortOrder, limit, offset)
    Index-->>Server: キーのリスト
    
    loop 各キー
        Server->>Set: Get(key, &value)
        Set-->>Server: 値
    end
    
    Server-->>Client: 結果を含むレスポンス
```

## コンポーネント間の相互作用

FuckBaseのコンポーネントは、以下のように相互作用します：

1. **HTTPサーバー → データベースマネージャ**: HTTPサーバーは、クライアントからのリクエストを受け取り、データベースマネージャに対して操作を要求します。

2. **データベースマネージャ → データベース**: データベースマネージャは、リクエストに応じて適切なデータベースを選択し、操作を転送します。

3. **データベース → Set/インデックス**: データベースは、リクエストに応じて適切なSetやインデックスを選択し、操作を転送します。

4. **HTTPサーバー → 認証マネージャ**: HTTPサーバーは、認証が必要なリクエストを認証マネージャに転送し、認証の結果に応じて処理を続行するかどうかを決定します。

5. **HTTPサーバー → S3コネクタ**: バックアップや復元のリクエストがあった場合、HTTPサーバーはS3コネクタに操作を要求します。

## まとめ

FuckBaseは、シンプルながらも効率的なアーキテクチャを持っています。各コンポーネントは明確な責任を持ち、相互に連携して動作します。この設計により、シンプルさを保ちながらも、必要な機能を提供することができています。

次の章では、FuckBaseの主要なデータ構造について詳しく見ていきます。