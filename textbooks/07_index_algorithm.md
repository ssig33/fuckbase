# 第7章: インデックス検索アルゴリズム

この章では、FuckBaseのインデックス検索アルゴリズムについて詳しく説明します。インデックスは、データベースのパフォーマンスを向上させるための重要な機能であり、FuckBaseでは効率的なインデックス検索を実現するためのアルゴリズムが実装されています。

## インデックスの基本概念

インデックスは、データの検索を高速化するためのデータ構造です。FuckBaseでは、特定のフィールドの値に基づいてデータを効率的に検索するための二次インデックスを提供しています。

```mermaid
graph TD
A("データ") --> B("インデックスなし - 全件スキャン")
A --> C("インデックスあり - インデックス検索")
B --> D("O(n)の時間複雑度")
C --> E("O(1)の時間複雑度")
E --> F("キーのリスト取得")
F --> G("データ取得")
D --> H("結果")
G --> H
```

インデックスがない場合、データを検索するには全件スキャンが必要で、時間複雑度はO(n)です。一方、インデックスがある場合、検索対象のフィールド値からキーのリストを直接取得できるため、時間複雑度はO(1)になります。

## FuckBaseのインデックス構造

FuckBaseのインデックスは、フィールド値からキーのリストへのマッピングを保持する構造になっています。

```mermaid
classDiagram
    class Index {
        +Name string
        +SetName string
        +Field string
        +Values map[string][]string
        +mu sync.RWMutex
    }
```

`Values`マップは、フィールド値をキーとし、その値を持つすべてのデータのキーのリストを値として持ちます。例えば、`email`フィールドにインデックスを作成した場合、`Values`マップは以下のようになります：

```
Values = {
    "john@example.com": ["user1", "user5"],
    "alice@example.com": ["user2"],
    "bob@example.com": ["user3", "user7", "user9"]
}
```

この構造により、特定のメールアドレスを持つすべてのユーザーを効率的に検索することができます。

インデックスの実装の詳細は [../internal/database/index.go](../internal/database/index.go) で確認できます。

## インデックスの構築アルゴリズム

インデックスの構築は、Setのすべてのエントリをスキャンして行われます。以下は、インデックス構築のアルゴリズムです：

```mermaid
sequenceDiagram
    participant Index as "インデックス"
    participant Set as Set
    Index->>Index: Values = {}（空のマップを初期化）
    Index->>Set: ForEach(callback)
    loop "各エントリ(key, value)"
        Set-->>Index: callback(key, value)
        Index->>Index: fieldValue = extractFieldValue(value)
        alt "フィールドが存在する"
            Index->>Index: Values[fieldValue] = append(Values[fieldValue], key)
        else "フィールドが存在しない"
            Index->>Index: スキップ
        end
    end
```

このアルゴリズムの実装は [../internal/database/index.go](../internal/database/index.go) の`Build`メソッドで確認できます。

### フィールド値の抽出

インデックス構築の重要な部分は、MessagePackエンコードされたデータからフィールド値を抽出することです。このプロセスは以下の手順で行われます：

1. MessagePackデータをデコードしてマップに変換
2. マップから指定されたフィールドの値を取得
3. 値を文字列に変換（数値、ブール値などの場合）

```mermaid
graph TD
    A("MessagePackデータ") -->|"デコード"| B("マップ")
    B -->|"フィールド検索"| C{"フィールドが存在?"}
    C -->|"はい"| D("値の型チェック")
    C -->|"いいえ"| E("エラー")
    D -->|"文字列"| F("そのまま使用")
    D -->|"数値"| G("文字列に変換")
    D -->|"ブール値"| H("文字列に変換")
    D -->|"その他"| I("エラー")
    F -->|"結果"| J("フィールド値")
    G -->|"結果"| J
    H -->|"結果"| J
```

このアルゴリズムの実装は [../internal/database/index.go](../internal/database/index.go) の`extractFieldValue`メソッドで確認できます。

## インデックス検索アルゴリズム

インデックス検索は、指定されたフィールド値に一致するキーのリストを返します。以下は、インデックス検索のアルゴリズムです：

```mermaid
graph TD
    A("検索開始") -->|"値指定"| B{"値が存在?"}
    B -->|"はい"| C("キーのリストをコピー")
    B -->|"いいえ"| D("空のリストを返す")
    C -->|"結果"| J("キーのリスト")
```

このアルゴリズムの実装は [../internal/database/index.go](../internal/database/index.go) の`Query`メソッドで確認できます。


## インデックスの更新アルゴリズム

FuckBaseでは、データの変更（追加、更新、削除）に応じてインデックスを更新する必要があります。以下は、インデックス更新のアルゴリズムです：

### エントリの追加

```mermaid
graph TD
    A("エントリ追加") -->|"フィールド値抽出"| B{"フィールドが存在?"}
    B -->|"はい"| C("Values[fieldValue]にキーを追加")
    B -->|"いいえ"| D("何もしない")
```

### エントリの削除

```mermaid
graph TD
    A("エントリ削除") -->|"フィールド値抽出"| B{"フィールドが存在?"}
    B -->|"はい"| C("Values[fieldValue]からキーを削除")
    B -->|"いいえ"| D("何もしない")
    C -->|"リストが空?"| E{"空?"}
    E -->|"はい"| F("Values[fieldValue]を削除")
    E -->|"いいえ"| G("更新されたリストを保持")
```

### エントリの更新

```mermaid
graph TD
    A("エントリ更新") -->|"古い値からフィールド値抽出"| B("古いフィールド値")
    A -->|"新しい値からフィールド値抽出"| C("新しいフィールド値")
    B -->|"削除処理"| D("Values[oldFieldValue]からキーを削除")
    C -->|"追加処理"| E("Values[newFieldValue]にキーを追加")
```

これらのアルゴリズムの実装は [../internal/database/index.go](../internal/database/index.go) の`AddEntry`、`RemoveEntry`、`UpdateEntry`メソッドで確認できます。

## インデックスの性能特性

FuckBaseのインデックスは、以下の性能特性を持っています：

### 時間複雑度

- **構築**: O(n)（nはSetのエントリ数）
- **検索**: O(1)（ハッシュマップルックアップ）
- **追加**: O(1)
- **削除**: O(k)（kはフィールド値に対応するキーの数）
- **更新**: O(k)（kは古いフィールド値に対応するキーの数）

### 空間複雑度

インデックスの空間複雑度はO(n)です（nはSetのエントリ数）。各エントリのキーがインデックスに保存されるため、エントリ数に比例してメモリ使用量が増加します。

## インデックスの種類

FuckBaseでは、以下の2種類のインデックスをサポートしています：

### 1. 基本インデックス（Basic Index）

基本インデックスは、単一フィールドの値に基づいてデータを検索するためのシンプルなインデックスです。

```mermaid
classDiagram
    class Index {
        +Name string
        +SetName string
        +Field string
        +Values map[string][]string
        +mu sync.RWMutex
    }
```

### 2. ソート可能インデックス（Sortable Index）

ソート可能インデックスは、プライマリフィールドでフィルタリングした結果を、セカンダリフィールドでソートするための拡張インデックスです。

```mermaid
classDiagram
    class SortableIndex {
        +Name string
        +SetName string
        +PrimaryField string
        +SortFields []string
        +Values map[string][]string
        +SortValues map[string]map[string]interface{}
        +mu sync.RWMutex
    }
```

`SortValues`マップは、キーからソートフィールドの値へのマッピングを保持します。例えば、`department`フィールドでフィルタリングし、`hireDate`フィールドでソートする場合、`SortValues`マップは以下のようになります：

```
SortValues = {
    "user1": {"hireDate": "2023-01-15"},
    "user2": {"hireDate": "2022-05-20"},
    "user3": {"hireDate": "2023-03-10"}
}
```

## ソート可能インデックスのアルゴリズム

### インデックスの構築

ソート可能インデックスの構築は、基本インデックスの構築と同様に行われますが、ソートフィールドの値も抽出して保存します。

```mermaid
sequenceDiagram
    participant Index as "ソート可能インデックス"
    participant Set as Set
    Index->>Index: Values = {}（空のマップを初期化）
    Index->>Index: SortValues = {}（空のマップを初期化）
    Index->>Set: ForEach(callback)
    loop "各エントリ(key, value)"
        Set-->>Index: callback(key, value)
        Index->>Index: primaryValue = extractFieldValue(value, PrimaryField)
        alt "プライマリフィールドが存在する"
            Index->>Index: Values[primaryValue] = append(Values[primaryValue], key)
            Index->>Index: sortValues = {}（空のマップを初期化）
            loop "各ソートフィールド"
                Index->>Index: sortValue = extractFieldValue(value, SortField)
                alt "ソートフィールドが存在する"
                    Index->>Index: sortValues[SortField] = sortValue
                end
            end
            Index->>Index: SortValues[key] = sortValues
        else "プライマリフィールドが存在しない"
            Index->>Index: スキップ
        end
    end
```

### ソート検索とページング

ソート可能インデックスを使用した検索は、以下の手順で行われます：

```mermaid
graph TD
    A("検索開始") -->|"プライマリ値指定"| B{"値が存在?"}
    B -->|"はい"| C("キーのリストを取得")
    B -->|"いいえ"| D("空のリストを返す")
    C -->|"ソートフィールド指定"| E("キーをソート")
    E -->|"ソート順序指定"| F("昇順/降順でソート")
    F -->|"ページング指定"| H("オフセット/リミット適用")
    H -->|"結果"| G("ソートされたキーのリスト")
    D -->|"結果"| G
```

ソートアルゴリズムは、指定されたソートフィールドの値に基づいてキーをソートします。複数のソートフィールドが指定された場合は、最初のフィールドで同値の場合に次のフィールドでソートするという多段ソートを行います。

#### ページング機能

ページング機能は、ソートされた結果に対してオフセットとリミットを適用することで実現されます：

1. **オフセット**: 結果セットの先頭からスキップするエントリの数
2. **リミット**: 返却するエントリの最大数

例えば、オフセット10、リミット20を指定すると、ソートされた結果の11番目から30番目までのエントリが返されます。これにより、大量のデータを効率的にページ単位で取得することができます。

ページングの時間複雑度はO(1)です。ソート済みのキーリストに対して単純なスライス操作を行うだけなので、データ量に関係なく一定の時間で処理できます。

## インデックスの制限事項

FuckBaseのインデックスには、以下の制限事項があります：

1. **完全一致のみ**: インデックス検索は完全一致のみをサポートしています。部分一致や範囲検索はサポートされていません。

2. **文字列変換**: 数値やブール値などの非文字列フィールドは、検索のために文字列に変換されます。これにより、数値の範囲検索などの高度な検索機能が制限されます。

3. **インデックス更新のオーバーヘッド**: データの変更時にインデックスを更新する必要があるため、書き込み操作のパフォーマンスに影響を与える可能性があります。

4. **ソートフィールドの欠落**: ソート可能インデックスでは、ソートフィールドが存在しないエントリはソート結果の最後に配置されます。

## インデックス使用の最適化

FuckBaseでインデックスを効果的に使用するためのいくつかのベストプラクティスを紹介します：

1. **頻繁に検索されるフィールドにインデックスを作成**: 検索頻度の高いフィールドにインデックスを作成することで、検索パフォーマンスを向上させることができます。

2. **不要なインデックスを避ける**: 使用されないインデックスはメモリを消費し、書き込みパフォーマンスに影響を与えるため、必要なインデックスのみを作成することをお勧めします。

3. **カーディナリティの高いフィールドを選択**: ユニークな値が多いフィールド（例：ID、メールアドレス）にインデックスを作成すると、検索の効率が向上します。


## まとめ

FuckBaseのインデックス検索アルゴリズムは、シンプルながらも効率的に設計されています。フィールド値からキーのリストへのマッピングを使用することで、O(1)の時間複雑度で検索を行うことができます。

インデックスの構築、更新、検索のアルゴリズムを理解することで、FuckBaseを効果的に使用し、アプリケーションのパフォーマンスを最適化することができます。

インデックスの実装の詳細については、[../internal/database/index.go](../internal/database/index.go)、[../internal/database/index_test.go](../internal/database/index_test.go)、[../internal/database/index_integration_test.go](../internal/database/index_integration_test.go)、[../internal/database/index_missing_field_test.go](../internal/database/index_missing_field_test.go) を参照してください。