# 第5章: HTTPサーバーとAPI

この章では、FuckBaseのHTTPサーバーとAPIの実装について説明します。FuckBaseはHTTPベースのデータベースであり、すべての操作はHTTP APIを通じて行われます。

## HTTPサーバーの概要

FuckBaseのHTTPサーバーは、Goの標準ライブラリ（`net/http`）を使用して実装されています。サーバーは指定されたポートでリクエストを待ち受け、各APIエンドポイントに対応するハンドラ関数を呼び出します。

```mermaid
graph TD
    A("クライアント") -->|"HTTPリクエスト"| B("HTTPサーバー")
    B -->|"ルーティング"| C("APIハンドラ")
    C -->|"データ操作"| D("データベースマネージャ")
    D --> E("データベース")
    E --> F("Set/インデックス")
    F -->|"結果"| G("レスポンス生成")
    G -->|"HTTPレスポンス"| A
```

サーバーの実装は [../internal/server/server.go](../internal/server/server.go) で確認できます。

## サーバーの起動と設定

FuckBaseサーバーは、コマンドラインオプションや環境変数から設定を読み込み、HTTPサーバーを起動します。主な設定項目には、ポート番号、ホスト名、管理者認証情報、S3バックアップ設定などがあります。

```mermaid
sequenceDiagram
    participant Main as main()
    participant Config as "設定マネージャ"
    participant Server as "HTTPサーバー"
    participant DBManager as "データベースマネージャ"
    
    Main->>Config: "コマンドライン引数と環境変数を解析"
    Config-->>Main: ServerConfig
    Main->>DBManager: NewManager()
    Main->>Server: NewServer(config, dbManager)
    Main->>Server: Start()
    Server->>Server: registerEndpoints()
    Server->>Server: ListenAndServe()
```

## APIエンドポイント

FuckBaseは、以下のカテゴリのAPIエンドポイントを提供しています：

1. データベース管理
2. Set操作
3. インデックス操作
4. サーバー情報
5. バックアップと復元（S3が有効な場合）

各エンドポイントの実装は [../internal/server/handlers.go](../internal/server/handlers.go) で確認できます。

### エンドポイント一覧

```mermaid
graph TD
    A("FuckBase API") --> B("データベース管理")
    A --> C("Set操作")
    A --> D("インデックス操作")
    A --> E("サーバー情報")
    A --> F("バックアップ/復元")
    
    B --> B1("/create")
    B --> B2("/drop")
    
    C --> C1("/set/create")
    C --> C2("/set/get")
    C --> C3("/set/put")
    C --> C4("/set/delete")
    C --> C5("/set/list")
    
    D --> D1("/index/create")
    D --> D2("/index/drop")
    D --> D3("/index/query")
    
    E --> E1("/server/info")
    
    F --> F1("/backup/create")
    F --> F2("/backup/list")
    F --> F3("/backup/restore")
```

## リクエスト処理フロー

FuckBaseのリクエスト処理フローは、以下の一般的なパターンに従います：

1. HTTPリクエストを受信
2. リクエストメソッドの検証（POSTのみ許可）
3. 認証チェック（必要な場合）
4. リクエストボディの解析
5. リクエストパラメータの検証
6. データベース操作の実行
7. レスポンスの生成と送信
8. ログの記録

```mermaid
sequenceDiagram
    participant Client as "クライアント"
    participant Handler as "APIハンドラ"
    participant Auth as "認証マネージャ"
    participant DB as "データベース"
    participant Logger as "ロガー"
    
    Client->>Handler: "HTTPリクエスト"
    Handler->>Handler: "メソッド検証（POSTのみ）"
    
    alt "認証が必要"
        Handler->>Auth: "認証情報の検証"
        Auth-->>Handler: "認証結果"
        
        alt "認証失敗"
            Handler-->>Client: "401 Unauthorized"
            Handler->>Logger: "ログ記録"
        end
    end
    
    Handler->>Handler: "リクエストボディ解析"
    Handler->>Handler: "パラメータ検証"
    
    alt "パラメータ無効"
        Handler-->>Client: "400 Bad Request"
        Handler->>Logger: "ログ記録"
    end
    
    Handler->>DB: "データベース操作"
    DB-->>Handler: "操作結果"
    
    alt "操作失敗"
        Handler-->>Client: "エラーレスポンス"
    else "操作成功"
        Handler-->>Client: "成功レスポンス"
    end
    
    Handler->>Logger: "ログ記録"
```

## 主要なAPIエンドポイントの詳細

### 1. データベース作成 (/create)

データベースを作成するエンドポイントです。オプションで認証情報を設定できます。

**リクエスト例**:
```json
{
  "name": "mydb",
  "auth": {
    "username": "user",
    "password": "pass"
  }
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "message": "Database created successfully",
  "data": {
    "database": "mydb"
  }
}
```

### 2. データ保存 (/set/put)

Setにデータを保存するエンドポイントです。

**リクエスト例**:
```json
{
  "database": "mydb",
  "set": "users",
  "key": "user1",
  "value": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "message": "Data stored successfully"
}
```

### 3. データ取得 (/set/get)

Setからデータを取得するエンドポイントです。

**リクエスト例**:
```json
{
  "database": "mydb",
  "set": "users",
  "key": "user1"
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "data": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

### 4. インデックス作成 (/index/create)

Setのフィールドにインデックスを作成するエンドポイントです。

**リクエスト例**:
```json
{
  "database": "mydb",
  "set": "users",
  "name": "email_index",
  "field": "email"
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "message": "Index created successfully",
  "data": {
    "index": "email_index"
  }
}
```

### 5. インデックスクエリ (/index/query)

インデックスを使用してデータをクエリするエンドポイントです。

**リクエスト例**:
```json
{
  "database": "mydb",
  "set": "users",
  "index": "email_index",
  "value": "john@example.com"
}
```

**レスポンス例**:
```json
{
  "status": "success",
  "data": {
    "count": 1,
    "data": [
      {
        "key": "user1",
        "value": {
          "name": "John Doe",
          "email": "john@example.com"
        }
      }
    ]
  }
}
```

## エラーハンドリング

FuckBaseは、エラーが発生した場合、適切なHTTPステータスコードとエラーメッセージを含むJSONレスポンスを返します。

```json
{
  "status": "error",
  "code": "DB_NOT_FOUND",
  "message": "Database not found"
}
```

主なエラーコードには以下のようなものがあります：

- `INVALID_REQUEST`: リクエストの形式が無効
- `METHOD_NOT_ALLOWED`: POSTメソッド以外のリクエスト
- `AUTH_FAILED`: 認証失敗
- `DB_NOT_FOUND`: データベースが見つからない
- `SET_NOT_FOUND`: Setが見つからない
- `KEY_NOT_FOUND`: キーが見つからない
- `INDEX_NOT_FOUND`: インデックスが見つからない
- `INTERNAL_ERROR`: 内部エラー

## 認証

FuckBaseは、2種類の認証をサポートしています：

1. **管理者認証**: データベースの作成や削除などの管理操作に必要
2. **データベース認証**: 特定のデータベースへのアクセスに必要

認証情報は、HTTPヘッダーまたはリクエストボディで提供できます。

```mermaid
graph TD
    A("リクエスト") --> B{"管理操作?"}
    B -->|"はい"| C{"管理者認証が有効?"}
    B -->|"いいえ"| F{"データベース認証が有効?"}
    
    C -->|"はい"| D("管理者認証チェック")
    C -->|"いいえ"| E("認証不要")
    
    D -->|"成功"| E
    D -->|"失敗"| K("401 Unauthorized")
    
    F -->|"はい"| G("データベース認証チェック")
    F -->|"いいえ"| J("認証不要")
    
    G -->|"成功"| J
    G -->|"失敗"| K
    
    E --> H("操作実行")
    J --> H
    
    H --> I("レスポンス")
```

## ロギング

FuckBaseは、すべてのリクエストとレスポンスをログに記録します。ログには、リクエストのメソッド、パス、ステータスコード、処理時間などの情報が含まれます。

```
INFO: POST /set/put 200 1.2ms
```

ログの実装は [../internal/logger/logger.go](../internal/logger/logger.go) で確認できます。

## まとめ

FuckBaseのHTTPサーバーとAPIは、シンプルながらも効率的に設計されています。すべての操作はHTTP POSTリクエストを通じて行われ、JSONレスポンスが返されます。認証、エラーハンドリング、ロギングなどの機能も備えており、実用的なデータベースサーバーとして機能します。

次の章では、FuckBaseのS3バックアップ機能について詳しく見ていきます。