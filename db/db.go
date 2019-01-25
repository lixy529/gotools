package db

import (
	"errors"
)

// DbBase
type DbBase struct {
	dbs map[string]*DbHandle
}

// NewDbBase return DbBase object.
// configs: configuration information, eg:
//   dbName => db_cn
//   driverName => mysql
//   maxConn = 200
//   maxIdle = 100
//   maxLife = 21600
//   master = user:pwd@tcp(ip:port)/dbname?charset=utf8
//   slave1 = user:pwd@tcp(ip:port)/dbname?charset=utf8
//   slave2 = user:pwd@tcp(ip:port)/dbname?charset=utf8
func NewDbBase(configs ...map[string]interface{}) (*DbBase, error) {
	n := len(configs)
	if n == 0 {
		return nil, errors.New("db: Db config is empty ")
	}

	dbBase := &DbBase{}
	dbBase.dbs = make(map[string]*DbHandle)

	for _, config := range configs {
		dbName := config["dbName"].(string)
		var maxConn int = 0
		var maxIdle int = -1
		var maxLife int64 = 0
		var driverName string
		var configs []string
		if val, ok := config["driverName"]; ok && val != "" {
			driverName = val.(string)
		} else {
			return dbBase, errors.New("db: Driver name is empty")
		}
		// Must have a master database
		if val, ok := config["master"]; ok {
			configs = append(configs, val.(string))
		} else {
			return dbBase, errors.New("db: Master config is empty")
		}

		// The slave database can be empty
		if len(config["slaves"].([]string)) > 0 {
			configs = append(configs, config["slaves"].([]string)...)
		}

		if val, ok := config["maxConn"]; ok {
			maxConn, _ = val.(int)
		}
		if val, ok := config["maxIdle"]; ok {
			maxIdle, _ = val.(int)
		}
		if val, ok := config["maxLife"]; ok {
			maxLife, _ = val.(int64)
		}

		h := NewDbHandle()
		err := h.Open(driverName, maxConn, maxIdle, maxLife, configs...)
		if err != nil {
			return dbBase, err
		}
		dbBase.dbs[dbName] = h
	}

	return dbBase, nil
}

// Db return a database connection by dbName.
func (b *DbBase) Db(dbName ...string) *DbHandle {
	name := "db"
	if len(dbName) > 0 && dbName[0] != "" {
		name = dbName[0]
	}

	if db, ok := b.dbs[name]; ok {
		return db
	}
	return nil
}

// Close close all connection pools
func (b *DbBase) Close() {
	for _, h := range b.dbs {
		h.Close()
	}
}
