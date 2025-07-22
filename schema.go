package main

import (
	"encoding/json"
	"os"
)

// テーブル定義をメモリ上で管理するマップ
var tableSchemas = map[string]*TableDef{}

// RegisterTableはテーブル定義をメモリ上に登録する
func RegisterTable(def *TableDef) {
    tableSchemas[def.Name] = def
}

// SaveTableSchemaはテーブル定義をJSONでファイル保存する
func SaveTableSchema(def *TableDef) error {
    data, err := json.MarshalIndent(def, "", "  ")
    if err != nil {
        return err
    }
    filename := def.Name + ".schema"
    return os.WriteFile(filename, data, 0644)
}

// LoadTableSchemaはファイルからテーブル定義を読み込む
func LoadTableSchema(tableName string) (*TableDef, error) {
    filename := tableName + ".schema"
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    var def TableDef
    if err := json.Unmarshal(data, &def); err != nil {
        return nil, err
    }
    return &def, nil
} 
