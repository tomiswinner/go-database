// storage.go: ユーザーデータの保存・読み込みを担当

package main

import (
	"encoding/csv" // CSVファイル操作用
	// エラーメッセージ出力用
	"os"      // ファイル操作用
	"reflect" // 構造体フィールド数取得用
	"strconv" // 文字列⇔数値変換用
)

// Userテーブルの1行分のデータ構造
// ID: ユーザーID
// Name: ユーザー名
type User struct {
    ID   int    // ユーザーID
    Name string // ユーザー名
}

// User構造体のフィールド数を取得
func getUserFieldCount() int {
    return reflect.TypeOf(User{}).NumField()
}

// users.dbファイルを新規作成する関数
// 既にファイルが存在する場合は中身を空にする
func CreateUsersTableFile() error {
    // os.Createはファイルを新規作成（既存なら上書き）
    f, err := os.Create("users.db")
    if err != nil {
        return err
    }
    defer f.Close()
    return nil
}

// UserをCSV形式でusers.dbに追記保存する関数
// 例: 1,alice\n 2,bob\n のように1行1ユーザーで保存
func AppendUser(user User) error {
    // os.OpenFileでファイルを開く
    // os.O_APPEND: 追記モード
    // os.O_CREATE: なければ新規作成
    // os.O_WRONLY: 書き込み専用
    f, err := os.OpenFile("users.db", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    writer := csv.NewWriter(f) // CSV書き込み用
    defer writer.Flush()

    record := []string{
        strconv.Itoa(user.ID), // int→string変換
        user.Name,
    }
    return writer.Write(record)
}

// users.dbから全件を読み込み、Userスライスとして返す関数
// ファイルが空でも空スライスを返す
func ReadAllUsers() ([]User, error) {
    f, err := os.Open("users.db") // 読み込み専用で開く
    if err != nil {
        return nil, err
    }
    defer f.Close()

    reader := csv.NewReader(f) // CSV読み込み用
    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    var users []User
    fieldCount := getUserFieldCount()
    
    for _, rec := range records {
        if len(rec) < fieldCount {
            continue // 不正な行はスキップ
        }
        id, err := strconv.Atoi(rec[0]) // string→int変換
        if err != nil {
            continue // 変換失敗はスキップ
        }
        users = append(users, User{
            ID:   id,
            Name: rec[1],
        })
    }
    return users, nil
}

// 実際のRDBでは、データ読み出しはページ単位（I/O最適化、キャッシュ管理もしやすい、WALもページ単位）
// N件ずつusers.dbから読み込む関数
// offset: 読み込み開始位置（0から）
// limit: 読み込む最大件数
func ReadUsersWithPaging(offset, limit int) ([]User, error) {
    f, err := os.Open("users.db")
    if err != nil {
        return nil, err
    }
    defer f.Close()

    reader := csv.NewReader(f)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    var users []User
    count := 0
    fieldCount := getUserFieldCount()
    
    for i, rec := range records {
        // オフセット分をスキップ
        if i < offset {
            continue
        }
        
        // 指定件数に達したら終了
        if count >= limit {
            break
        }
        
        if len(rec) < fieldCount {
            //  user id, name だけのハード実装
            // 不正な行はスキップ（カウントしない）
            continue 
        }
        
        id, err := strconv.Atoi(rec[0])
        if err != nil {
            continue // 変換失敗はスキップ（カウントしない）
        }
        
        users = append(users, User{
            ID:   id,
            Name: rec[1],
        })
        count++
    }
    
    return users, nil
}

// users.dbの総レコード数を取得する関数
func CountUsers() (int, error) {
    f, err := os.Open("users.db")
    if err != nil {
        return 0, err
    }
    defer f.Close()

    reader := csv.NewReader(f)
    records, err := reader.ReadAll()
    if err != nil {
        return 0, err
    }

    count := 0
    fieldCount := getUserFieldCount()
    
    for _, rec := range records {
        if len(rec) >= fieldCount {
            if _, err := strconv.Atoi(rec[0]); err == nil {
                count++
            }
        }
    }
    
    return count, nil
} 
