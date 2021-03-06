package db

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

// TestDbHandle
func TestDbHandle(t *testing.T) {
	h := NewDbHandle()
	var err error

	// Open
	err = h.Open("mysql", 200, 100, 21600, "root:root123@tcp(127.0.0.1:3306)/passport?charset=utf8", "root:root123@tcp(127.0.0.1:3306)/passport?charset=utf8", "root:root123@tcp(127.0.0.1:3306)/passport?charset=utf8")
	if err != nil {
		t.Errorf("handle.Open error, [%s]", err.Error())
		return
	}

	// Insert
	userId, err := h.Insert("insert into t_test(name,age,addr) values(?,?,?)", "demo", 20, "beijing")
	if err != nil {
		t.Errorf("handle.Insert error, [%s]", err.Error())
		return
	}
	t.Logf("userId = %d\n", userId)

	// FetchOne
	res, err := h.FetchOne("select * from t_test where name=?", "lixy")
	if err != nil {
		t.Errorf("handle.FetchOne error, [%s]", err.Error())
		return
	}
	t.Log(res)

	// FetchAll
	res1, err := h.FetchAll("select * from t_test")
	if err != nil {
		t.Errorf("handle.FetchAll error, [%s]", err.Error())
		return
	}
	for _, val := range *res1 {
		t.Log(val)
	}

	// Exec update
	cnt, err := h.Exec("update t_test set age=? where name=?", 200, "demo")
	if err != nil {
		t.Errorf("handle.Insert error, [%s]", err.Error())
		return
	}
	t.Logf("cnt = %d\n", cnt)

	// FetchOneMaster
	res, err = h.FetchOneMaster("select * from t_test where name=?", "lixy")
	if err != nil {
		t.Errorf("handle.FetchOneMaster error, [%s]", err.Error())
		return
	}
	t.Log(res)

	// FetchAllMaster
	res1, err = h.FetchAllMaster("select * from t_test where name=?", "lixy")
	if err != nil {
		t.Errorf("handle.FetchAllMaster error, [%s]", err.Error())
		return
	}
	t.Log(res)

	// Exec delete
	cnt, err = h.Exec("delete from t_test where name=?", "demo")
	if err != nil {
		t.Errorf("handle.Exec error, [%s]", err.Error())
		return
	}
	t.Logf("cnt = %d\n", cnt)
}

// TestTxDbHandle test transaction.
func TestTxDbHandle(t *testing.T) {
	h := NewDbHandle()

	// Open
	err := h.Open("mysql", 200, 100, 21600, "root:root123@tcp(127.0.0.1:3306)/passport?charset=utf8", "root:root123@tcp(127.0.0.1:3306)/passport?charset=utf8", "root:root123@tcp(127.0.0.1:3306)/passport?charset=utf8")
	if err != nil {
		t.Errorf("handle.Open error, [%s]", err.Error())
		return
	}

	///////////////////////////////////Insert///////////////////////////////////////

	t.Log("/////////TxInsert///////////")
	tx, err := h.Begin()
	if err != nil {
		t.Errorf("handle.Begin error, [%s]", err.Error())
		return
	}

	// TxInsert
	userId, err := h.TxInsert(tx, "insert into t_test(name,age,addr) values(?,?,?)", "txTest", 10, "beijing")
	if err != nil {
		h.Rollback(tx)
		t.Errorf("handle.TxInsert error, [%s]", err.Error())
		return
	}
	t.Logf("userId = %d\n", userId)

	// FetchOne
	res, err := h.TxFetchOne(tx, "select * from t_test where name=?", "txTest")
	if err != nil {
		t.Errorf("handle.FetchOne11 error, [%s]", err.Error())
		return
	}
	t.Log(res)
	h.Commit(tx)

	////////////////////////////////////Update//////////////////////////////////////

	t.Log("/////////TxExec Update///////////")
	tx, err = h.Begin()
	if err != nil {
		t.Errorf("handle.Begin error, [%s]", err.Error())
		return
	}

	// TxExec update
	cnt, err := h.TxExec(tx, "update t_test set age=? where name=?", 20, "txTest")
	if err != nil {
		h.Rollback(tx)
		t.Errorf("handle.Insert error, [%s]", err.Error())
		return
	}
	t.Logf("cnt = %d\n", cnt)

	// FetchOne
	res, err = h.TxFetchOne(tx, "select * from t_test where name=?", "txTest")
	if err != nil {
		t.Errorf("handle.FetchOne error, [%s]", err.Error())
		return
	}
	t.Log(res)
	h.Commit(tx)

	////////////////////////////////////Delete//////////////////////////////////////
	t.Log("/////////TxExec Delete///////////")
	tx, err = h.Begin()
	if err != nil {
		t.Errorf("handle.Begin error, [%s]", err.Error())
		return
	}

	// Exec delete
	cnt, err = h.TxExec(tx, "delete from t_test where name=?", "txTest")
	if err != nil {
		h.Rollback(tx)
		t.Errorf("handle.TxExec error, [%s]", err.Error())
		return
	}
	t.Logf("cnt = %d\n", cnt)

	// FetchOne
	res, err = h.TxFetchOne(tx, "select * from t_test where name=?", "txTest")
	if err != nil {
		t.Errorf("handle.FetchOne error, [%s]", err.Error())
		return
	}
	h.Commit(tx)
	t.Log(res)

	///////////////////////////////////TxFetchAll///////////////////////////////////////

	t.Log("/////////TxFetchAll///////////")
	tx, err = h.Begin()
	if err != nil {
		t.Errorf("handle.Begin error, [%s]", err.Error())
		return
	}

	// TxInsert
	cnt, err = h.TxExec(tx, "update t_test set name =? where name=?", "tttt", "lixy")
	if err != nil {
		h.Rollback(tx)
		t.Errorf("handle.TxExec error, [%s]", err.Error())
		return
	}
	t.Logf("cnt = %d\n", cnt)

	// TxFetchAll
	res1, err := h.TxFetchAll(tx, "select * from t_test")
	if err != nil {
		t.Errorf("handle.FetchAll error, [%s]", err.Error())
		return
	}
	for _, val := range *res1 {
		t.Log(val)
	}

	h.Rollback(tx)
}
