// redis cluster adapter
package redisc

import (
	"github.com/lixy529/gotools/cache"
	"github.com/go-redis/redis"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"strings"
	"errors"
)

const NOT_EXIST = "redis: nil"

// RediscCache redis cluster cache
type RediscCache struct {
	client *redis.ClusterClient

	addr         string        // Host and port, eg: 127.0.0.1:1900,127.0.0.2:1900.
	auth         string        // Auth password.
	dialTimeout  time.Duration // Connection timeout, in seconds, default 5 seconds.
	readTimeout  time.Duration // Read timeout, in seconds, -1-No timeout，0-Use the default 3 seconds.
	writeTimeout time.Duration // Write timeout, in seconds, default use readTimeout.
	poolSize     int           // Number of connections per node connection pool, default is five times the number of CPUs.
	minIdleConns int           // Minimum number of idle connections, default is 0.
	maxConnAge   time.Duration // Maximum connection time, in seconds, default is 0.
	poolTimeout  time.Duration // Waiting time when all connections are busy, default is readTimeout+1.
	idleTimeout  time.Duration // Maximum idle time, in seconds, default is 5 minutes.

	prefix    string // Key prefix.
	encodeKey []byte // Encode key, length of 16 times, using Aes encryption.
}

// NewRediscCache return RediscCache object.
func NewRediscCache() cache.Cache {
	return &RediscCache{}
}

// Init initialization configuration.
// Eg:
// {
// "addr":"127.0.0.1:19100,127.0.0.2:19100,127.0.0.3:19100",
// "auth":"xxxx",
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
func (c *RediscCache) Init(config string) error {
	var mapCfg map[string]string
	var err error

	err = json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("RediscCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	// Host and port
	c.addr = mapCfg["addr"]

	// Auth password
	c.auth = mapCfg["auth"]

	// Connection timeout
	dialTimeout, err := strconv.Atoi(mapCfg["dialTimeout"])
	if err != nil || dialTimeout < 0 {
		c.dialTimeout = 5
	} else {
		c.dialTimeout = time.Duration(dialTimeout)
	}

	// Read timeout
	readTimeout, err := strconv.Atoi(mapCfg["readTimeout"])
	if err != nil {
		c.readTimeout = 3
	} else if readTimeout < 0 {
		c.readTimeout = -1
	} else {
		c.readTimeout = time.Duration(readTimeout)
	}

	// Write timeout
	writeTimeout, err := strconv.Atoi(mapCfg["writeTimeout"])
	if err != nil {
		c.writeTimeout = c.readTimeout
	} else if writeTimeout < 0 {
		c.writeTimeout = -1
	} else {
		c.writeTimeout = time.Duration(writeTimeout)
	}

	// Number of connections per node connection pool
	poolSize, err := strconv.Atoi(mapCfg["poolSize"])
	if err != nil || poolSize < 0 {
		c.poolSize = 0
	} else {
		c.poolSize = poolSize
	}

	// Minimum number of idle connections
	minIdleConns, err := strconv.Atoi(mapCfg["minIdleConns"])
	if err != nil || minIdleConns < 0 {
		c.minIdleConns = 0
	} else {
		c.minIdleConns = minIdleConns
	}

	// Maximum connection time
	maxConnAge, err := strconv.Atoi(mapCfg["maxConnAge"])
	if err != nil || maxConnAge < 0 {
		c.maxConnAge = 0
	} else {
		c.maxConnAge = time.Duration(maxConnAge)
	}

	// Wait timeout
	poolTimeout, err := strconv.Atoi(mapCfg["poolTimeout"])
	if err != nil || poolTimeout < 0 {
		c.poolTimeout = c.readTimeout + 1
	} else {
		c.poolTimeout = time.Duration(poolTimeout)
	}

	// Maximum idle time
	idleTimeout, err := strconv.Atoi(mapCfg["idleTimeout"])
	if err != nil || idleTimeout < 0 {
		c.idleTimeout = 300
	} else {
		c.idleTimeout = time.Duration(idleTimeout)
	}

	// Key prefix
	if prefix, ok := mapCfg["prefix"]; ok {
		c.prefix = prefix
	}

	// Encode key
	if tmp, ok := mapCfg["encodeKey"]; ok && tmp != "" {
		c.encodeKey = []byte(tmp)
	}

	// connect
	c.client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        strings.Split(c.addr, ","),
		Password:     c.auth,
		DialTimeout:  c.dialTimeout * time.Second,
		ReadTimeout:  c.readTimeout * time.Second,
		WriteTimeout: c.writeTimeout * time.Second,
		PoolSize:     c.poolSize,
		MinIdleConns: c.minIdleConns,
		MaxConnAge:   c.maxConnAge * time.Second,
		PoolTimeout:  c.poolTimeout * time.Second,
		IdleTimeout:  c.idleTimeout * time.Second,
	})

	return nil
}

// Set set a cache value.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (c *RediscCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
	// conversion type
	data, err := cache.InterToByte(val)
	if err != nil {
		return err
	}

	// encode
	encode = append(encode, false)
	if encode[0] {
		data, err = cache.Encode(data, c.encodeKey)
		if err != nil {
			return err
		}
	}

	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.Set(key, data, time.Duration(expire)*time.Second).Err()
}

// Get get a cache value.
func (c *RediscCache) Get(key string, val interface{}) (error, bool) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	v, err := c.client.Get(key).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return nil, false
		}
		return err, false
	}

	// decode
	data, err := cache.Decode([]byte(v), c.encodeKey)
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
func (c *RediscCache) Del(key string) error {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.Del(key).Err()
}

// MSet set multiple cache values.
// If use c.client.MSet function, will report "CROSSSLOT Keys in request don't hash to the same slot" error.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (c *RediscCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	for k, v := range mList {
		err := c.Set(k, v, expire, encode...)
		if err != nil {
			return err
		}
	}

	return nil
}

// MGet get multiple cache values.
// If use c.client.MGet function, will report "CROSSSLOT Keys in request don't hash to the same slot" error.
func (c *RediscCache) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})
	for _, k := range keys {
		v := ""
		err, b := c.Get(k, &v)
		if err != nil {
			return mList, err
		} else if !b {
			v = ""
		}
		mList[k] = v
	}

	return mList, nil
}

// MDel delete multiple cache values.
// If use c.client.Del function, will report "CROSSSLOT Keys in request don't hash to the same slot" error.
func (c *RediscCache) MDel(keys ...string) error {
	for _, key := range keys {
		err := c.Del(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Incr atomically increments key by delta. The return value is
// the new value after being incremented or an error.If the key
// didn't exist will create a key and return 1.
func (c *RediscCache) Incr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}
	v, err := c.client.IncrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// Decr atomically decrements key by delta. The return value is
// the new value after being decremented or an error.If the key
// didn't exist will create a key and return -1.
func (c *RediscCache) Decr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}
	v, err := c.client.DecrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// IsExist check the key is exists.
func (c *RediscCache) IsExist(key string) (bool, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	n, err := c.client.Exists(key).Result()
	if err != nil {
		return false, err
	}

	if n == 1 {
		return true, nil
	}

	return false, nil
}

// ClearAll delete all values.
func (c *RediscCache) ClearAll() error {
	keys, err := c.client.Keys("*").Result()
	if err != nil {
		return err
	}

	return c.client.Del(keys...).Err()
}

// Hset set hashtable value by key and field.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
func (c *RediscCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	// conversion type
	data, err := cache.InterToByte(val)
	if err != nil {
		return -1, err
	}

	if c.prefix != "" {
		key = c.prefix + key
	}

	err = c.client.HSet(key, field, data).Err()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		c.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return 1, err
}

// HGet return value by key and field.
func (c *RediscCache) HGet(key string, field string, val interface{}) (error, bool) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	v, err := c.client.HGet(key, field).Result()
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
func (c *RediscCache) HDel(key string, fields ...string) error {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.HDel(key, fields...).Err()
}

// HGetAll return all values by key.
// Require caller to call json.Unmarshal function, if type is struct or map.
func (c *RediscCache) HGetAll(key string) (map[string]interface{}, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	res := make(map[string]interface{})
	val, err := c.client.HGetAll(key).Result()
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
func (c *RediscCache) HMSet(key string, fields map[string]interface{}, expire int32) error {
	if c.prefix != "" {
		key = c.prefix + key
	}

	err := c.client.HMSet(key, fields).Err()
	if err != nil {
		return err
	}

	if expire > 0 {
		c.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return nil
}

// HMGet return all values by key and fields.
// Require caller to call json.Unmarshal function, if type is struct or map.
func (c *RediscCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	res := make(map[string]interface{})

	v, err := c.client.HMGet(key, fields...).Result()
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
func (c *RediscCache) HVals(key string) ([]interface{}, error) {
	vals, err := c.client.HVals(key).Result()
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
func (c *RediscCache) HIncr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.HIncrBy(key, fields, int64(delta[0])).Result()
}

// HIncr atomically decrements key by delta.
// The delta default is 1.
func (c *RediscCache) HDecr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.HIncrBy(key, fields, 0-int64(delta[0])).Result()
}

// ZSet set one or more member elements and their score values to an ordered set.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// The data is paired, with score (float64) in front and value in back.
func (c *RediscCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
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

	if c.prefix != "" {
		key = c.prefix + key
	}

	n, err := c.client.ZAdd(key, vals...).Result()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		c.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return n, err
}

// ZGet return values from ordered set.
// start: Start offset, 0 for the first, - 1 for the last, - 2 for the penultimate.
// stop: Stop offset, 0 for the first, - 1 for the last, - 2 for the penultimate.
// withScores: return score or not.
// isRev: true-decrement (using the ZREVRANGE), false-increment (using the ZRANGE).
func (c *RediscCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	var err error
	vals := []redis.Z{}
	res := []string{}

	if c.prefix != "" {
		key = c.prefix + key
	}

	if isRev {

		if withScores {
			vals, err = c.client.ZRevRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return c.client.ZRevRange(key, int64(start), int64(stop)).Result()
		}
	} else {
		if withScores {
			vals, err = c.client.ZRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return c.client.ZRange(key, int64(start), int64(stop)).Result()
		}
	}

	for _, val := range vals {
		res = append(res, fmt.Sprintf("%f", val.Score), fmt.Sprintf("%v", val.Member))
	}

	return res, err
}

// ZDel delete ordered set data by key and field.
func (c *RediscCache) ZDel(key string, field ...string) (int64, error) {
	var args []interface{}
	for _, f := range field {
		args = append(args, f)
	}

	if c.prefix != "" {
		key = c.prefix + key
	}
	return c.client.ZRem(key, args...).Result()
}

// ZRemRangeByRank delete ordered set data within a specified ranking interval.
func (c *RediscCache) ZRemRangeByRank(key string, start, end int64) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.ZRemRangeByRank(key, start, end).Result()
}

// ZRemRangeByScore deletes the ordered set data within the specified score interval.
func (c *RediscCache) ZRemRangeByScore(key string, start, end string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.ZRemRangeByScore(key, start, end).Result()
}

// ZRemRangeByLex Delete the ordered set data in the specified variable interval.
// Delete all elements between min and max, if all members have the same score.
func (c *RediscCache) ZRemRangeByLex(key string, start, end string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.ZRemRangeByLex(key, start, end).Result()
}

// ZCard returns the cardinality by key.
func (c *RediscCache) ZCard(key string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	return c.client.ZCard(key).Result()
}

// SetBit 设置或清除指定偏移量上的位(bit)
//   参数
//     key:    位图key值
//     offset: 位图偏移量
//     value:  位图值，取值：0或1
//     expire: 失效时长，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     指定偏移量原来储存的位、错误信息
func (c *RediscCache) SetBit(key string, offset int64, value int, expire int32) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	res, err := c.client.SetBit(key, offset, value).Result()
	if err != nil {
		return res, err
	}

	if expire > 0 {
		c.client.Expire(key, time.Duration(expire)*time.Second)
	}

	return res, err
}

// GetBit 获取指定偏移量上的位(bit)
//   参数
//     key:    位图key值
//     offset: 位图偏移量
//   返回
//     字符串值指定偏移量上的位(bit)、错误信息
func (c *RediscCache) GetBit(key string, offset int64) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.GetBit(key, offset).Result()
}

// BitCount 计算给定字符串中被设置为 1 的比特位的数量
//   参数
//     key:      位图key值
//     bitCount: 指定额外的 start 或 end 参数，统计只在特定的位上进行，为nil时统计所有的
//   返回
//     给定字符串中被设置为 1 的比特位的数量、错误信息
func (c *RediscCache) BitCount(key string, bitCount *cache.BitCount) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	if bitCount != nil {
		bc := redis.BitCount{Start: bitCount.Start, End: bitCount.End}
		return c.client.BitCount(key, &bc).Result()
	}

	return c.client.BitCount(key, nil).Result()
}

// Pipeline call pipeline command.
// Eg:
//   pipe := rc.Pipeline(false).Pipe
//   incr := pipe.Incr("pipeline_counter")
//   pipe.Expire("pipeline_counter", time.Hour)
//   _, err := pipe.Exec()
//   fmt.Println(incr.Val(), err)
func (c *RediscCache) Pipeline(isTx bool) cache.Pipeliner {
	p := cache.Pipeliner{}
	if isTx {
		p.Pipe = c.client.TxPipeline()
	} else {
		p.Pipe = c.client.Pipeline()
	}

	return p
}
