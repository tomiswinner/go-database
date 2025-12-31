package main

type OpType string

// 以下のような状況で、どこまで適用したかを判断するために必要
// wal がディスクにさえのれば復元できるから、実装としては wal を fsync した時点で、commit 完了になる
// ディスク上の wal は必ずしもトランザクションが完了したものとは限らず、トランザクションの途中でクラッシュしてしまう場合もある
// [./study.md](./study.md) を参照

const (
	OpTypeInsert OpType = "INSERT"
	OpTypeUpdate OpType = "UPDATE"
	OpTypeDelete OpType = "DELETE"
	OpTypeBegin OpType = "BEGIN"
	OpTypeCommit OpType = "COMMIT"
	OpTypeRollback OpType = "ROLLBACK"
)
