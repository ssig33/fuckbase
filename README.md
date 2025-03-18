# FuckBase

FuckBaseは、Goで実装されたシンプルで高速なHTTPベースのデータベースです。キーバリューストアの機能に加え、強力なインデックス機能とS3連携機能を備えています。

## 特徴

- **シンプルなHTTP API**: すべての操作がHTTP POSTリクエストで行えます
- **MessagePackエンコーディング**: 効率的なデータ保存と取得
- **強力なインデックス機能**: 複数のインデックスをサポートし、ソートとページネーションが可能
- **S3連携**: バックアップと復元機能（開発中）
- **軽量**: 最小限のリソース消費で高速に動作

## 実装状況

現在の実装では、以下の機能が利用可能です：

- データベースの作成と削除
- Setの作成と一覧取得
- データの保存、取得、削除
- インデックスの作成、削除、クエリ
- 基本的な認証機能

以下の機能は現在開発中です：

- S3連携機能（バックアップと復元）

## 使用例

FuckBaseは以下のようなユースケースに適しています：

- 高速なキーバリューストア
- 単純なドキュメントデータベース
- キャッシュシステム
- 一時的なデータストレージ
- マイクロサービスのバックエンドストレージ

## クイックスタート

### インストール

```bash
go get github.com/ssig33/fuckbase
```

または、リリースページからバイナリをダウンロードしてください。

### サーバーの起動

基本的な起動:
```bash
fuckbase --port 8080
```

管理ユーザーを設定して起動:
```bash
fuckbase --port 8080 --admin-username admin --admin-password secure_password
```

S3バックアップを有効にして起動:
```bash
fuckbase --port 8080 --s3-endpoint https://s3.amazonaws.com --s3-bucket my-backup-bucket --s3-access-key ACCESS_KEY --s3-secret-key SECRET_KEY
```

### Dockerを使用する

Docker Composeを使用して簡単に起動できます:

```bash
docker-compose up -d
```

これにより、FuckBaseサーバーがポート8080で起動します。

### E2Eテスト

E2Eテストを実行するには:

```bash
./run-e2e-tests.sh
```

このスクリプトは、Docker Composeを使用してFuckBaseサーバーとテストクライアントを起動し、基本的な機能をテストします。テストは以下の機能を検証します:

- データベースの作成と削除
- Setの作成
- データの保存と取得
- インデックスの作成、クエリ、削除
- データの削除

テスト結果は標準出力に表示されます。

## Ruby クライアント

FuckBaseには公式のRubyクライアントライブラリが用意されています。詳細は[Rubyクライアントのドキュメント](ruby/ruby-client-README.md)を参照してください。

### 基本的な使用例

```ruby
# クライアントの作成
client = FuckBase::Client.new(host: 'localhost', port: 8080)

# データベースの作成
db = client.create_database('my_database')

# セットの作成
users_set = db.create_set('users')

# データの保存
users_set.put('user1', {
  'name' => 'John Doe',
  'email' => 'john@example.com'
})

# データの取得
user = users_set.get('user1')
puts "User: #{user['name']}, Email: #{user['email']}"
```

### データベースの作成

```bash
curl -X POST http://localhost:8080/create -d '{"name": "mydb"}'
```

管理ユーザーが設定されている場合:
```bash
curl -X POST http://localhost:8080/create -H "X-Admin-Authorization: Basic $(echo -n 'admin:secure_password' | base64)" -d '{"name": "mydb"}'
```

### データの保存

```bash
curl -X POST http://localhost:8080/set/put -d '{
  "database": "mydb",
  "set": "users",
  "key": "user1",
  "value": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}'
```

### データの取得

```bash
curl -X POST http://localhost:8080/set/get -d '{
  "database": "mydb",
  "set": "users",
  "key": "user1"
}'
```

### インデックスの作成

```bash
curl -X POST http://localhost:8080/index/create -d '{
  "database": "mydb",
  "set": "users",
  "name": "email_index",
  "field": "email"
}'
```

### インデックスを使用したクエリ

```bash
curl -X POST http://localhost:8080/index/query -d '{
  "database": "mydb",
  "set": "users",
  "index": "email_index",
  "value": "john@example.com",
  "sort": "asc",
  "limit": 10,
  "offset": 0
}'
```

### データの削除

```bash
curl -X POST http://localhost:8080/set/delete -d '{
  "database": "mydb",
  "set": "users",
  "key": "user1"
}'
```

## ドキュメント

詳細なドキュメントは以下のファイルを参照してください：

- [仕様書](docs/SPEC.md) - プロジェクトの詳細な仕様
- [API仕様書](docs/API_SPEC.md) - APIエンドポイントの詳細
- [アーキテクチャ](docs/ARCHITECTURE.md) - 内部設計と実装の詳細
- [ロードマップ](docs/ROADMAP.md) - 開発計画と今後の機能
- [S3バックアップ](docs/s3-backup.md) - S3バックアップと復元機能の詳細

## 開発方針

FuckBaseの開発は以下の原則に基づいています：

- **テスト駆動開発**: すべての機能に対して必ずテストを書く
- **迅速な実装**: 機能を素早く実装し、継続的に改善する
- **継続的テスト**: 実装後、頻繁にテストを実行して品質を確保する
- **簡潔なドキュメント**: 必要最小限のドキュメントを維持し、不要になったドキュメントやコードは積極的に削除する

## CI/CD

このプロジェクトはGitHub Actionsを使用して継続的インテグレーションを実施しています：

- **ユニットテスト**: すべてのパッケージのユニットテストを自動実行
- **E2Eテスト**: Docker Composeを使用して実際のサーバーを起動し、APIエンドポイントをテスト
- **ビルド検証**: 異なるプラットフォームでのビルド検証

PRを作成する際は、すべてのテストが通過することを確認してください。

## 貢献

貢献は大歓迎です！以下の手順で貢献できます：

1. リポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## ライセンス

このプロジェクトはWTFPLライセンス（Do What The Fuck You Want To Public License）の下で公開されています。詳細は[LICENSE](LICENSE)ファイルを参照してください。