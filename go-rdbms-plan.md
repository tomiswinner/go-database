---
alwaysApply: true
---

# 目的

Go を使って簡易的な RDBMS（PostgreSQL/MySQL 風）を 0 から自作すること

# 機能リスト（段階別）

## Step 0: セットアップ

- [x] CLI インターフェース（REPL モード or コマンド実行）
- [x] 最小限のエントリポイント（main.go）

## Step 1: ストレージエンジン（ファイルベース）

- [x] テーブルごとのデータファイル作成（例: `users.db`）
- [x] データ行の構造体設計
- [x] データの追記保存（`INSERT`）
- [x] データの全件読み込み（`SELECT *`）

## Step 2: SQL パーサ（構文解析）

- [x] `CREATE TABLE` の構文解釈
- [x] `INSERT INTO` の構文解釈
- [x] `SELECT` の構文解釈
- [x] `WHERE` 句（=, >, <, >=, <= 演算子対応）

## Step 3: スキーマとメタ情報

- [x] テーブル定義（カラム名・型）管理
- [x] スキーマ情報の永続化
- [x] 存在しないテーブル・カラムへのバリデーション

## Step 4: クエリエンジン

- [x] `SELECT * FROM ... WHERE ...` 条件一致検索
- [x] 複数レコードの返却
- [x] カラム単位での出力（`SELECT id, name`）

## Step 5: 最低限のエンジン最適化

- [x] ページング（一定数ごとの読み書き）
- [x] B+Tree インデックス（主キー用）
- [x] インデックス活用検索（WHERE id = value）

## Step 6: トランザクション / 永続性

- [ ] Write Ahead Log（WAL）風の仕組み
- [ ] 簡易なロールバック処理（atomic insert）
- [ ] トランザクションを張って commit できるように

## Step 7: 拡張機能（任意）

- [ ] `DELETE FROM` 機能
- [ ] `UPDATE` 機能
- [ ] `ORDER BY`, `LIMIT` 対応
- [ ] JOIN 構文（Nested Loop Join から）

# 注意点

- シンプルな構成から始め、段階的に拡張すること
- バックエンドストレージは最初はファイル単体で実装
- SQL パーサ部分は `participle` などの利用も検討
