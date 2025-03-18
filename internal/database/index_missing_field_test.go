package database

import (
	"testing"
)

// TestIndexQueryWithMissingField は、インデックスフィールドを持たないデータの扱いを
// 明確に示すためのテストです
func TestIndexQueryWithMissingField(t *testing.T) {
	// データベースを作成
	db := NewDatabase("test_db", nil)

	// セットを作成
	set, err := db.CreateSet("users")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Nameフィールドを持つデータ型
	type UserWithName struct {
		Name  string
		Age   int
		Email string
	}

	// Nameフィールドを持たないデータ型
	type UserWithoutName struct {
		Age   int
		Email string
	}

	// Nameフィールドを持つデータを追加
	set.Put("user1", UserWithName{Name: "Alice", Age: 30, Email: "alice@example.com"})
	set.Put("user2", UserWithName{Name: "Bob", Age: 25, Email: "bob@example.com"})

	// Nameフィールドを持たないデータを追加
	set.Put("user3", UserWithoutName{Age: 40, Email: "noname1@example.com"})
	set.Put("user4", UserWithoutName{Age: 50, Email: "noname2@example.com"})

	// Nameフィールドにインデックスを作成
	index, err := db.CreateIndex("name_index", "users", "Name")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// テスト1: インデックスフィールドを持たないデータはインデックスに追加されないことを確認
	t.Run("Missing fields are not indexed", func(t *testing.T) {
		// インデックスに含まれるすべての値を取得
		allValues := index.GetAllValues()
		
		// すべての値に対してクエリを実行し、user3とuser4が含まれていないことを確認
		for _, value := range allValues {
			keys, err := index.Query(value, "", 0, 0)
			if err != nil {
				t.Fatalf("Failed to query index: %v", err)
			}
			
			for _, key := range keys {
				if key == "user3" || key == "user4" {
					t.Errorf("Found user3 or user4 in index under value '%s', but they should not be indexed", value)
				}
			}
		}
	})

	// テスト2: 通常のフィールド値でインデックス検索
	// 期待結果: そのフィールド値を持つデータのキーが返される
	t.Run("Query for existing field value", func(t *testing.T) {
		// "Alice"で検索
		keys, err := index.Query("Alice", "", 0, 0)
		if err != nil {
			t.Fatalf("Failed to query index: %v", err)
		}

		// 結果の数を確認
		if len(keys) != 1 {
			t.Errorf("Expected 1 key for Alice, got %d: %v", len(keys), keys)
		}

		// user1が含まれているか確認
		if len(keys) > 0 && keys[0] != "user1" {
			t.Errorf("Expected to find user1 for Alice, got %s", keys[0])
		}

		// "Bob"で検索
		keys, err = index.Query("Bob", "", 0, 0)
		if err != nil {
			t.Fatalf("Failed to query index: %v", err)
		}

		// 結果の数を確認
		if len(keys) != 1 {
			t.Errorf("Expected 1 key for Bob, got %d: %v", len(keys), keys)
		}

		// user2が含まれているか確認
		if len(keys) > 0 && keys[0] != "user2" {
			t.Errorf("Expected to find user2 for Bob, got %s", keys[0])
		}
	})

	// テスト3: 存在しないフィールド値でインデックス検索
	// 期待結果: 空の結果が返される
	t.Run("Query for non-existent field value", func(t *testing.T) {
		keys, err := index.Query("NonExistent", "", 0, 0)
		if err != nil {
			t.Fatalf("Failed to query index: %v", err)
		}

		// 結果の数を確認
		if len(keys) != 0 {
			t.Errorf("Expected 0 keys for NonExistent, got %d: %v", len(keys), keys)
		}
	})
}