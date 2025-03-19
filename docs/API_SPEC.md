# FuckBase API 仕様書

このドキュメントでは、FuckBaseのHTTP APIインターフェースについて詳細に説明します。

## 基本情報

- すべてのAPIリクエストはHTTP POSTメソッドを使用します
- リクエスト/レスポンスのボディはJSON形式です
- 認証が必要な場合は、リクエストヘッダーに認証情報を含めます

## エンドポイント一覧

### データベース管理

#### データベース作成

```
POST /create
```

**リクエスト**:
```json
{
  "name": "my_database",
  "auth": {
    "username": "admin",
    "password": "secure_password"
  }
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Database created successfully",
  "database": "my_database"
}
```

**注意**: サーバーに管理ユーザーが設定されている場合、このエンドポイントには管理者認証が必要です。

#### データベース削除

```
POST /drop
```

**リクエスト**:
```json
{
  "name": "my_database"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Database dropped successfully"
}
```

**注意**: サーバーに管理ユーザーが設定されている場合、このエンドポイントには管理者認証が必要です。

### Set操作

#### Set作成

```
POST /set/create
```

**リクエスト**:
```json
{
  "database": "my_database",
  "name": "users"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Set created successfully",
  "set": "users"
}
```

#### データ挿入/更新

```
POST /set/put
```

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "key": "user123",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }
}
```

**注意**: `value`フィールドはサーバー側でMessagePackにエンコードされます。

**レスポンス**:
```json
{
  "status": "success",
  "message": "Data stored successfully"
}
```

#### データ取得

```
POST /set/get
```

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "key": "user123"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "data": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }
}
```

**注意**: レスポンスの`data`フィールドはMessagePackからデコードされた値です。

#### データ削除

```
POST /set/delete
```

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "key": "user123"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Data deleted successfully"
}
```

#### Set一覧取得

```
POST /set/list
```

**リクエスト**:
```json
{
  "database": "my_database"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "sets": ["users", "products", "orders"]
}
```

### インデックス操作

#### 基本インデックス作成

```
POST /index/create
```

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "name": "email_index",
  "field": "email"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Index created successfully",
  "index": "email_index"
}
```

#### ソート可能インデックス作成

```
POST /index/create/sortable
```

**説明**:
このエンドポイントは、ソート可能インデックスを作成します。ソート可能インデックスは、プライマリフィールド（フィルタリング用）と1つ以上のソートフィールド（ソート用）を持ちます。これにより、「あるカラムの値で絞った上で、他のカラムでソート」という操作が効率的に実行できます。

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "name": "department_hiredate_index",
  "primary_field": "department",   // フィルタリングに使用するフィールド
  "sort_fields": ["hireDate", "name"] // ソートに使用するフィールド（複数指定可能）
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Sortable index created successfully",
  "index": "department_hiredate_index"
}
```

**注意**:
- `primary_field`は必須で、このフィールドの値に基づいてデータがフィルタリングされます。
- `sort_fields`は1つ以上のフィールドを指定でき、これらのフィールドに基づいてデータがソートされます。
- 複数のソートフィールドを指定した場合、最初のフィールドで同じ値を持つエントリは、次のフィールドでソートされます。
- インデックス作成時に、指定されたフィールドが存在しないエントリは、そのフィールドについてはインデックスに追加されません。

#### 基本インデックスによるクエリ

```
POST /index/query
```

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "index": "email_index",
  "value": "john@example.com"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "count": 1,
  "data": [
    {
      "key": "user123",
      "value": {
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30
      }
    }
  ]
}
```

#### ソート可能インデックスによるクエリ

```
POST /index/query/sorted
```

**説明**:
このエンドポイントは、ソート可能インデックスを使用して、プライマリフィールドの値でデータをフィルタリングし、指定されたソートフィールドでソートした結果を返します。例えば、「部署が"営業"の従業員を、入社日の新しい順に取得する」といったクエリが可能です。

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "index": "department_hiredate_index",
  "value": "sales",           // プライマリフィールド（department）の値
  "sort": {
    "field": "hireDate",      // ソートフィールド
    "order": "desc"           // ソート順序: "asc"（昇順）または "desc"（降順）
  },
  "pagination": {
    "offset": 0,              // 結果セットの開始位置
    "limit": 10               // 取得する最大件数
  }
}
```

**レスポンス**:
```json
{
  "status": "success",
  "count": 3,                 // 返された結果の件数
  "total": 3,                 // フィルタリング条件に一致する総件数
  "offset": 0,
  "limit": 10,
  "data": [
    {
      "key": "user456",
      "value": {
        "name": "Jane Smith",
        "department": "sales",
        "hireDate": "2023-05-15",
        "position": "Sales Manager"
      }
    },
    {
      "key": "user789",
      "value": {
        "name": "Bob Johnson",
        "department": "sales",
        "hireDate": "2022-11-03",
        "position": "Sales Representative"
      }
    },
    {
      "key": "user123",
      "value": {
        "name": "John Doe",
        "department": "sales",
        "hireDate": "2021-08-20",
        "position": "Sales Director"
      }
    }
  ]
}
```

#### 複数条件ソートによるクエリ

```
POST /index/query/multi-sorted
```

**説明**:
このエンドポイントは、ソート可能インデックスを使用して、プライマリフィールドの値でデータをフィルタリングし、複数のソートフィールドによる多段ソートを適用した結果を返します。例えば、「部署が"営業"の従業員を、役職の昇順、その中で入社日の新しい順にソートする」といった複雑なクエリが可能です。

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "index": "department_hiredate_index",
  "value": "sales",           // プライマリフィールド（department）の値
  "sort": [
    {
      "field": "position",    // 第1ソートフィールド
      "order": "asc"          // 第1ソートの順序: "asc"（昇順）
    },
    {
      "field": "hireDate",    // 第2ソートフィールド
      "order": "desc"         // 第2ソートの順序: "desc"（降順）
    }
  ],
  "pagination": {
    "offset": 0,              // 結果セットの開始位置
    "limit": 10               // 取得する最大件数
  }
}
```

**レスポンス**:
```json
{
  "status": "success",
  "count": 3,                 // 返された結果の件数
  "total": 3,                 // フィルタリング条件に一致する総件数
  "offset": 0,
  "limit": 10,
  "data": [
    {
      "key": "user123",
      "value": {
        "name": "John Doe",
        "department": "sales",
        "hireDate": "2021-08-20",
        "position": "Sales Director"
      }
    },
    {
      "key": "user456",
      "value": {
        "name": "Jane Smith",
        "department": "sales",
        "hireDate": "2023-05-15",
        "position": "Sales Manager"
      }
    },
    {
      "key": "user789",
      "value": {
        "name": "Bob Johnson",
        "department": "sales",
        "hireDate": "2022-11-03",
        "position": "Sales Representative"
      }
    }
  ]
}
```

**注意**:
- 複数のソートフィールドを指定する場合、最初のフィールドで同じ値を持つエントリは、次のフィールドでソートされます。
- 各ソートフィールドに対して個別にソート順序（昇順/降順）を指定できます。
- ソートフィールドが存在しないエントリは、ソート結果の最後に配置されます。

#### インデックス削除

```
POST /index/drop
```

**リクエスト**:
```json
{
  "database": "my_database",
  "set": "users",
  "name": "email_index"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Index dropped successfully"
}
```

### S3連携機能

#### バックアップ実行

```
POST /backup/create
```

**リクエスト**:
```json
{
  "database": "my_database"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Database backed up successfully"
}
```

**注意**: サーバーに管理ユーザーが設定されている場合、このエンドポイントには管理者認証が必要です。

#### バックアップ一覧取得

```
POST /backup/list
```

**リクエスト**:
```json
{
  "database": "my_database"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "backups": [
    {
      "name": "backups/my_database/20250318-123456.json",
      "timestamp": "2025-03-18T12:34:56Z",
      "size": 1024,
      "database": "my_database"
    }
  ]
}
```

**注意**: サーバーに管理ユーザーが設定されている場合、このエンドポイントには管理者認証が必要です。

#### バックアップからの復元

```
POST /backup/restore
```

**リクエスト**:
```json
{
  "backup_name": "backups/my_database/20250318-123456.json"
}
```

**レスポンス**:
```json
{
  "status": "success",
  "message": "Database restored successfully"
}
```

**注意**: サーバーに管理ユーザーが設定されている場合、このエンドポイントには管理者認証が必要です。

### サーバー管理

#### サーバー情報取得

```
POST /server/info
```

**リクエスト**:
```json
{}
```

**レスポンス**:
```json
{
  "status": "success",
  "version": "0.0.1",
  "uptime": "10d 2h 30m",
  "databases_count": 5
}
```

**注意**: サーバーに管理ユーザーが設定されている場合、このエンドポイントには管理者認証が必要です。

## エラーレスポンス

エラーが発生した場合、以下の形式でレスポンスが返されます：

```json
{
  "status": "error",
  "code": "ERROR_CODE",
  "message": "詳細なエラーメッセージ"
}
```

### 一般的なエラーコード

- `DB_NOT_FOUND`: 指定されたデータベースが存在しない
- `SET_NOT_FOUND`: 指定されたSetが存在しない
- `INDEX_NOT_FOUND`: 指定されたインデックスが存在しない
- `KEY_NOT_FOUND`: 指定されたキーが存在しない
- `AUTH_FAILED`: 認証失敗
- `ADMIN_AUTH_REQUIRED`: 管理者認証が必要
- `INVALID_REQUEST`: リクエスト形式が不正
- `INTERNAL_ERROR`: サーバー内部エラー

## 認証

### データベース認証

データベースに設定された認証情報を使用して、そのデータベースのリソースにアクセスするための認証です。

認証が必要なデータベースにアクセスする場合、以下のヘッダーを含める必要があります：

```
Authorization: Basic base64(username:password)
```

または、リクエストボディに認証情報を含めることも可能です：

```json
{
  "database": "my_database",
  "auth": {
    "username": "admin",
    "password": "secure_password"
  },
  // その他のパラメータ
}
```

### 管理者認証

サーバー起動時に`--admin-username`と`--admin-password`オプションで設定された管理ユーザーの認証情報を使用して、管理操作を実行するための認証です。

管理者認証が必要なエンドポイントにアクセスする場合、以下のヘッダーを含める必要があります：

```
X-Admin-Authorization: Basic base64(admin_username:admin_password)
```

または、リクエストボディに管理者認証情報を含めることも可能です：

```json
{
  "admin_auth": {
    "username": "admin_username",
    "password": "admin_password"
  },
  // その他のパラメータ
}
```

管理者認証が必要なエンドポイント：
- `/create` - データベース作成
- `/drop` - データベース削除
- `/backup` - バックアップ実行
- `/restore` - バックアップからの復元
- `/server/info` - サーバー情報取得

**注意**: サーバー起動時に管理ユーザーが設定されていない場合、これらのエンドポイントは認証なしでアクセス可能です。