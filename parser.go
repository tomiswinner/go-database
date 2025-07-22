package main

import (
	"fmt"
	"regexp"
	"strings"
)

// ColumnDefはカラム名と型を表す
type ColumnDef struct {
    Name string // カラム名
    Type string // 型（例: INT, TEXT）
}

// TableDefはテーブル名とカラム定義のリストを表す
type TableDef struct {
    Name    string      // テーブル名
    Columns []ColumnDef // カラム定義
}

// InsertDefはINSERT文の内容を表す
type InsertDef struct {
    TableName string   // テーブル名
    Columns   []string // カラム名のリスト
    Values    []string // 値のリスト
}

// SelectDefはSELECT文の内容を表す
type SelectDef struct {
    TableName   string       // テーブル名
    Columns     []string     // 選択するカラム名（*の場合は全カラム）
    IsSelectAll bool         // SELECT * かどうか
    WhereClause *WhereClause // WHERE句（nilの場合は条件なし）
}

// WhereClauseはWHERE句の条件を表す
type WhereClause struct {
    Column   string // カラム名
    Operator string // 演算子（=, >, <, >=, <=）
    Value    string // 比較値
}

// ParseCreateTableはCREATE TABLE文をパースし、TableDefを返す
func ParseCreateTable(sql string) (*TableDef, error) {
    // 例: CREATE TABLE users (id INT, name TEXT);
    re := regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(\w+)\s*\((.+)\)\s*;?$`)
    matches := re.FindStringSubmatch(strings.TrimSpace(sql))
    if len(matches) != 3 {
        return nil, fmt.Errorf("invalid CREATE TABLE syntax")
    }
    tableName := matches[1]
    columnsStr := matches[2]
    columns := []ColumnDef{}
    for _, col := range strings.Split(columnsStr, ",") {
        parts := strings.Fields(strings.TrimSpace(col))
        if len(parts) != 2 {
            continue
        }
        columns = append(columns, ColumnDef{
            Name: parts[0],
            Type: parts[1],
        })
    }
    if len(columns) == 0 {
        return nil, fmt.Errorf("no columns defined")
    }
    return &TableDef{
        Name:    tableName,
        Columns: columns,
    }, nil
}

// ParseInsertはINSERT文をパースし、InsertDefを返す
func ParseInsert(sql string) (*InsertDef, error) {
    // 例: INSERT INTO users (id, name) VALUES (1, 'Alice');
    re := regexp.MustCompile(`(?i)^INSERT\s+INTO\s+(\w+)\s*\(([^)]+)\)\s+VALUES\s*\(([^)]+)\)\s*;?$`)
    matches := re.FindStringSubmatch(strings.TrimSpace(sql))
    if len(matches) != 4 {
        return nil, fmt.Errorf("invalid INSERT syntax")
    }
    
    tableName := matches[1]
    columnsStr := matches[2]
    valuesStr := matches[3]
    
    // カラム名をパース
    columns := []string{}
    for _, col := range strings.Split(columnsStr, ",") {
        columns = append(columns, strings.TrimSpace(col))
    }
    
    // 値をパース（シンプルに文字列分割、クォート処理は後で改善）
    values := []string{}
    for _, val := range strings.Split(valuesStr, ",") {
        val = strings.TrimSpace(val)
        // シングルクォートを除去
        if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
            val = val[1 : len(val)-1]
        }
        values = append(values, val)
    }
    
    if len(columns) != len(values) {
        return nil, fmt.Errorf("column count does not match value count")
    }
    
    return &InsertDef{
        TableName: tableName,
        Columns:   columns,
        Values:    values,
    }, nil
}

// ParseSelectはSELECT文をパースし、SelectDefを返す
func ParseSelect(sql string) (*SelectDef, error) {
    // 例: SELECT * FROM users;
    // 例: SELECT id, name FROM users;
    // 例: SELECT * FROM users WHERE id = 1;
    
    // WHERE句があるかチェック
    whereIndex := strings.Index(strings.ToUpper(sql), "WHERE")
    var baseSQL, whereSQL string
    
    if whereIndex != -1 {
        baseSQL = strings.TrimSpace(sql[:whereIndex])
        whereSQL = strings.TrimSpace(sql[whereIndex+5:]) // "WHERE"の後
    } else {
        baseSQL = sql
        whereSQL = ""
    }
    
    // 基本のSELECT文をパース
    re := regexp.MustCompile(`(?i)^SELECT\s+(.+?)\s+FROM\s+(\w+)\s*;?$`)
    matches := re.FindStringSubmatch(strings.TrimSpace(baseSQL))
    if len(matches) != 3 {
        return nil, fmt.Errorf("invalid SELECT syntax")
    }
    
    columnsStr := strings.TrimSpace(matches[1])
    tableName := matches[2]
    
    var columns []string
    var isSelectAll bool
    
    if columnsStr == "*" {
        isSelectAll = true
        columns = []string{} // 空のまま（後でスキーマから取得）
    } else {
        isSelectAll = false
        for _, col := range strings.Split(columnsStr, ",") {
            columns = append(columns, strings.TrimSpace(col))
        }
    }
    
    // WHERE句をパース
    var whereClause *WhereClause
    if whereSQL != "" {
        var err error
        whereClause, err = parseWhere(whereSQL)
        if err != nil {
            return nil, fmt.Errorf("WHERE句パースエラー: %v", err)
        }
    }
    
    return &SelectDef{
        TableName:   tableName,
        Columns:     columns,
        IsSelectAll: isSelectAll,
        WhereClause: whereClause,
    }, nil
}

// parseWhere - WHERE句をパースする内部関数
func parseWhere(whereSQL string) (*WhereClause, error) {
    // セミコロンを除去
    whereSQL = strings.TrimSuffix(strings.TrimSpace(whereSQL), ";")
    
    // 演算子のパターンを定義（長い順にマッチング）
    operators := []string{">=", "<=", "=", ">", "<"}
    
    for _, op := range operators {
        if strings.Contains(whereSQL, op) {
            parts := strings.SplitN(whereSQL, op, 2)
            if len(parts) == 2 {
                column := strings.TrimSpace(parts[0])
                value := strings.TrimSpace(parts[1])
                
                // シングルクォートを除去
                if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
                    value = value[1 : len(value)-1]
                }
                
                return &WhereClause{
                    Column:   column,
                    Operator: op,
                    Value:    value,
                }, nil
            }
        }
    }
    
    return nil, fmt.Errorf("サポートされていないWHERE句形式: %s", whereSQL)
} 
