// トランザクション管理
// Go では構造体に他の場所からメソッド追加もOK
package main

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string

const (
	TransactionActive TransactionStatus = "ACTIVE"
	TransactionCommitted TransactionStatus = "COMMITTED"
	TransactionRolledBack TransactionStatus = "ROLLEDBACK"
)

type Transaction struct {
	ID string
	Status TransactionStatus
	StartTime int64
}



// ファクトリ
func (db *Database) newTransaction() *Transaction {
	tx := &Transaction{
		ID: uuid.New().String(),
		Status: TransactionActive,

		StartTime: time.Now().Unix(),
	}
	return tx
}



// トランザクション開始
func (db *Database) BeginTransaction() *Transaction {
	tx := db.newTransaction()
	entry := WALEntry.newWALEntry(tx, "BEGIN", "", nil)

}

// トランザクションコミット
func (db *Database) CommitTransaction(tx *Transaction) {
	tx.Status = TransactionCommitted
}
