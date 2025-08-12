package main

import (
	"bufio"
	"fmt"
	"os"
)

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
