package db

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"
)

// DbAdapter
type DbHandle struct {
	master    *sql.DB   // Master database.
	slavers   []*sql.DB // Slave database.
	slaverCnt int       // Number of slave database.
}

// NewDbHander return DbHandle object
func NewDbHandle() *DbHandle {
	return &DbHandle{}
}

// Open set database configuration.
// driverName: driver name, eg: mysql.
// maxConn: maximum connection.
// maxIdle: Maximum number of idle connections.
// maxLife: The maximum time interval that can be reused, in seconds, will be permanently reused if it is less than 0.
// configs: Connection string, the first is the master database, the other is the slave database.
func (h *DbHandle) Open(driverName string, maxConn, maxIdle int, maxLife int64, configs ...string) error {
	if driverName == "" {
		return errors.New("db: Driver Name is empty")
	} else if len(configs) == 0 {
		return errors.New("db: Config is empty")
	}

	if maxConn < 0 {
		maxConn = 0
	}

	var err error

	// master database
	h.master, err = sql.Open(driverName, configs[0])
	if err != nil {
		return err
	}
	h.master.SetMaxOpenConns(maxConn)
	if maxIdle >= 0 {
		h.master.SetMaxIdleConns(maxIdle)
	}
	h.master.SetConnMaxLifetime(time.Duration(maxLife) * time.Second)

	// slave database
	h.slaverCnt = len(configs) - 1
	if h.slaverCnt > 0 {
		for i := 1; i <= h.slaverCnt; i++ {
			t, err1 := sql.Open(driverName, configs[i])
			if err1 != nil {
				return err1
			}
			t.SetMaxOpenConns(maxConn)
			t.SetMaxIdleConns(maxIdle)
			t.SetConnMaxLifetime(time.Duration(maxLife) * time.Second)

			h.slavers = append(h.slavers, t)
		}
	}

	return nil
}

// GetMaster return master database.
func (h *DbHandle) GetMaster() *sql.DB {
	return h.master
}

// GetSlave return slave database.
// Return master database if hasn't salve database.
// Return to the n slave database if n slave database is exist.
// Other cases, return a random slave library.
func (h *DbHandle) GetSlave(n ...int) *sql.DB {
	if h.slaverCnt <= 0 {
		return h.master
	} else if h.slaverCnt == 1 {
		return h.slavers[0]
	}

	// n exist
	if len(n) > 0 && n[0] >= 0 && n[0] < h.slaverCnt {
		return h.slavers[n[0]]
	}

	// random
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(h.slaverCnt)

	return h.slavers[i]
}

// FetchOne returns the first line data, query from slave dbtabase.
func (h *DbHandle) FetchOne(sqlStr string, args ...interface{}) (map[string]string, error) {
	db := h.GetSlave()
	if db == nil {
		return nil, errors.New("db: Slave DB is nil")
	}

	return h.queryOne(db, sqlStr, args...)
}

// FetchOneMaster returns the first line data, query from master dbtabase.
func (h *DbHandle) FetchOneMaster(sqlStr string, args ...interface{}) (map[string]string, error) {
	db := h.GetMaster()
	if db == nil {
		return nil, errors.New("db: Master DB is nil")
	}

	return h.queryOne(db, sqlStr, args...)
}

// queryOne returns the first line data.
func (h *DbHandle) queryOne(db *sql.DB, sqlStr string, args ...interface{}) (map[string]string, error) {
	rows, err := db.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// all fields
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make(map[string]string, colCnt)
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		for i, col := range values {
			if col == nil {
				res[columns[i]] = ""
			} else {
				res[columns[i]] = string(col)
			}
		}

		break
	}

	return res, nil
}

// FetchAll returns all data, query from slave dbtabase.
func (h *DbHandle) FetchAll(sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	db := h.GetSlave()
	if db == nil {
		return nil, errors.New("db: Slave DB is nil")
	}

	return h.queryAll(db, sqlStr, args...)
}

// FetchAllMaster returns all data, query from master dbtabase.
func (h *DbHandle) FetchAllMaster(sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	db := h.GetMaster()
	if db == nil {
		return nil, errors.New("db: Master DB is nil")
	}

	return h.queryAll(db, sqlStr, args...)
}

// FetchAll returns all data.
func (h *DbHandle) queryAll(db *sql.DB, sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	rows, err := db.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// all fields
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make([]map[string]string, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		mapVal := make(map[string]string, colCnt)
		for i, col := range values {
			if col == nil {
				mapVal[columns[i]] = ""
			} else {
				mapVal[columns[i]] = string(col)
			}
		}
		res = append(res, mapVal)
	}

	return &res, nil
}

// Insert add data, don't support transaction.
func (h *DbHandle) Insert(sqlStr string, args ...interface{}) (int64, error) {
	db := h.GetMaster()
	if db == nil {
		return -1, errors.New("db: Master DB is nil")
	}

	stmtIns, err := db.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	res, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

// Exec update and delete data, don't support transaction.
func (h *DbHandle) Exec(sqlStr string, args ...interface{}) (int64, error) {
	db := h.GetMaster()
	if db == nil {
		return -1, errors.New("db: Master DB is nil")
	}

	stmtIns, err := db.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return result.RowsAffected()
}

// FetchOne returns the first line data, query from slave dbtabase, support transaction.
func (h *DbHandle) TxFetchOne(tx *sql.Tx, sqlStr string, args ...interface{}) (map[string]string, error) {
	rows, err := tx.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// all fields
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make(map[string]string, colCnt)
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		for i, col := range values {
			if col == nil {
				res[columns[i]] = ""
			} else {
				res[columns[i]] = string(col)
			}
		}

		break
	}

	return res, nil
}

// FetchAll returns all data, query from slave dbtabase, support transaction.
func (h *DbHandle) TxFetchAll(tx *sql.Tx, sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	rows, err := tx.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// all fields
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make([]map[string]string, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		mapVal := make(map[string]string, colCnt)
		for i, col := range values {
			if col == nil {
				mapVal[columns[i]] = ""
			} else {
				mapVal[columns[i]] = string(col)
			}
		}
		res = append(res, mapVal)
	}

	return &res, nil
}

// TxInsert add data, support transaction.
func (h *DbHandle) TxInsert(tx *sql.Tx, sqlStr string, args ...interface{}) (int64, error) {
	stmtIns, err := tx.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	res, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

// TxExec update and delete data, support transaction.
func (h *DbHandle) TxExec(tx *sql.Tx, sqlStr string, args ...interface{}) (int64, error) {
	stmtIns, err := tx.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return result.RowsAffected()
}

// Begin start transaction, operation master database.
func (h *DbHandle) Begin() (*sql.Tx, error) {
	db := h.GetMaster()
	if db == nil {
		return nil, errors.New("db: Master DB is nil")
	}

	return db.Begin()
}

// Commit commit transaction, operation master database.
func (h *DbHandle) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

// Rollback rollback transaction, operation master database.
func (h *DbHandle) Rollback(tx *sql.Tx) error {
	return tx.Rollback()
}

// Close close connect.
func (h *DbHandle) Close() {
	if h.master != nil {
		h.master.Close()
	}

	if h.slaverCnt > 0 {
		for _, db := range h.slavers {
			if db != nil {
				db.Close()
			}
		}
	}

}
