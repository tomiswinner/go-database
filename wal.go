package main

import (
	"os"
	"time"
)

type WALEntry struct {
	TransactionID string
	LSN int64
	Operation OpType
	TableName string
	Data string // 面倒なので string で実装
	TimeStamp int64
}

// wal ファイル全体の管理
// めんどくさいので wal は csv で実装(本来は byte 列とかな気がする)
type WALManager struct {
	walFile *os.File
	mutex sync.Mutex // ファイル操作の排他制御
	walPath string
	latestLSN int64
}

// ファクトリ
func (db *WALEntry) NewWALEntry(tx *Transaction, operation OpType, tableName string, data []byte, wm *WALManager) WALEntry {
	entry := WALEntry{
		TransactionID: tx.ID,
		LSN: wm.latestLSN,
		Operation: operation,
		TableName: tableName,
		Data: data,
	}
	return entry
}

// ファクトリ
func (wl *WALManager) NewWALManager(walPath string) *WALManager {
	walFile, err := os.OpenFile("./wal.log", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to open WAL file: %v", err)
	}
	return &WALManager{
		walFile: walFile,
		walPath: walPath,
	}
}

// LSN を取得する関数
func (wm *WALManager) getLSN() (int64, error) {
	// 既存の wal ファイルが存在するか確認
	// if _, err := os.Stat(wm.walPath); os.IsNotExist(err) { のような os.Stat だと、他プロセスが削除したりするケースも出てくるので開いた方が良い
	// TODO: ここは IsNotExist 以外のエラーも出る可能性があるので、それもハンドリングする必要がある
	f, err := os.OpenFile(wm.walPath, os.O_CREATE|os.O_RDWR, 0644); os.IsNotExist(err) {
		return 0, errors.New("WALファイルが存在しません")
	}
	defer f.Close()

	// TODO:  本当は後ろから探すほうが効率的かも
	// 既存の wal の最新の LSN を取得
	var lastLine string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return 0, errors.New("WALファイルの読み込みに失敗しました")
	}

	// TODO: 末尾行がクラッシュしているケースを考慮していない
	lsn, err := strconv.ParseInt(lastLine, 10, 64)
	if err != nil {
		return 0, errors.New("LSNのパースに失敗しました")
	}
	return lsn + 1, nil
}

