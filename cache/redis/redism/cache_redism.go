// Master-slave mode
package redism

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/lixy529/gotools/cache"
	"strconv"
	"time"
)

const NOT_EXIST = "redis: nil"

// RedismCache
type RedismCache struct {
	master *RedisPool // Master.
	slave  *RedisPool // Slave.

	mAddr  string // Master address, eg: 127.0.0.1:1900,127.0.0.2:1900.
	mDbNum int    // Master db number.
	mAuth  string // Master auth password.

	sAddr  string // Slave address, eg: 127.0.0.1:1900,127.0.0.2:1900.
	sDbNum int    // Slave db number.
	sAuth  string // Slave auth password.

	dialTimeout  time.Duration // Connection timeout, in seconds, default 5 seconds.
	readTimeout  time.Duration // Read timeout, in seconds, -1-No timeout，0-Use the default 3 seconds.
	writeTimeout time.Duration // Write timeout, in seconds, default use readTimeout.
	poolSize     int           // Number of connections per node connection pool, default is ten times the number of CPUs.
	minIdleConns int           // Minimum number of idle connections, default is 0.
	maxConnAge   time.Duration // Maximum connection time, in seconds, default is 0.
	poolTimeout  time.Duration // Waiting time when all connections are busy, default is readTimeout+1.
	idleTimeout  time.Duration // Maximum idle time, in seconds, default is 5 minutes.

	prefix    string // Key prefix.
	encodeKey []byte // Encode key, length of 16 times, using Aes encryption
}

// NewRedismCache return RedismCache object.
func NewRedismCache() cache.Cache {
	return &RedismCache{}
}

// Init initialization configuration.
// Eg:
// {
// "master":{"addr":"127.0.0.1:6379","dbNum":"0","auth":"xxxxx"},
// "slave":{"addr":"127.0.0.2:6379","dbNum":"1","auth":"xxxxx"},
// "dialTimeout":"5",
// "readTimeout":"5",
// "writeTimeout":"5",
// "poolSize":"5",
// "minIdleConns":"5",
// "maxConnAge":"5",
// "poolTimeout":"5",
// "idleTimeout":"5",
// "prefix":"le_",
// "encodeKey":"abcdefghij123456",
// }
func (rc *RedismCache) Init(config string) error {
	var mapCfg map[string]string
	var err error

	err = json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("RedismCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	// Connection timeout
	dialTimeout, err := strconv.Atoi(mapCfg["dialTimeout"])
	if err != nil || dialTimeout < 0 {
		rc.dialTimeout = 5
	} else {
		rc.dialTimeout = time.Duration(dialTimeout)
	}

	// Read timeout
	readTimeout, err := strconv.Atoi(mapCfg["readTimeout"])
	if err != nil {
		rc.readTimeout = 3
	} else if readTimeout < 0 {
		rc.readTimeout = -1
	} else {
		rc.readTimeout = time.Duration(readTimeout)
	}

	// Write timeout
	writeTimeout, err := strconv.Atoi(mapCfg["writeTimeout"])
	if err != nil {
		rc.writeTimeout = rc.readTimeout
	} else if writeTimeout < 0 {
		rc.writeTimeout = -1
	} else {
		rc.writeTimeout = time.Duration(writeTimeout)
	}

	// Number of connections per node
	poolSize, err := strconv.Atoi(mapCfg["poolSize"])
	if err != nil || poolSize < 0 {
		rc.poolSize = 0
	} else {
		rc.poolSize = poolSize
	}

	// Minimum number of idle connections
	minIdleConns, err := strconv.Atoi(mapCfg["minIdleConns"])
	if err != nil || minIdleConns < 0 {
		rc.minIdleConns = 0
	} else {
		rc.minIdleConns = minIdleConns
	}

	// Maximum connection time
	maxConnAge, err := strconv.Atoi(mapCfg["maxConnAge"])
	if err != nil || maxConnAge < 0 {
		rc.maxConnAge = 0
	} else {
		rc.maxConnAge = time.Duration(maxConnAge)
	}

	// Wait timeout
	poolTimeout, err := strconv.Atoi(mapCfg["poolTimeout"])
	if err != nil || poolTimeout < 0 {
		rc.poolTimeout = rc.readTimeout + 1
	} else {
		rc.poolTimeout = time.Duration(poolTimeout)
	}

	// Maximum idle time
	idleTimeout, err := strconv.Atoi(mapCfg["idleTimeout"])
	if err != nil || idleTimeout < 0 {
		rc.idleTimeout = 300
	} else {
		rc.idleTimeout = time.Duration(idleTimeout)
	}

	// Key prefix
	if prefix, ok := mapCfg["prefix"]; ok {
		rc.prefix = prefix
	}

	// Encode key
	if tmp, ok := mapCfg["encodeKey"]; ok && tmp != "" {
		rc.encodeKey = []byte(tmp)
	}

	// Master address
	rc.mAddr = mapCfg["mAddr"]
	if rc.mAddr == "" {
		return errors.New("RedismCache: Master addr is empty")
	}

	dbNum, err := strconv.Atoi(mapCfg["mDbNum"])
	if err != nil {
		rc.mDbNum = 0
	} else {
		rc.mDbNum = dbNum
	}

	rc.mAuth = mapCfg["mAuth"]

	// Master pool
	rc.master = NewRedisPool(rc.mAddr, rc.mAuth, rc.mDbNum, rc.dialTimeout, rc.readTimeout, rc.writeTimeout, rc.poolSize, rc.minIdleConns, rc.maxConnAge, rc.poolTimeout, rc.idleTimeout, rc.prefix, rc.encodeKey)

	// Slave address
	rc.sAddr = mapCfg["sAddr"]
	if rc.sAddr == "" {
		rc.slave = rc.master
	} else {
		dbNum, err := strconv.Atoi(mapCfg["sDbNum"])
		if err != nil {
			rc.sDbNum = 0
		} else {
			rc.sDbNum = dbNum
		}

		rc.sAuth = mapCfg["sAuth"]
		rc.slave = NewRedisPool(rc.sAddr, rc.sAuth, rc.sDbNum, rc.dialTimeout, rc.readTimeout, rc.writeTimeout, rc.poolSize, rc.minIdleConns, rc.maxConnAge, rc.poolTimeout, rc.idleTimeout, rc.prefix, rc.encodeKey)
	}

	return nil
}

// Set set a cache value.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (rc *RedismCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
	return rc.master.Set(key, val, expire, encode...)
}

// Get get a cache value.
func (rc *RedismCache) Get(key string, val interface{}) (error, bool) {
	return rc.slave.Get(key, val)
}

// Del delete a cache value.
func (rc *RedismCache) Del(key string) error {
	return rc.master.Del(key)
}

// MSet set multiple cache values.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (rc *RedismCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	return rc.master.MSet(mList, expire, encode...)
}

// MGet get multiple cache values.
func (rc *RedismCache) MGet(keys ...string) (map[string]interface{}, error) {
	return rc.slave.MGet(keys...)
}

// MDel delete multiple cache values.
func (rc *RedismCache) MDel(keys ...string) error {
	return rc.master.MDel(keys...)
}

// Incr atomically increments key by delta. The return value is
// the new value after being incremented or an error.If the key
// didn't exist will create a key and return 1.
func (rc *RedismCache) Incr(key string, delta ...uint64) (int64, error) {
	return rc.master.Incr(key, delta...)
}

// Decr atomically decrements key by delta. The return value is
// the new value after being decremented or an error.If the key
// didn't exist will create a key and return -1.
func (rc *RedismCache) Decr(key string, delta ...uint64) (int64, error) {
	return rc.master.Decr(key, delta...)
}

// IsExist check the key is exists.
func (rc *RedismCache) IsExist(key string) (bool, error) {
	return rc.slave.IsExist(key)
}

// ClearAll delete all values.
func (rc *RedismCache) ClearAll() error {
	return rc.master.ClearAll()
}

// Hset set hashtable value by key and field.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
func (rc *RedismCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	return rc.master.HSet(key, field, val, expire)
}

// HGet return value by key and field.
func (rc *RedismCache) HGet(key string, field string, val interface{}) (error, bool) {
	return rc.slave.HGet(key, field, val)
}

// HDel delete hashtable.
func (rc *RedismCache) HDel(key string, fields ...string) error {
	return rc.master.HDel(key, fields...)
}

// HGetAll return all values by key.
// Require caller to call json.Unmarshal function, if type is struct or map.
func (rc *RedismCache) HGetAll(key string) (map[string]interface{}, error) {
	return rc.slave.HGetAll(key)
}

// HMSet set multiple key-value pairs to hash tables.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
func (rc *RedismCache) HMSet(key string, fields map[string]interface{}, expire int32) error {
	return rc.slave.HMSet(key, fields, expire)
}

// HMGet return all values by key and fields.
// Require caller to call json.Unmarshal function, if type is struct or map.
func (rc *RedismCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	return rc.slave.HMGet(key, fields...)
}

// HVals return all field and value by key.
func (rc *RedismCache) HVals(key string) ([]interface{}, error) {
	return rc.slave.HVals(key)
}

// HIncr atomically increments key by delta.
// The delta default is 1.
func (rc *RedismCache) HIncr(key, fields string, delta ...uint64) (int64, error) {
	return rc.master.HIncr(key, fields, delta...)
}

// HDecr atomically decrements key by delta.
// The delta default is 1.
func (rc *RedismCache) HDecr(key, fields string, delta ...uint64) (int64, error) {
	return rc.master.HDecr(key, fields, delta...)
}

// ZSet set one or more member elements and their score values to an ordered set.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// The data is paired, with score (float64) in front and value in back.
func (rc *RedismCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	return rc.master.ZSet(key, expire, val...)
}

// ZGet return values from ordered set.
// start: Start offset, 0 for the first, - 1 for the last, - 2 for the penultimate.
// stop: Stop offset, 0 for the first, - 1 for the last, - 2 for the penultimate.
// withScores: return score or not.
// isRev: true-decrement (using the ZREVRANGE), false-increment (using the ZRANGE).
func (rc *RedismCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	return rc.slave.ZGet(key, start, stop, withScores, isRev)
}

// ZDel delete ordered set data by key and field.
func (rc *RedismCache) ZDel(key string, field ...string) (int64, error) {
	return rc.master.ZDel(key, field...)
}

// ZRemRangeByRank delete ordered set data within a specified ranking interval.
func (rc *RedismCache) ZRemRangeByRank(key string, start, end int64) (int64, error) {
	return rc.master.ZRemRangeByRank(key, start, end)
}

// ZRemRangeByScore deletes the ordered set data within the specified score interval.
func (rc *RedismCache) ZRemRangeByScore(key string, start, end string) (int64, error) {
	return rc.master.ZRemRangeByScore(key, start, end)
}

// ZRemRangeByLex Delete the ordered set data in the specified variable interval.
// Delete all elements between min and max, if all members have the same score.
func (rc *RedismCache) ZRemRangeByLex(key string, start, end string) (int64, error) {
	return rc.master.ZRemRangeByLex(key, start, end)
}

// ZCard returns the cardinality by key.
func (rc *RedismCache) ZCard(key string) (int64, error) {
	return rc.slave.ZCard(key)
}

// SetBit 设置或清除指定偏移量上的位(bit)
//   参数
//     key:    位图key值
//     offset: 位图偏移量
//     value:  位图值，取值：0或1
//     expire: 失效时长，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     指定偏移量原来储存的位、错误信息
func (rc *RedismCache) SetBit(key string, offset int64, value int, expire int32) (int64, error) {
	return rc.master.SetBit(key, offset, value, expire)
}

// GetBit 获取指定偏移量上的位(bit)
//   参数
//     key:    位图key值
//     offset: 位图偏移量
//   返回
//     字符串值指定偏移量上的位(bit)、错误信息
func (rc *RedismCache) GetBit(key string, offset int64) (int64, error) {
	return rc.slave.GetBit(key, offset)
}

// BitCount 计算给定字符串中被设置为 1 的比特位的数量
//   参数
//     key:      位图key值
//     bitCount: 指定额外的 start 或 end 参数，统计只在特定的位上进行，为nil时统计所有的
//   返回
//     给定字符串中被设置为 1 的比特位的数量、错误信息
func (rc *RedismCache) BitCount(key string, bitCount *cache.BitCount) (int64, error) {
	return rc.slave.BitCount(key, bitCount)
}

// Pipeline call pipeline command.
// Eg:
//   pipe := rc.Pipeline(false).Pipe
//   incr := pipe.Incr("pipeline_counter")
//   pipe.Expire("pipeline_counter", time.Hour)
//   _, err := pipe.Exec()
//   fmt.Println(incr.Val(), err)
func (rc *RedismCache) Pipeline(isTx bool) cache.Pipeliner {
	return rc.master.Pipeline(isTx)
}

// RedisPool Redis缓存
type RedisPool struct {
	client *redis.Client // 连接池

	addr         string        // Host and port, eg: 127.0.0.1:1900,127.0.0.2:1900.
	auth         string        // Auth password.
	dbNum        int           // DB number, default is 0.
	dialTimeout  time.Duration // Connection timeout, in seconds, default 5 seconds.
	readTimeout  time.Duration // Read timeout, in seconds, -1-No timeout，0-Use the default 3 seconds.
	writeTimeout time.Duration // Write timeout, in seconds, default use readTimeout.
	poolSize     int           // Number of connections per node connection pool, default is ten times the number of CPUs.
	minIdleConns int           // Minimum number of idle connections, default is 0.
	maxConnAge   time.Duration // Maximum connection time, in seconds, default is 0.
	poolTimeout  time.Duration // Waiting time when all connections are busy, default is readTimeout+1.
	idleTimeout  time.Duration // Maximum idle time, in seconds, default is 5 minutes.

	prefix    string // Key prefix.
	encodeKey []byte // Encode key, length of 16 times, using Aes encryption.
}

// NewRedisPool return RedisPool object.
func NewRedisPool(addr, auth string, dbNum int, dialTimeout, readTimeout, writeTimeout time.Duration, poolSize, minIdleConns int, maxConnAge, poolTimeout, idleTimeout time.Duration, prefix string, encodeKey []byte) *RedisPool {
	rp := &RedisPool{
		addr:         addr,
		auth:         auth,
		dbNum:        dbNum,
		dialTimeout:  dialTimeout,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		poolSize:     poolSize,
		minIdleConns: minIdleConns,
		maxConnAge:   maxConnAge,
		poolTimeout:  poolTimeout,
		idleTimeout:  idleTimeout,
		prefix:       prefix,
		encodeKey:    encodeKey,
	}

	rp.connect()

	return rp
}

// connect create connect.
func (rp *RedisPool) connect() {
	rp.client = redis.NewClient(&redis.Options{
		Addr:         rp.addr,
		Password:     rp.auth,
		DB:           rp.dbNum,
		DialTimeout:  rp.dialTimeout * time.Second,
		ReadTimeout:  rp.readTimeout * time.Second,
		WriteTimeout: rp.writeTimeout * time.Second,
		PoolSize:     rp.poolSize,
		MinIdleConns: rp.minIdleConns,
		MaxConnAge:   rp.maxConnAge * time.Second,
		PoolTimeout:  rp.poolTimeout * time.Second,
		IdleTimeout:  rp.idleTimeout * time.Second,
	})

	return
}

// Set set a cache value.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (rp *RedisPool) Set(key string, val interface{}, expire int32, encode ...bool) error {
	// conversion type
	data, err := cache.InterToByte(val)
	if err != nil {
		return err
	}

	// encode
	encode = append(encode, false)
	if encode[0] {
		data, err = cache.Encode(data, rp.encodeKey)
		if err != nil {
			return err
		}
	}

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.Set(key, data, time.Duration(expire)*time.Second).Err()
}

// Get get a cache value.
func (rp *RedisPool) Get(key string, val interface{}) (error, bool) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	v, err := rp.client.Get(key).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return nil, false
		}
		return err, false
	}

	// decode
	data, err := cache.Decode([]byte(v), rp.encodeKey)
	if err != nil {
		return err, true
	}

	// conversion type
	err = cache.ByteToInter(data, val)
	if err != nil {
		return err, true
	}

	return nil, true
}

// Del delete a cache value.
func (rp *RedisPool) Del(key string) error {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.Del(key).Err()
}

// MSet set multiple cache values.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (rp *RedisPool) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	var v []interface{}
	for key, val := range mList {
		// conversion type
		data, err := cache.InterToByte(val)
		if err != nil {
			return err
		}

		// encode
		encode = append(encode, false)
		if encode[0] {
			data, err = cache.Encode(data, rp.encodeKey)
			if err != nil {
				return err
			}
		}

		if rp.prefix != "" {
			key = rp.prefix + key
		}
		v = append(v, key, data)
	}

	err := rp.client.MSet(v...).Err()
	if err != nil {
		return err
	}

	// set expire
	if expire > 0 {
		for key := range mList {
			if rp.prefix != "" {
				key = rp.prefix + key
			}
			rp.client.Expire(key, time.Duration(expire)*time.Second)
		}
	}

	return err
}

// MGet get multiple cache values.
func (rp *RedisPool) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})
	args := []string{}
	for _, k := range keys {
		if rp.prefix != "" {
			k = rp.prefix + k
		}
		args = append(args, k)
	}

	v, err := rp.client.MGet(args...).Result()
	if err != nil {
		return mList, err
	}

	i := 0
	for _, val := range v {
		if val == nil {
			// not exist
			mList[keys[i]] = nil
		} else {
			// decode
			data, err := cache.Decode([]byte(val.(string)), rp.encodeKey)
			if err != nil {
				return mList, err
			}

			mList[keys[i]] = string(data)
		}

		i++
	}

	return mList, nil
}

// MDel delete multiple cache values.
func (rp *RedisPool) MDel(keys ...string) error {
	args := make([]string, len(keys))
	for k, v := range keys {
		if rp.prefix != "" {
			v = rp.prefix + v
		}
		args[k] = v
	}

	rp.client.Del(args...)
	return nil
}

// Incr atomically increments key by delta. The return value is
// the new value after being incremented or an error.If the key
// didn't exist will create a key and return 1.
func (rp *RedisPool) Incr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	v, err := rp.client.IncrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// Decr atomically decrements key by delta. The return value is
// the new value after being decremented or an error.If the key
// didn't exist will create a key and return -1.
func (rp *RedisPool) Decr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	v, err := rp.client.DecrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// IsExist check the key is exists.
func (rp *RedisPool) IsExist(key string) (bool, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	n, err := rp.client.Exists(key).Result()
	if err != nil {
		return false, err
	}

	if n == 1 {
		return true, nil
	}

	return false, nil
}

// ClearAll delete all values.
func (rp *RedisPool) ClearAll() error {
	keys, err := rp.client.Keys("*").Result()
	if err != nil {
		return err
	}

	return rp.client.Del(keys...).Err()
}

// Hset set hashtable value by key and field.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
func (rp *RedisPool) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	// conversion type
	data, err := cache.InterToByte(val)
	if err != nil {
		return -1, err
	}

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	err = rp.client.HSet(key, field, data).Err()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		rp.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return 1, err
}

// HGet return value by key and field.
func (rp *RedisPool) HGet(key string, field string, val interface{}) (error, bool) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	v, err := rp.client.HGet(key, field).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return nil, false
		}
		return err, false
	}

	// conversion type
	err = cache.ByteToInter([]byte(v), val)
	if err != nil {
		return err, true
	}

	return nil, true
}

// HDel delete hashtable.
func (rp *RedisPool) HDel(key string, fields ...string) error {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.HDel(key, fields...).Err()
}

// HGetAll return all values by key.
// Require caller to call json.Unmarshal function, if type is struct or map.
func (rp *RedisPool) HGetAll(key string) (map[string]interface{}, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	res := make(map[string]interface{})
	val, err := rp.client.HGetAll(key).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return res, nil
		}

		return nil, err
	}

	for k, v := range val {
		res[k] = v
	}

	return res, err
}

// HMSet set multiple key-value pairs to hash tables.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
func (rp *RedisPool) HMSet(key string, fields map[string]interface{}, expire int32) error {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	err := rp.client.HMSet(key, fields).Err()
	if err != nil {
		return err
	}

	if expire > 0 {
		rp.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return nil
}

// HMGet return all values by key and fields.
// Require caller to call json.Unmarshal function, if type is struct or map.
func (rp *RedisPool) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	res := make(map[string]interface{})

	v, err := rp.client.HMGet(key, fields...).Result()
	if err != nil {
		return nil, err
	}

	if v == nil {
		return res, nil
	}

	i := 0
	for _, field := range fields {
		res[field] = v[i]
		i++
	}

	return res, err
}

// HVals return all field and value by key.
func (rp *RedisPool) HVals(key string) ([]interface{}, error) {
	vals, err := rp.client.HVals(key).Result()
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(vals))
	for k, v := range vals {
		res[k] = v
	}

	return res, nil
}

// HIncr atomically increments key by delta.
// The delta default is 1.
func (rp *RedisPool) HIncr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.HIncrBy(key, fields, int64(delta[0])).Result()
}

// HDecr atomically decrements key by delta.
// The delta default is 1.
func (rp *RedisPool) HDecr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.HIncrBy(key, fields, 0-int64(delta[0])).Result()
}

// ZSet set one or more member elements and their score values to an ordered set.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// The data is paired, with score (float64) in front and value in back.
func (rp *RedisPool) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	valLen := len(val)
	if valLen < 2 || valLen%2 != 0 {
		return -1, errors.New("val param error")
	}
	vals := []redis.Z{}
	for i := 0; i < valLen-1; i += 2 {
		stZ := redis.Z{
			Score:  val[i].(float64),
			Member: val[i+1],
		}
		vals = append(vals, stZ)
	}

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	n, err := rp.client.ZAdd(key, vals...).Result()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		rp.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return n, err
}

// ZGet return values from ordered set.
// start: Start offset, 0 for the first, - 1 for the last, - 2 for the penultimate.
// stop: Stop offset, 0 for the first, - 1 for the last, - 2 for the penultimate.
// withScores: return score or not.
// isRev: true-decrement (using the ZREVRANGE), false-increment (using the ZRANGE).
func (rp *RedisPool) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	var err error
	vals := []redis.Z{}
	res := []string{}

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	if isRev {

		if withScores {
			vals, err = rp.client.ZRevRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return rp.client.ZRevRange(key, int64(start), int64(stop)).Result()
		}
	} else {
		if withScores {
			vals, err = rp.client.ZRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return rp.client.ZRange(key, int64(start), int64(stop)).Result()
		}
	}

	for _, val := range vals {
		res = append(res, fmt.Sprintf("%f", val.Score), fmt.Sprintf("%v", val.Member))
	}

	return res, err
}

// ZDel delete ordered set data by key and field.
func (rp *RedisPool) ZDel(key string, field ...string) (int64, error) {
	var args []interface{}
	for _, f := range field {
		args = append(args, f)
	}

	if rp.prefix != "" {
		key = rp.prefix + key
	}
	return rp.client.ZRem(key, args...).Result()
}

// ZRemRangeByRank delete ordered set data within a specified ranking interval.
func (rp *RedisPool) ZRemRangeByRank(key string, start, end int64) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.ZRemRangeByRank(key, start, end).Result()
}

// ZRemRangeByScore deletes the ordered set data within the specified score interval.
func (rp *RedisPool) ZRemRangeByScore(key string, start, end string) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.ZRemRangeByScore(key, start, end).Result()
}

// ZRemRangeByLex Delete the ordered set data in the specified variable interval.
// Delete all elements between min and max, if all members have the same score.
func (rp *RedisPool) ZRemRangeByLex(key string, start, end string) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.ZRemRangeByLex(key, start, end).Result()
}

// ZCard returns the cardinality by key.
func (rp *RedisPool) ZCard(key string) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	return rp.client.ZCard(key).Result()
}

// SetBit 设置或清除指定偏移量上的位(bit)
//   参数
//     key:    位图key值
//     offset: 位图偏移量
//     value:  位图值，取值：0或1
//     expire: 失效时长，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     指定偏移量原来储存的位、错误信息
func (rp *RedisPool) SetBit(key string, offset int64, value int, expire int32) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	res, err := rp.client.SetBit(key, offset, value).Result()
	if err != nil {
		return res, err
	}

	if expire > 0 {
		rp.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return res, err
}

// GetBit 获取指定偏移量上的位(bit)
//   参数
//     key:    位图key值
//     offset: 位图偏移量
//   返回
//     字符串值指定偏移量上的位(bit)、错误信息
func (rp *RedisPool) GetBit(key string, offset int64) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.GetBit(key, offset).Result()
}

// BitCount 计算给定字符串中被设置为 1 的比特位的数量
//   参数
//     key:      位图key值
//     bitCount: 指定额外的 start 或 end 参数，统计只在特定的位上进行，为nil时统计所有的
//   返回
//     给定字符串中被设置为 1 的比特位的数量、错误信息
func (rp *RedisPool) BitCount(key string, bitCount *cache.BitCount) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	if bitCount != nil {
		bc := redis.BitCount{Start: bitCount.Start, End: bitCount.End}
		return rp.client.BitCount(key, &bc).Result()
	}

	return rp.client.BitCount(key, nil).Result()
}

// Pipeline call pipeline command.
// Eg:
//   pipe := rc.Pipeline(false).Pipe
//   incr := pipe.Incr("pipeline_counter")
//   pipe.Expire("pipeline_counter", time.Hour)
//   _, err := pipe.Exec()
//   fmt.Println(incr.Val(), err)
func (rp *RedisPool) Pipeline(isTx bool) cache.Pipeliner {
	p := cache.Pipeliner{}
	if isTx {
		p.Pipe = rp.client.TxPipeline()
	} else {
		p.Pipe = rp.client.Pipeline()
	}

	return p
}
