# 第6章: S3バックアップ機能

この章では、FuckBaseのS3バックアップ機能について説明します。FuckBaseは、Amazon S3やMinIOなどのS3互換ストレージとの連携機能を提供しており、データベースのバックアップと復元が可能です。

## S3バックアップの概要

S3バックアップ機能は、FuckBaseのデータを外部のS3互換ストレージに保存し、必要に応じて復元するための機能です。この機能により、以下のようなメリットがあります：

1. データの永続化
2. サーバー障害からの復旧
3. データの移行
4. バージョン管理

```mermaid
graph TD
    A[FuckBase] -->|バックアップ| B[S3コネクタ]
    B -->|アップロード| C[S3互換ストレージ]
    C -->|ダウンロード| B
    B -->|復元| A
```

S3バックアップ機能の実装は [../internal/s3/backup.go](../internal/s3/backup.go) と [../internal/s3/client.go](../internal/s3/client.go) で確認できます。

## S3クライアント

FuckBaseは、AWS SDK for Goを使用してS3 APIと通信します。S3クライアントは、以下の主要な操作をサポートしています：

1. ファイルのアップロード
2. ファイルのダウンロード
3. ファイルの一覧取得
4. ファイルの削除

```mermaid
classDiagram
    class Client {
        +s3Client *s3.Client
        +bucket string
        +UploadFile(objectName string, data []byte, contentType string) error
        +DownloadFile(objectName string) ([]byte, error)
        +ListFiles(prefix string) ([]string, error)
        +DeleteFile(objectName string) error
    }
```

## バックアップマネージャ

バックアップマネージャは、S3クライアントとデータベースマネージャを連携させ、バックアップと復元の処理を行うコンポーネントです。

```mermaid
classDiagram
    class BackupManager {
        +s3Client *Client
        +dbManager *database.Manager
        +BackupDatabase(dbName string) error
        +BackupAllDatabases() error
        +RestoreDatabase(objectName string) error
        +RestoreAllDatabases(objectName string) error
        +ListBackups() ([]string, error)
    }
```

## バックアップデータ構造

バックアップデータは、JSONフォーマットで構造化されています。主なデータ構造は以下の通りです：

```mermaid
classDiagram
    class FullBackup {
        +Metadata BackupMetadata
        +Databases map[string]DatabaseBackup
    }
    
    class BackupMetadata {
        +Timestamp time.Time
        +Version string
        +DatabaseCount int
        +SetCount int
        +EntryCount int
    }
    
    class DatabaseBackup {
        +Name string
        +Sets map[string]SetBackup
        +Indexes map[string]IndexBackup
        +Auth *database.AuthConfig
    }
    
    class SetBackup {
        +Name string
        +Data map[string]interface{}
    }
    
    class IndexBackup {
        +Name string
        +SetName string
        +Field string
    }
    
    FullBackup --> BackupMetadata
    FullBackup --> DatabaseBackup
    DatabaseBackup --> SetBackup
    DatabaseBackup --> IndexBackup
```

## バックアップ処理フロー

バックアップ処理は、以下の手順で行われます：

1. データベースの読み取りロックを取得
2. データベースの状態をシリアライズ
3. S3にアップロード
4. 読み取りロックを解放

```mermaid
sequenceDiagram
    participant Client as クライアント
    participant Handler as APIハンドラ
    participant BM as バックアップマネージャ
    participant DB as データベース
    participant S3 as S3クライアント
    
    Client->>Handler: バックアップリクエスト
    Handler->>BM: BackupDatabase(dbName)
    BM->>DB: GetDatabase(dbName)
    DB-->>BM: データベースオブジェクト
    
    BM->>BM: createDatabaseBackup(db)
    
    loop 各Set
        BM->>DB: GetSet(setName)
        DB-->>BM: Setオブジェクト
        BM->>BM: バックアップデータに追加
    end
    
    loop 各インデックス
        BM->>DB: GetIndex(indexName)
        DB-->>BM: インデックスオブジェクト
        BM->>BM: バックアップデータに追加
    end
    
    BM->>BM: JSON変換
    BM->>S3: UploadFile(objectName, data)
    S3-->>BM: アップロード結果
    BM-->>Handler: バックアップ結果
    Handler-->>Client: レスポンス
```

## 復元処理フロー

復元処理は、以下の手順で行われます：

1. S3からバックアップデータをダウンロード
2. JSONデータをパース
3. 既存のデータベースを削除（存在する場合）
4. 新しいデータベースを作成
5. Setとインデックスを復元

```mermaid
sequenceDiagram
    participant Client as クライアント
    participant Handler as APIハンドラ
    participant BM as バックアップマネージャ
    participant DB as データベースマネージャ
    participant S3 as S3クライアント
    
    Client->>Handler: 復元リクエスト
    Handler->>BM: RestoreDatabase(objectName)
    BM->>S3: DownloadFile(objectName)
    S3-->>BM: バックアップデータ
    
    BM->>BM: JSONパース
    
    alt データベースが存在する
        BM->>DB: DeleteDatabase(backup.Name)
    end
    
    BM->>DB: CreateDatabase(backup.Name, backup.Auth)
    DB-->>BM: 新しいデータベース
    
    loop 各Set
        BM->>DB: CreateSet(setBackup.Name)
        DB-->>BM: 新しいSet
        
        loop 各キーバリューペア
            BM->>DB: Put(key, value)
        end
    end
    
    loop 各インデックス
        BM->>DB: CreateIndex(indexBackup.Name, indexBackup.SetName, indexBackup.Field)
    end
    
    BM-->>Handler: 復元結果
    Handler-->>Client: レスポンス
```

## 自動バックアップ

FuckBaseは、定期的な自動バックアップをサポートしています。バックアップ間隔は、サーバー設定で指定できます。

```mermaid
graph TD
    A[サーバー起動] --> B{自動バックアップが有効?}
    B -->|はい| C[バックアップスケジューラ開始]
    B -->|いいえ| D[スキップ]
    
    C --> E[バックアップ間隔待機]
    E --> F[すべてのデータベースをバックアップ]
    F --> E
```

自動バックアップの実装は [../internal/server/server.go](../internal/server/server.go) の `startAutomaticBackups` メソッドで確認できます。

## バックアップAPIエンドポイント

FuckBaseは、バックアップと復元のための以下のAPIエンドポイントを提供しています：

1. `/backup/create`: バックアップを作成
2. `/backup/list`: バックアップの一覧を取得
3. `/backup/restore`: バックアップから復元

これらのエンドポイントは、S3が有効な場合にのみ利用可能です。

### バックアップ作成 (/backup/create)

指定されたデータベースのバックアップを作成します。

**リクエスト例**:
```json
{
  "database": "mydb"
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "message": "Backup created successfully",
  "data": {
    "object_name": "backups/mydb/2023-01-01T12:00:00Z.json"
  }
}
```

### バックアップ一覧取得 (/backup/list)

利用可能なバックアップの一覧を取得します。

**リクエスト例**:
```json
{
  "database": "mydb"
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "data": {
    "backups": [
      "backups/mydb/2023-01-01T12:00:00Z.json",
      "backups/mydb/2023-01-02T12:00:00Z.json"
    ]
  }
}
```

### バックアップ復元 (/backup/restore)

指定されたバックアップからデータベースを復元します。

**リクエスト例**:
```json
{
  "object_name": "backups/mydb/2023-01-01T12:00:00Z.json"
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "message": "Database restored successfully"
}
```

## S3設定

FuckBaseのS3連携機能を使用するには、以下の設定が必要です：

```mermaid
graph TD
    A[S3設定] --> B[エンドポイント]
    A --> C[バケット名]
    A --> D[アクセスキー]
    A --> E[シークレットキー]
    A --> F[リージョン]
    A --> G[有効/無効]
```

これらの設定は、コマンドラインオプションまたは環境変数で指定できます。

```bash
fuckbase --s3-endpoint https://s3.amazonaws.com --s3-bucket my-backup-bucket --s3-access-key ACCESS_KEY --s3-secret-key SECRET_KEY --s3-region us-east-1
```

## S3互換ストレージ

FuckBaseは、以下のようなS3互換ストレージと連携できます：

1. Amazon S3
2. MinIO
3. Wasabi
4. Backblaze B2
5. その他のS3互換ストレージ

## まとめ

FuckBaseのS3バックアップ機能は、データの永続化と障害復旧のための重要な機能です。S3互換ストレージとの連携により、データを安全に保存し、必要に応じて復元することができます。自動バックアップ機能を使用すれば、定期的なバックアップを自動化することも可能です。

次の章では、FuckBaseの実践的な使用例について見ていきます。