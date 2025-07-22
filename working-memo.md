# 作業ログ

## プロジェクト初期化

### Go プロジェクトの設定

- `go mod init go-database` を実行し、Go プロジェクトを初期化

### 基本構造体の定義

- `storage.go` に User 構造体を定義（ID: int, Name: string）
- users.db ファイルを作成する CreateUsersTableFile 関数を実装

## ストレージエンジン実装

### CSV ベースのデータ保存

- `storage.go` に AppendUser 関数を実装（CSV 形式でユーザー追記保存）
- `storage.go` に ReadAllUsers 関数を実装（CSV 形式から全ユーザー読み込み）
- 各関数に詳しいコメントを追加（os.O_APPEND, ファイル操作の説明）

### ページング機能の追加

- `storage.go` に ReadUsersWithPaging 関数を実装（offset, limit 指定読み込み）
- `storage.go` に CountUsers 関数を実装（総レコード数取得）
- getUserFieldCount 関数を追加（リフレクションで構造体フィールド数を動的取得）

## SQL パーサー実装

### CREATE TABLE パーサー

- `parser.go` に ColumnDef, TableDef 構造体を定義
- `parser.go` に ParseCreateTable 関数を実装（正規表現による構文解析）

### INSERT パーサー

- `parser.go` に InsertDef 構造体を定義
- `parser.go` に ParseInsert 関数を実装（カラム名・値の抽出、シングルクォート除去）

### SELECT パーサー

- `parser.go` に SelectDef 構造体を定義
- `parser.go` に ParseSelect 関数を実装（SELECT \* と指定カラム選択の両方対応）

### パーサーテスト

- `parser_test.go` を作成
- CREATE TABLE と INSERT 文の包括的なテスト実装（正常ケース・エラーケース）
- 全テストが PASS することを確認

## スキーマ管理実装

### テーブル定義の永続化

- `schema.go` にテーブル定義の JSON 保存・読み込み機能を実装
- SaveTableSchema, LoadTableSchema, RegisterTable 関数を実装
- ioutil.WriteFile を os.WriteFile に修正（deprecated 対応）

## メイン機能統合

### CLI インターフェース

- `main.go` を更新し、CREATE TABLE 文を標準入力から受け付ける機能を実装
- 任意のテーブル名・カラム定義に対応、スキーマファイル（.schema）として保存
- handleSelect 関数を追加、表形式での見やすい出力を実現

### データベースエンジン化

- `main.go` に Database 構造体を追加（データベースエンジンの中心）
- CreateTable, Insert, Select メソッドを実装
- ExecuteSQL で統一的な SQL 実行インターフェースを提供
- エラーハンドリングを統一化

## インデックス実装

### B+Tree の基本実装

- `btree.go` を新規作成
- BTreeNode 構造体を実装（リーフ・内部ノード）
- Insert メソッドを実装（ノード分割含む）
- Search メソッドを実装（O(log n)検索）
- PrintTree を実装（デバッグ用ツリー構造表示）
- 主キー用インデックスの基盤を構築

### Database エンジンへのインデックス統合

- `main.go` の Database 構造体に indexes フィールドを追加
- CreateTable 時に主キー（第一カラム）用 B+Tree インデックスを自動作成
- Insert 時に B+Tree インデックスへキー・レコード位置を自動登録
- SHOW INDEX コマンドを追加（デバッグ用インデックス状況表示）

## WHERE 句とインデックス検索

### WHERE 句パーサー実装

- `parser.go` に WhereClause 構造体を追加
- `parser.go` の ParseSelect 関数を拡張して WHERE 句に対応
- parseWhere 内部関数を実装（=, >, <, >=, <= 演算子対応）
- SelectDef 構造体に WhereClause フィールドを追加

### インデックス活用検索

- `main.go` の Select 関数を大幅リファクタリング
- WHERE id = value の場合に B+Tree インデックス検索を実装（O(log n)）
- その他の条件では全件スキャンによるフィルタリング（O(n)）
- 検索方式（インデックス vs 全件スキャン）をコンソールに表示

### 関数分離によるリファクタリング

- Select 関数を 5 つの専門関数に分離
- selectUsersWithWhere（検索戦略決定）
- searchByIndex（B+Tree インデックス検索）
- searchByFullScan（全件スキャン検索）
- displayResults（結果表示）
- filterUsers（条件フィルタリング）

## 進捗管理

### 実装完了項目の記録

- `go-rdbms-plan.md` の実装済み項目をチェック済みにマーク
- Step 0-3 がほぼ完了、Step 4 も一部完了の状況を記録
