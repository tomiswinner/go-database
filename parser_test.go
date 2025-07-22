package main

import (
	"testing"
)

func TestParseCreateTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *TableDef
		hasError bool
	}{
		{
			name: "基本的なCREATE TABLE",
			sql:  "CREATE TABLE users (id INT, name TEXT);",
			expected: &TableDef{
				Name: "users",
				Columns: []ColumnDef{
					{Name: "id", Type: "INT"},
					{Name: "name", Type: "TEXT"},
				},
			},
			hasError: false,
		},
		{
			name: "セミコロンなし",
			sql:  "CREATE TABLE products (price DECIMAL, category TEXT)",
			expected: &TableDef{
				Name: "products",
				Columns: []ColumnDef{
					{Name: "price", Type: "DECIMAL"},
					{Name: "category", Type: "TEXT"},
				},
			},
			hasError: false,
		},
		{
			name: "大文字小文字混在",
			sql:  "create table Orders (order_id INT, user_id INT);",
			expected: &TableDef{
				Name: "Orders",
				Columns: []ColumnDef{
					{Name: "order_id", Type: "INT"},
					{Name: "user_id", Type: "INT"},
				},
			},
			hasError: false,
		},
		{
			name:     "無効な構文",
			sql:      "CREATE users (id INT);",
			expected: nil,
			hasError: true,
		},
		{
			name:     "カラムなし",
			sql:      "CREATE TABLE empty ();",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCreateTable(tt.sql)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
				}
				return
			}
			
			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}
			
			if result.Name != tt.expected.Name {
				t.Errorf("テーブル名が一致しません。期待: %s, 実際: %s", tt.expected.Name, result.Name)
			}
			
			if len(result.Columns) != len(tt.expected.Columns) {
				t.Errorf("カラム数が一致しません。期待: %d, 実際: %d", len(tt.expected.Columns), len(result.Columns))
				return
			}
			
			for i, col := range result.Columns {
				expected := tt.expected.Columns[i]
				if col.Name != expected.Name || col.Type != expected.Type {
					t.Errorf("カラム[%d]が一致しません。期待: %+v, 実際: %+v", i, expected, col)
				}
			}
		})
	}
}

func TestParseInsert(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *InsertDef
		hasError bool
	}{
		{
			name: "基本的なINSERT",
			sql:  "INSERT INTO users (id, name) VALUES (1, 'Alice');",
			expected: &InsertDef{
				TableName: "users",
				Columns:   []string{"id", "name"},
				Values:    []string{"1", "Alice"},
			},
			hasError: false,
		},
		{
			name: "セミコロンなし",
			sql:  "INSERT INTO products (price, category) VALUES (100, 'book')",
			expected: &InsertDef{
				TableName: "products",
				Columns:   []string{"price", "category"},
				Values:    []string{"100", "book"},
			},
			hasError: false,
		},
		{
			name: "大文字小文字混在",
			sql:  "insert into Orders (order_id, user_id) values (1, 10);",
			expected: &InsertDef{
				TableName: "Orders",
				Columns:   []string{"order_id", "user_id"},
				Values:    []string{"1", "10"},
			},
			hasError: false,
		},
		{
			name:     "カラム数と値の数が不一致",
			sql:      "INSERT INTO users (id, name) VALUES (1);",
			expected: nil,
			hasError: true,
		},
		{
			name:     "無効な構文",
			sql:      "INSERT users VALUES (1, 'Alice');",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInsert(tt.sql)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
				}
				return
			}
			
			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}
			
			if result.TableName != tt.expected.TableName {
				t.Errorf("テーブル名が一致しません。期待: %s, 実際: %s", tt.expected.TableName, result.TableName)
			}
			
			if len(result.Columns) != len(tt.expected.Columns) {
				t.Errorf("カラム数が一致しません。期待: %d, 実際: %d", len(tt.expected.Columns), len(result.Columns))
				return
			}
			
			for i, col := range result.Columns {
				if col != tt.expected.Columns[i] {
					t.Errorf("カラム[%d]が一致しません。期待: %s, 実際: %s", i, tt.expected.Columns[i], col)
				}
			}
			
			if len(result.Values) != len(tt.expected.Values) {
				t.Errorf("値の数が一致しません。期待: %d, 実際: %d", len(tt.expected.Values), len(result.Values))
				return
			}
			
			for i, val := range result.Values {
				if val != tt.expected.Values[i] {
					t.Errorf("値[%d]が一致しません。期待: %s, 実際: %s", i, tt.expected.Values[i], val)
				}
			}
		})
	}
} 
