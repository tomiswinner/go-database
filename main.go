package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Database構造体 - データベースエンジンの中心
type Database struct {
	name    string                 // データベース名
	tables  map[string]*TableDef   // メモリ上のテーブル定義管理
	indexes map[string]*BTree      // テーブルごとのB+Treeインデックス（主キー用）
}

// NewDatabase - 新しいデータベースインスタンスを作成
func NewDatabase(name string) *Database {
	return &Database{
		name:    name,
		tables:  make(map[string]*TableDef),
		indexes: make(map[string]*BTree),
	}
}

// CreateTable - CREATE TABLE文を実行
func (db *Database) CreateTable(sql string) error {
	tableDef, err := ParseCreateTable(sql)
	if err != nil {
		return fmt.Errorf("パースエラー: %v", err)
	}
	
	// スキーマファイルに保存
	err = SaveTableSchema(tableDef)
	if err != nil {
		return fmt.Errorf("スキーマ保存エラー: %v", err)
	}
	
	// メモリにも登録
	db.tables[tableDef.Name] = tableDef
	
	// 主キー（最初のカラム）用のB+Treeインデックスを作成
	if len(tableDef.Columns) > 0 {
		primaryKeyColumn := tableDef.Columns[0].Name
		db.indexes[tableDef.Name] = NewBTree(tableDef.Name, primaryKeyColumn)
		fmt.Printf("主キーインデックス '%s.%s' を作成しました\n", tableDef.Name, primaryKeyColumn)
	}
	
	fmt.Printf("テーブル '%s' を作成しました\n", tableDef.Name)
	fmt.Printf("カラム: ")
	for i, col := range tableDef.Columns {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s(%s)", col.Name, col.Type)
	}
	fmt.Println()
	
	return nil
}

// Insert - INSERT文を実行（usersテーブル専用）
func (db *Database) Insert(sql string) error {
	insertDef, err := ParseInsert(sql)
	if err != nil {
		return fmt.Errorf("パースエラー: %v", err)
	}
	
	// usersテーブル以外は未対応
	if insertDef.TableName != "users" {
		return fmt.Errorf("テーブル '%s' への挿入は未対応です（usersテーブルのみ対応）", insertDef.TableName)
	}
	
	// usersテーブルの場合、User構造体に変換
	if len(insertDef.Values) != 2 {
		return fmt.Errorf("usersテーブルには2つの値（id, name）が必要です")
	}
	
	id, err := strconv.Atoi(insertDef.Values[0])
	if err != nil {
		return fmt.Errorf("IDは数値である必要があります: %v", err)
	}
	
	user := User{
		ID:   id,
		Name: insertDef.Values[1],
	}
	
	// レコード位置を取得（現在のレコード数）
	recordCount, err := CountUsers()
	if err != nil {
		return fmt.Errorf("レコード数取得エラー: %v", err)
	}
	
	// 既存のAppendUser関数を使用
	err = AppendUser(user)
	if err != nil {
		return fmt.Errorf("ユーザー保存エラー: %v", err)
	}
	
	// B+Treeインデックスに主キー（ID）とレコード位置を登録
	if btree, exists := db.indexes[insertDef.TableName]; exists {
		btree.Insert(id, recordCount)
		fmt.Printf("インデックスに登録: key=%d, position=%d\n", id, recordCount)
	}
	
	fmt.Printf("ユーザー '%s' (ID: %d) を追加しました\n", user.Name, user.ID)
	return nil
}

// Select - SELECT文を実行（usersテーブル専用）
func (db *Database) Select(sql string) error {
	selectDef, err := ParseSelect(sql)
	if err != nil {
		return fmt.Errorf("パースエラー: %v", err)
	}
	
	// usersテーブル以外は未対応
	if selectDef.TableName != "users" {
		return fmt.Errorf("テーブル '%s' からの選択は未対応です（usersテーブルのみ対応）", selectDef.TableName)
	}
	
	// WHERE句に基づいてデータを取得
	users, err := db.selectUsersWithWhere(selectDef)
	if err != nil {
		return err
	}
	
	if len(users) == 0 {
		fmt.Println("条件に一致するデータがありません")
		return nil
	}
	
	// 結果を表示
	return db.displayResults(selectDef, users)
}

// selectUsersWithWhere - WHERE句に基づいてユーザーデータを取得
func (db *Database) selectUsersWithWhere(selectDef *SelectDef) ([]User, error) {
	// WHERE句がない場合は全件取得
	if selectDef.WhereClause == nil {
		return ReadAllUsers()
	}
	
	where := selectDef.WhereClause
	
	// 主キー（id）での等価検索の場合、B+Treeインデックスを使用
	if where.Column == "id" && where.Operator == "=" {
		return db.searchByIndex(selectDef.TableName, where.Value)
	}
	
	// その他の条件の場合は全件スキャンでフィルタリング
	return db.searchByFullScan(where)
}

// searchByIndex - B+Treeインデックスを使用した検索
func (db *Database) searchByIndex(tableName, idValue string) ([]User, error) {
	id, err := strconv.Atoi(idValue)
	if err != nil {
		return nil, fmt.Errorf("IDは数値である必要があります: %v", err)
	}
	
	btree, exists := db.indexes[tableName]
	if !exists {
		return nil, fmt.Errorf("インデックスが存在しません")
	}
	
	position, found := btree.Search(id)
	if !found {
		return []User{}, nil // 見つからない場合は空のスライス
	}
	
	fmt.Printf("インデックス検索: key=%d, position=%d\n", id, position)
	
	// 指定位置のレコードを取得
	user, err := db.getUserByPosition(position)
	if err != nil {
		return nil, fmt.Errorf("レコード取得エラー: %v", err)
	}
	
	return []User{user}, nil
}

// searchByFullScan - 全件スキャンによる検索
func (db *Database) searchByFullScan(where *WhereClause) ([]User, error) {
	fmt.Println("全件スキャンで検索中...")
	
	allUsers, err := ReadAllUsers()
	if err != nil {
		return nil, fmt.Errorf("データ読み込みエラー: %v", err)
	}
	
	// 条件にマッチするレコードをフィルタリング
	return db.filterUsers(allUsers, where), nil
}

// displayResults - 検索結果を表示
func (db *Database) displayResults(selectDef *SelectDef, users []User) error {
	// ヘッダーを出力
	if selectDef.IsSelectAll {
		// - : 左よせ
		// 5 : 5文字分のスペースを確保
		// s : string 対象
		fmt.Printf("%-5s | %-20s\n", "ID", "Name")
		fmt.Println("------+----------------------")
	} else {
		// 指定されたカラムのみ（簡易実装）
		for i, col := range selectDef.Columns {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Printf("%-10s", col)
		}
		fmt.Println()
		for range selectDef.Columns {
			fmt.Print("-----------+")
		}
		fmt.Println()
	}
	
	// データを出力
	for _, user := range users {
		if selectDef.IsSelectAll {
			fmt.Printf("%-5d | %-20s\n", user.ID, user.Name)
		} else {
			// 指定されたカラムのみ（簡易実装）
			for i, col := range selectDef.Columns {
				if i > 0 {
					fmt.Print(" | ")
				}
				if col == "id" {
					fmt.Printf("%-10d", user.ID)
				} else if col == "name" {
					fmt.Printf("%-10s", user.Name)
				} else {
					fmt.Printf("%-10s", "?")
				}
			}
			fmt.Println()
		}
	}
	
	return nil
}

// getUserByPosition - 指定位置のユーザーレコードを取得
func (db *Database) getUserByPosition(position int) (User, error) {
	// 指定位置から1件だけ取得
	users, err := ReadUsersWithPaging(position, 1)
	if err != nil {
		return User{}, err
	}
	
	if len(users) == 0 {
		return User{}, fmt.Errorf("指定位置にレコードが存在しません")
	}
	
	return users[0], nil
}

// filterUsers - WHERE条件でユーザーをフィルタリング
func (db *Database) filterUsers(users []User, where *WhereClause) []User {
	var result []User
	
	for _, user := range users {
		match := false
		
		switch where.Column {
		case "id":
			id, err := strconv.Atoi(where.Value)
			if err != nil {
				continue
			}
			
			switch where.Operator {
			case "=":
				match = user.ID == id
			case ">":
				match = user.ID > id
			case "<":
				match = user.ID < id
			case ">=":
				match = user.ID >= id
			case "<=":
				match = user.ID <= id
			}
			
		case "name":
			switch where.Operator {
			case "=":
				match = user.Name == where.Value
			// 文字列の大小比較は一旦省略
			}
		}
		
		if match {
			result = append(result, user)
		}
	}
	
	return result
}

// ExecuteSQL - SQL文を判定して適切なメソッドを呼び出す
func (db *Database) ExecuteSQL(sql string) error {
	sql = strings.TrimSpace(sql)
	
	if strings.HasPrefix(strings.ToUpper(sql), "CREATE TABLE") {
		return db.CreateTable(sql)
	} else if strings.HasPrefix(strings.ToUpper(sql), "INSERT INTO") {
		return db.Insert(sql)
	} else if strings.HasPrefix(strings.ToUpper(sql), "SELECT") {
		return db.Select(sql)
	} else if strings.ToUpper(sql) == "SHOW INDEX" {
		// デバッグ用：インデックスの状況を表示
		return db.ShowIndex()
	} else {
		return fmt.Errorf("サポートされていないSQL文です")
	}
}

// ShowIndex - デバッグ用：B+Treeインデックスの状況を表示
func (db *Database) ShowIndex() error {
	fmt.Println("=== インデックス状況 ===")
	for tableName, btree := range db.indexes {
		fmt.Printf("\nテーブル: %s\n", tableName)
		btree.PrintTree()
	}
	return nil
}

func main() {
	// データベースエンジンを初期化
	db := NewDatabase("go-database")
	
	fmt.Println("Go Database Engine with B+Tree Index - CREATE TABLE、INSERT、SELECT を試してみましょう")
	fmt.Println("例:")
	fmt.Println("  CREATE TABLE users (id INT, name TEXT);")
	fmt.Println("  INSERT INTO users (id, name) VALUES (1, 'Alice');")
	fmt.Println("  SELECT * FROM users;")
	fmt.Println("  SELECT * FROM users WHERE id = 1; (インデックス検索)")
	fmt.Println("  SELECT * FROM users WHERE id > 1; (全件スキャン)")
	fmt.Println("  SHOW INDEX; (インデックス状況表示)")
	fmt.Print("SQL> ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		sql := scanner.Text()
		
		// データベースエンジンでSQL文を実行
		err := db.ExecuteSQL(sql)
		if err != nil {
			fmt.Println("エラー:", err)
		}
	}
} 
