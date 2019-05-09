package redisc

import (
	"fmt"
	"testing"
	"encoding/json"
	"time"
)

var gConfig = `{"addr":"127.0.0.1:6014,127.0.0.1:7014","auth":"123456","dialTimeout":"5","readTimeout":"1","writeTimeout":"1","poolSize":"100","minIdleConns":"10","maxConnAge":"3600","poolTimeout":"1","idleTimeout":"300","prefix":"le_"}`

func TestRediscCache(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	////////////////////////string test////////////////////////////
	k1 := "k1"
	v1 := "HelloWorld"
	err = adapter.Set(k1, v1, 20)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	var v11 string
	err, exist := adapter.Get(k1, &v11)
	if err != nil {
		t.Errorf("Redisc Get failed. err: %s.", err.Error())
		return
	} else if !exist {
		t.Errorf("Redisc Get failed. %s is not exist.", k1)
		return
	} else if v11 != v1 {
		t.Errorf("Redis Get failed. Got %s, expected %s.", v11, v1)
		return
	}

	isExist, err := adapter.IsExist(k1)
	if err != nil {
		t.Errorf("Redisc Get IsExist. err: %s.", err.Error())
		return
	} else if !isExist {
		t.Error("Redisc Get failed. Got false, expected true.")
		return
	}

	err = adapter.Del(k1)
	if err != nil {
		t.Errorf("Redisc Delete failed. err: %s.", err.Error())
		return
	}

	isExist, err = adapter.IsExist(k1)
	if err != nil {
		t.Errorf("Redisc Get IsExist. err: %s.", err.Error())
		return
	} else if isExist {
		t.Error("Redisc Get failed. Got true, expected false.")
		return
	}

	v11 = ""
	err, _ = adapter.Get(k1, &v11)
	if err != nil {
		t.Errorf("Redisc Get failed. err: %s.", err.Error())
		return
	} else if v11 != "" {
		t.Errorf("Redisc Get failed. Got %s, expected nil.", v11)
		return
	}

	////////////////////////////int32 test////////////////////////////
	k2 := "k2"
	v2 := 100
	err = adapter.Set(k2, int32(v2), 30)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	var v22 int
	err, _ = adapter.Get(k2, &v22)
	if err != nil {
		t.Errorf("Redisc Get failed. err: %s.", err.Error())
		return
	} else if v22 != v2 {
		t.Errorf("Redisc Get failed. Got %d, expected %d.", v22, v2)
		return
	}

	////////////////////////////float64 test////////////////////////////
	k3 := "k3"
	v3 := 100.01
	err = adapter.Set(k3, v3, 30)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	var v33 float64
	err, _ = adapter.Get(k3, &v33)
	if err != nil {
		t.Errorf("Redisc Get failed. err: %s.", err.Error())
		return
	} else if v33 != v3 {
		t.Errorf("Redisc Get failed. Got %f, expected %f.", v33, v3)
		return
	}

	////////////////////////////Incr and Decr test////////////////////////////
	k5 := "k5"
	v5 := 100
	err = adapter.Set(k5, v5, 30)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	var v55 int
	err, _ = adapter.Get(k5, &v55)
	if err != nil {
		t.Errorf("Redisc Get failed. err: %s.", err.Error())
		return
	} else if v55 != v5 {
		t.Errorf("Redisc Get failed. Got %d, expected %d.", v55, v5)
		return
	}

	newVal5, _ := adapter.Incr(k5)
	if newVal5 != 101 {
		t.Errorf("Redisc Incr failed. Got %d, expected %d.", newVal5, 101)
		return
	}

	newVal5, _ = adapter.Decr(k5)
	if newVal5 != 100 {
		t.Errorf("Redisc Incr failed. Got %d, expected %d.", newVal5, 100)
		return
	}

	newVal5, _ = adapter.Incr(k5, 10)
	if newVal5 != 110 {
		t.Errorf("Redisc Incr failed. Got %d, expected %d.", newVal5, 110)
		return
	}

	newVal5, _ = adapter.Decr(k5, 10)
	if newVal5 != 100 {
		t.Errorf("Redisc Incr failed. Got %d, expected %d.", newVal5, 100)
		return
	}

	////////////////////////////Hashtable test////////////////////////////
	k6 := "addr"
	f6 := "google"
	v6 := "www.google.com"
	adapter.HSet(k6, "baidu", "www.baidu.com", 60)
	adapter.HSet(k6, "le", "www.le.com", 60)
	r, err := adapter.HSet(k6, f6, v6, 60)
	fmt.Println(r)
	if err != nil {
		t.Errorf("Redisc HSet failed. err: %s.", err.Error())
		return
	}

	var v66 string
	err, _ = adapter.HGet(k6, f6, &v66)
	if err != nil {
		t.Errorf("Redisc HGet failed. err: %s.", err.Error())
		return
	} else if v66 != v6 {
		t.Errorf("Redisc HGet failed. Got %s, expected %s.", v66, v6)
		return
	}

	// HGetAll
	fmt.Println("=== HGetAll Begin ===")
	v77, err := adapter.HGetAll(k6)
	if err != nil {
		t.Errorf("Redisc HGetAll failed. err: %s.", err.Error())
		return
	}
	for k, v := range v77 {
		var val string
		//json.Unmarshal(v.([]byte), &val)
		val = v.(string)
		fmt.Println(k, val)
	}
	fmt.Println("=== HGetAll End ===")

	// HMGet
	fmt.Println("=== HMGet Begin ===")
	v99, err := adapter.HMGet(k6, "google", "baidu", "le")
	if err != nil {
		t.Errorf("Redisc HMGet failed. err: %s.", err.Error())
		return
	}
	for k, v := range v99 {
		if v == nil {
			fmt.Println(k, v)
			continue
		}
		var val string
		//json.Unmarshal(v.([]byte), &val)
		val = v.(string)
		fmt.Println(k, val)
	}
	fmt.Println("=== HMGet End ===")

	// HVals
	v88, err := adapter.HVals(k6)
	if err != nil {
		t.Errorf("Redisc HVals failed. err: %s.", err.Error())
		return
	}
	for _, v := range v88 {
		//var val string
		//json.Unmarshal(v.([]byte), &val)
		//fmt.Println(val)
		fmt.Println(string(v.([]byte)))
	}

	err = adapter.HDel(k6, f6, "baidu")
	if err != nil {
		t.Errorf("Redisc HDel failed. err: %s.", err.Error())
		return
	}

	///////////////////////HIncr and HDecr test //////////////////////
	k7 := "count"
	f7 := "aaa"
	fmt.Println("=== HIncr Begin ===")
	r7, err := adapter.HIncr(k7, f7, 2)
	if err != nil {
		t.Errorf("Redisd HIncr failed. err: %s.", err.Error())
		return
	}
	fmt.Println("test HIncr:", k7, f7, r7)
	fmt.Println("=== HIncr End ===")

	fmt.Println("=== HDecr Begin ===")
	r7, err = adapter.HDecr(k7, f7, 2)
	if err != nil {
		t.Errorf("Redisd HDecr failed. err: %s.", err.Error())
		return
	}
	fmt.Println("test HDecr:", k7, f7, r7)
	fmt.Println("=== HDecr End ===")

	err = adapter.Del(k7)
	if err != nil {
		t.Errorf("Redisd HDel failed. err: %s.", err.Error())
		return
	}

	////////////////////////ClearAll test////////////////////////////
	//err = adapter.ClearAll()
	//if err != nil {
	//	t.Errorf("Redisc ClearAll failed. err: %s.", err.Error())
	//	return
	//}
}

// TestRediscHash
func TestRediscHash(t *testing.T) {
	adapter := &RediscCache{}
	err := adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	// HMSet
	fmt.Println("=== HMSet Begin ===")
	key := "website"
	fields := map[string]interface{}{
		"google": "www.google.com",
		"yahoo":  "www.yahoo.com",
		"baidu":  "www.baidu.com",
	}
	err = adapter.HMSet(key, fields, 10)
	if err != nil {
		t.Errorf("Redisc HMSet failed. err: %s.", err.Error())
		return
	}
	fmt.Println("=== HMSet End ===")

	// HMGet
	fmt.Println("=== HMGet Begin ===")
	val, err := adapter.HMGet(key, "google", "baidu", "le")
	if err != nil {
		t.Errorf("Redisc HMGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(val)
	fmt.Println("=== HMGet End ===")
}

// TestRedisMulti
func TestRediscMulti(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	mList := make(map[string]interface{})
	mList["k1"] = "val1111"
	mList["k2"] = "val2222"
	mList["k3"] = "val3333"
	mList["k4"] = "val4444"
	err = adapter.MSet(mList, 60)
	if err != nil {
		t.Errorf("Redisc MSet failed. err: %s.", err.Error())
		return
	}

	mList2, err := adapter.MGet("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Redisc MGet failed. err: %s.", err.Error())
		return
	}

	var v1, v2, v3, v4 string
	if mList2["k1"] != "" {
		//json.Unmarshal(mList2["k1"].([]byte), &v1)
		v1 = mList2["k1"].(string)
	}
	if mList2["k2"] != "" {
		//json.Unmarshal(mList2["k2"].([]byte), &v2)
		v2 = mList2["k2"].(string)
	}
	if mList2["k3"] != "" {
		//json.Unmarshal(mList2["k3"].([]byte), &v3)
		v3 = mList2["k3"].(string)
	}
	if mList2["k4"] != "" {
		//json.Unmarshal(mList2["k4"].([]byte), &v4)
		v4 = mList2["k4"].(string)
	}

	if v1 != mList["k1"] || v2 != mList["k2"] || v3 != mList["k3"] || v4 != mList["k4"] {
		t.Errorf("Redisc MGet failed. v1:%s v2:%s v3:%s v4:%s.", v1, v2, v3, v4)
		return
	}

	err = adapter.MDel("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Redisc MDelete failed. err: %s.", err.Error())
		return
	}
}

// TestRediscSet test ordered set.
func TestRediscSet(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	key := "sets"
	// add
	n, err := adapter.ZSet(key, 60, 5.0, "val5", 3.5, "val3.5", 1.0, "100", 4.0, 400, 0.5, "val0.5", 1.0, "val1")
	fmt.Println(n)
	if err != nil {
		t.Errorf("Redisc ZSet failed. err: %s.", err.Error())
		return
	}

	// query, increment
	res, err := adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redisc ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	res, err = adapter.ZGet(key, 0, -1, false, false)
	if err != nil {
		t.Errorf("Redisc ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// query, decrement
	res, err = adapter.ZGet(key, 0, -1, true, true)
	if err != nil {
		t.Errorf("Redisc ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	res, err = adapter.ZGet(key, 0, -1, false, true)
	if err != nil {
		t.Errorf("Redisc ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// cardinality
	n, err = adapter.ZCard(key)
	if err != nil {
		t.Errorf("Redisc ZCard failed. err: %d.", n)
		return
	}
	if n != 6 {
		t.Errorf("Redisc ZCard failed. Got %d, expected 6.", n)
		return
	}

	// delete
	n, err = adapter.ZDel(key, "val3.5", "400")
	if err != nil {
		t.Errorf("Redisc ZDel failed. err: %s.", err.Error())
		return
	}
	fmt.Println(n)

	// query
	res, err = adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redisc ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// cardinality
	n, err = adapter.ZCard(key)
	if err != nil {
		t.Errorf("Redisc ZCard failed. err: %d.", n)
		return
	} else if n != 4 {
		t.Errorf("Redisc ZCard failed. Got %d, expected 4.", n)
		return
	}
}

// TestZRemRangeByRank test ZRemRangeByRank function.
func TestZRemRangeByRank(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	key := "salary1"
	// add
	n, err := adapter.ZSet(key, 60, 2000.0, "jack", 5000.0, "tom", 3500.0, "peter")
	fmt.Println(n)
	if err != nil {
		t.Errorf("Redis ZSet failed. err: %s.", err.Error())
		return
	}

	// query
	res, err := adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// delete
	n, err = adapter.ZRemRangeByRank(key, 0, 1)
	if err != nil {
		t.Errorf("Redis ZDel failed. err: %s.", err.Error())
		return
	}
	fmt.Println(n)

	// query
	res, err = adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)
}

// TestZRemRangeByScore test ZRemRangeByScore function.
func TestZRemRangeByScore(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	key := "salary2"
	// add
	n, err := adapter.ZSet(key, 60, 2000.0, "jack", 5000.0, "tom", 3500.0, "peter")
	fmt.Println(n)
	if err != nil {
		t.Errorf("Redis ZSet failed. err: %s.", err.Error())
		return
	}

	// query
	res, err := adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// delete
	n, err = adapter.ZRemRangeByScore(key, "1500", "3500")
	if err != nil {
		t.Errorf("Redis ZDel failed. err: %s.", err.Error())
		return
	}
	fmt.Println(n)

	// query
	res, err = adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)
}

// TestTestZRemRangeByLex test ZRemRangeByLex function.
func TestZRemRangeByLex(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	key := "salary3"
	// add
	n, err := adapter.ZSet(key, 60, 0.0, "aaaa", 0.0, "b", 0.0, "c", 0.0, "d", 0.0, "e")
	fmt.Println(n)
	if err != nil {
		t.Errorf("Redis ZSet failed. err: %s.", err.Error())
		return
	}
	n, err = adapter.ZSet(key, 60, 0.0, "foo", 0.0, "zap", 0.0, "zip", 0.0, "ALPHA", 0.0, "alpha")
	fmt.Println(n)
	if err != nil {
		t.Errorf("Redis ZSet failed. err: %s.", err.Error())
		return
	}

	// query
	res, err := adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// delete
	n, err = adapter.ZRemRangeByLex(key, "[alpha", "[omega")
	if err != nil {
		t.Errorf("Redis ZDel failed. err: %s.", err.Error())
		return
	}
	fmt.Println(n)

	// query
	res, err = adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)
}

type User struct {
	Id   int
	Name string
}

// TestStruct
func TestStruct(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	sk1 := "k1"
	sv1 := User{
		Id:   1001,
		Name: "lixioaya",
	}
	err = adapter.Set(sk1, sv1, 10)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	var sv11 User
	err, _ = adapter.Get(sk1, &sv11)
	if err != nil {
		t.Errorf("Redisc Get failed. err: %s.", err.Error())
		return
	} else if sv11.Id != sv1.Id || sv11.Name != sv1.Name {
		t.Errorf("Redisc Get failed. id[%d] name[%s].", sv11.Id, sv11.Name)
		return
	}

	/////////////Hashtable////////////
	k6 := "addr"
	f6 := "google"
	v6 := User{
		Id:   1001,
		Name: "lixioaya",
	}

	_, err = adapter.HSet(k6, f6, v6, 60)
	if err != nil {
		t.Errorf("Redisc HSet failed. err: %s.", err.Error())
		return
	}

	var v66 User
	err, exist := adapter.HGet(k6, f6, &v66)
	if err != nil {
		t.Errorf("Redisc HGet failed. err: %s.", err.Error())
		return
	} else if !exist {
		t.Errorf("Redisc Get failed. %s - %s is not exist.", k6, f6)
		return
	} else if sv11.Id != sv1.Id || sv11.Name != sv1.Name {
		t.Errorf("Redisc Get failed. id[%d] name[%s].", sv11.Id, sv11.Name)
		return
	}
}

// TestRediscEncode test encode and decode.
func TestRediscEncode(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(`{"addr":"127.0.0.1:6014,127.0.0.1:7014","auth":"123456","dialTimeout":"5","readTimeout":"1","writeTimeout":"1","poolSize":"100","minIdleConns":"10","maxConnAge":"3600","poolTimeout":"1","idleTimeout":"300","prefix":"le_","encodeKey":"lxy123"}`)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	sk1 := "k1"
	sv1 := User{
		Id:   1001,
		Name: "lixioaya",
	}
	err = adapter.Set(sk1, sv1, 60, true)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	var sv11 User
	err, _ = adapter.Get(sk1, &sv11)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	} else if sv11.Id != sv1.Id || sv11.Name != sv1.Name {
		t.Errorf("Redisc Get failed. id[%d] name[%s].", sv11.Id, sv11.Name)
		return
	}

	///////// MSet MGet ////////

	mList := make(map[string]interface{})
	mList["k1"] = "val1111"
	mList["k2"] = "val2222"
	mList["k3"] = "val3333"
	mList["k4"] = "val4444"
	err = adapter.MSet(mList, 60, true)
	if err != nil {
		t.Errorf("Redisc MSet failed. err: %s.", err.Error())
		return
	}

	mList2, err := adapter.MGet("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Redisc MGet failed. err: %s.", err.Error())
		return
	}

	var v1, v2, v3, v4 string
	if mList2["k1"] != nil {
		v1 = mList2["k1"].(string)
	}
	if mList2["k2"] != nil {
		v2 = mList2["k2"].(string)
	}
	if mList2["k3"] != nil {
		v3 = mList2["k3"].(string)
	}
	if mList2["k4"] != nil {
		v4 = mList2["k4"].(string)
	}

	if v1 != mList["k1"] || v2 != mList["k2"] || v3 != mList["k3"] || v4 != mList["k4"] {
		t.Errorf("Redisc MGet failed. v1:%s v2:%s v3:%s v4:%s.", v1, v2, v3, v4)
		return
	}

	//////// int ///////

	sk2 := "k2"
	sv2 := 100
	err = adapter.Set(sk2, sv2, 60, true)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	var sv22 int
	err, _ = adapter.Get(sk2, &sv22)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	} else if sv2 != sv22 {
		t.Errorf("Redisc Get failed. id[%d] name[%d].", sv22, sv2)
		return
	}
}

////////// IJson test /////////////
type Item struct {
	uid  int32
	name string
}

func (this *Item) MarshalJSON() ([]byte, error) {
	fmt.Println("Item MarshalJSON")
	str := fmt.Sprintf(`{"uid":%d, "name":"%s"}`, this.uid, this.name)
	return []byte(str), nil
}

func (this *Item) UnmarshalJSON(data []byte) error {
	fmt.Println("Item UnmarshalJSON")

	val := make(map[string]interface{})
	json.Unmarshal(data, &val)
	uid, _ := val["uid"]
	this.uid = int32(uid.(float64))
	name, _ := val["name"]
	this.name = name.(string)
	return nil
}

// TestRediscIJson test IJson.
func TestRediscIJson(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	key := "k1"
	val1 := Item{
		uid:  1000,
		name: "nick",
	}
	// Set
	err = adapter.Set(key, &val1, 3600)
	if err != nil {
		t.Errorf("Redisc Set failed. err: %s.", err.Error())
		return
	}

	// Get
	val2 := Item{}
	err, _ = adapter.Get(key, &val2)
	if err != nil {
		t.Errorf("Redisc Get failed. err: %s.", err.Error())
		return
	} else if val1.uid != val2.uid || val1.name != val2.name {
		t.Errorf("Redisc Get failed. Got: %d-%s expected: %d-%s.", val2.uid, val2.name, val1.uid, val1.name)
		return
	}

	// Mset
	mList := make(map[string]interface{})
	mList["k1"] = &Item{
		uid:  1001,
		name: "nick1",
	}
	mList["k2"] = &Item{
		uid:  1002,
		name: "nick2",
	}
	mList["k3"] = &Item{
		uid:  1003,
		name: "nick3",
	}
	err = adapter.MSet(mList, 600)
	if err != nil {
		t.Errorf("Redisc MSet failed. err: %s.", err.Error())
		return
	}

	// HSet
	k6 := "addr"
	_, err = adapter.HSet(k6, "baidu", &Item{
		uid:  1001,
		name: "baidu",
	}, 60)
	if err != nil {
		t.Errorf("Redisc HSet failed. err: %s.", err.Error())
		return
	}
	_, err = adapter.HSet(k6, "le", &Item{
		uid:  1002,
		name: "leeco",
	}, 60)
	if err != nil {
		t.Errorf("Redisc HSet failed. err: %s.", err.Error())
		return
	}

	// HGet
	v66 := Item{}
	err, _ = adapter.HGet(k6, "baidu", &v66)
	if err != nil {
		t.Errorf("Redisc HGet failed. err: %s.", err.Error())
		return
	} else if v66.uid != 1001 || v66.name != "baidu" {
		t.Errorf("Redisc Get failed. Got: %d-%s expected: %d-%s.", v66.uid, v66.name, 1001, "baidu")
		return
	}
}

// TestRediscBit 测试bit相关函数
func TestRediscBit(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	key := "bits"
	val := 1
	res, err := adapter.SetBit(key, 0, val, 30)
	if err != nil {
		t.Errorf("SetBit failed. err: %s.", err.Error())
		return
	}
	fmt.Println("SetBit res:", res)

	res, err = adapter.GetBit(key, 0)
	if err != nil {
		t.Errorf("GetBit failed. err: %s.", err.Error())
		return
	}
	fmt.Println("GetBit res:", res)

	res, err = adapter.BitCount(key, nil)
	if err != nil {
		t.Errorf("BitCount failed. err: %s.", err.Error())
		return
	}
	fmt.Println("BitCount res:", res)
}

// TestRediscPipeline
func TestRediscPipeline(t *testing.T) {
	var err error
	adapter := &RediscCache{}
	err = adapter.Init(gConfig)
	if err != nil {
		t.Errorf("Redisc Init failed. err: %s.", err.Error())
		return
	}

	// no transaction
	pipe := adapter.Pipeline(false).Pipe
	key := "foo"
	r1 := pipe.Set(key, 100, 10*time.Second)
	r2 := pipe.Get(key)
	r3 := pipe.Incr(key)
	r4 := pipe.Del(key)
	_, err = pipe.Exec()
	if err != nil {
		fmt.Println("Exec err:", err)
		return
	}
	fmt.Println("r1:", r1.Val(), "r2:", r2.Val(), "r3:", r3.Val(), "r4:", r4.Val())

	// transaction
	pipe = adapter.Pipeline(false).Pipe
	r1 = pipe.Set(key, 100, 10*time.Second)
	r2 = pipe.Get(key)
	r3 = pipe.Incr(key)
	r4 = pipe.Del(key)
	_, err = pipe.Exec()
	if err != nil {
		fmt.Println("Exec err:", err)
		return
	}
	fmt.Println("r1:", r1.Val(), "r2:", r2.Val(), "r3:", r3.Val(), "r4:", r4.Val())
}
