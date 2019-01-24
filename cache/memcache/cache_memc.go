package memcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/lixy529/gotools/cache"
	"github.com/lixy529/gotools/utils"
	"strconv"
	"strings"
	"time"
)

const (
	NOT_EXIST               = "cache miss"
	COMPRESSION_ZLIB        = "zlib"
	FLAGES_INT_UNCOMPRESS   = 1
	FLAGES_FLOAT_UNCOMPRESS = 2
	FLAGES_JSON_UNCOMPRESS  = 6
	FLAGES_JSON_COMPRESS    = 54
	FLAGES_STR_UNCOMPRESS   = 0
	FLAGES_STR_COMPRESS     = 48
)

// MemcCache memcache cache
type MemcCache struct {
	conn      *memcache.Client
	connCfg   []string
	maxIdle   int           // Maximum number of idle connections, default is 2, If the value is less than 1, use the default value.
	ioTimeOut time.Duration // IO timeout, default is 100 milliseconds, less than or equal to 0 use the default time, the unit is milliseconds.
	prefix    string        // Key prefix.
	encodeKey []byte        // Encode key, length of 16 times, using Aes encryption

	serializer        string // Serialization, only supports json.
	compressType      string // Compression type, only supports zlib.
	compressThreshold int    // Compression over size.
}

// NewMemcCache return MemcCache object.
func NewMemcCache() cache.Cache {
	return &MemcCache{}
}

// connect connect memcache.
func (mc *MemcCache) connect() error {
	if mc.conn != nil {
		return nil
	}

	mc.conn = memcache.New(mc.connCfg...)
	if mc.conn == nil {
		return fmt.Errorf("MemcCache Connect memcache [%s] failed", strings.Join(mc.connCfg, ","))
	}

	if mc.maxIdle > 0 {
		mc.conn.MaxIdleConns = mc.maxIdle
	}

	if mc.ioTimeOut >= 0 {
		mc.conn.Timeout = mc.ioTimeOut * time.Millisecond
	}

	return nil
}

// Init initialization configuration.
// Eg:
// {
// "addr":"127.0.0.1:11211,127.0.0.2:11211",
// "maxIdle":"3",
// "ioTimeOut":"1",
// "prefix":"le_",
// "serializer":"json",
// "compressType":"zlib",
// "compressThreshold":"256",
// "encodeKey":"abcdefghij123456",
// }
func (mc *MemcCache) Init(config string) error {
	var mapCfg map[string]string
	var ok bool
	err := json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("MemcCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	if _, ok = mapCfg["addr"]; !ok {
		return errors.New("MemcCache: Config hasn't address.")
	}

	mc.connCfg = strings.Split(mapCfg["addr"], ",")
	if _, ok = mapCfg["maxIdle"]; ok {
		mc.maxIdle, _ = strconv.Atoi(mapCfg["maxIdle"])
	} else {
		mc.maxIdle = -1
	}
	if _, ok = mapCfg["ioTimeOut"]; ok {
		ioTimeOut, _ := strconv.Atoi(mapCfg["ioTimeOut"])
		mc.ioTimeOut = time.Duration(ioTimeOut)
	} else {
		mc.ioTimeOut = -1
	}
	if prefix, ok := mapCfg["prefix"]; ok {
		mc.prefix = prefix
	}

	mc.serializer = "json"

	// Compression type, only supports zlib.
	mc.compressType, _ = mapCfg["compressType"]
	if mc.compressType != "" {
		if mc.compressType != COMPRESSION_ZLIB {
			return fmt.Errorf("MemcCache: Compress type don't support %s", mc.compressType)
		}

		if _, ok = mapCfg["compressThreshold"]; ok {
			var err error
			mc.compressThreshold, err = strconv.Atoi(mapCfg["compressThreshold"])
			if err != nil {
				return fmt.Errorf("MemcCache: Compress threshold error, %s", err.Error())
			}
		}

	}

	// Encode key
	if tmp, ok := mapCfg["encodeKey"]; ok {
		mc.encodeKey = []byte(tmp)
	}

	err = mc.connect()
	if err != nil {
		return err
	}

	return nil
}

// Set set a cache value.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (mc *MemcCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
	if err := mc.connect(); err != nil {
		return err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}

	// Use Unix epoch time if expire more than 30 days.
	if expire > 86400*30 {
		expire = int32(time.Now().Unix()) + expire
	}
	item := memcache.Item{Key: key, Expiration: expire}

	// conversion type
	data, err := cache.InterToByte(val)
	if err != nil {
		return err
	}

	// encode
	encode = append(encode, false)
	if encode[0] {
		data, err = cache.Encode(data, mc.encodeKey)
		if err != nil {
			return err
		}
	}

	// set flags
	flags := FLAGES_JSON_UNCOMPRESS
	valType := ""
	if _, ok := val.(string); ok {
		valType = "string"
		flags = FLAGES_STR_UNCOMPRESS
	} else if _, ok := val.(float64); ok {
		valType = "float"
		flags = FLAGES_FLOAT_UNCOMPRESS
	} else if _, ok := val.(int64); ok {
		valType = "int"
		flags = FLAGES_INT_UNCOMPRESS
	} else if _, ok := val.(int32); ok {
		valType = "int"
		flags = FLAGES_INT_UNCOMPRESS
	} else if _, ok := val.(int); ok {
		valType = "int"
		flags = FLAGES_INT_UNCOMPRESS
	}

	// compress, only supports zlib
	if mc.compressType == COMPRESSION_ZLIB {
		dataLen := len(data)
		if dataLen > mc.compressThreshold {
			if valType == "string" {
				flags = FLAGES_STR_COMPRESS
			} else {
				flags = FLAGES_JSON_COMPRESS
			}
			data, err = utils.ZlibEncode(data)
			if err != nil {
				return err
			}
			data = []byte(string(utils.Int32ToByte(int32(dataLen), false)) + string(data))
		}
	}

	item.Flags = uint32(flags)
	item.Value = data
	return mc.conn.Set(&item)
}

// Get get a cache value.
func (mc *MemcCache) Get(key string, val interface{}) (error, bool) {
	if err := mc.connect(); err != nil {
		return err, false
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	item, err := mc.conn.Get(key)
	if err != nil {
		if strings.Contains(err.Error(), NOT_EXIST) {
			return nil, false
		}
		return err, false
	}

	// uncompress
	data := item.Value
	if item.Flags == FLAGES_JSON_COMPRESS || item.Flags == FLAGES_STR_COMPRESS {
		data, err = utils.ZlibDecode(data[4:])
		if err != nil {
			return err, true
		}
	}

	// decode
	data, err = cache.Decode(data, mc.encodeKey)
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
func (mc *MemcCache) Del(key string) error {
	if err := mc.connect(); err != nil {
		return err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}

	err := mc.conn.Delete(key)
	if err == nil || strings.Contains(err.Error(), NOT_EXIST) {
		return nil
	}

	return err
}

// MSet set multiple cache values.
// The expiration time is in seconds, is relative time from now , zero means the Item has no expiration time.
// Value will be encrypt if encode is true.
func (mc *MemcCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	for key, val := range mList {
		err := mc.Set(key, val, expire, encode...)
		if err != nil {
			return err
		}
	}
	return nil
}

// MSet get multiple cache values.
func (mc *MemcCache) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})

	if err := mc.connect(); err != nil {
		return mList, err
	}

	if mc.prefix != "" {
		for k, v := range keys {
			keys[k] = mc.prefix + v
		}
	}

	mv, err := mc.conn.GetMulti(keys)
	if err != nil {
		return mList, err
	}

	for key, val := range mv {
		// uncompress
		data := val.Value
		if val.Flags == FLAGES_JSON_COMPRESS || val.Flags == FLAGES_STR_COMPRESS {
			data, err = utils.ZlibDecode(data[4:])
			if err != nil {
				mList[key] = nil
				continue
			}
		}

		// decode
		data, err = cache.Decode(data, mc.encodeKey)
		if err != nil {
			return mList, err
		}

		if mc.prefix != "" {
			key = key[len(mc.prefix):]
		}
		mList[key] = data
	}

	return mList, nil
}

// MDel delete multiple cache values.
func (mc *MemcCache) MDel(keys ...string) error {
	for _, key := range keys {
		err := mc.Del(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Incr atomically increments key by delta. The return value is
// the new value after being incremented or an error. If the value
// didn't exist in memcached the error is ErrCacheMiss. The value in
// memcached must be an decimal number, or an error will be returned.
// On 64-bit overflow, the new value wraps around.
func (mc *MemcCache) Incr(key string, delta ...uint64) (int64, error) {
	if err := mc.connect(); err != nil {
		return 0, err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	delta = append(delta, 1)
	v, err := mc.conn.Increment(key, delta[0])
	return int64(v), err
}

// Decr atomically decrements key by delta. The return value is
// the new value after being decremented or an error. If the value
// didn't exist in memcached the error is ErrCacheMiss. The value in
// memcached must be an decimal number, or an error will be returned.
// On underflow, the new value is capped at zero and does not wrap
// around.
func (mc *MemcCache) Decr(key string, delta ...uint64) (int64, error) {
	if err := mc.connect(); err != nil {
		return 0, err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	delta = append(delta, 1)
	v, err := mc.conn.Decrement(key, delta[0])
	return int64(v), err
}

// IsExist check the key is exists.
func (mc *MemcCache) IsExist(key string) (bool, error) {
	if err := mc.connect(); err != nil {
		return false, err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	_, err := mc.conn.Get(key)
	if err != nil {
		if strings.Contains(err.Error(), NOT_EXIST) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ClearAll delete all values.
func (mc *MemcCache) ClearAll() error {
	if err := mc.connect(); err != nil {
		return err
	}

	return mc.conn.FlushAll()
}

// Hset set hashtable value by key and field, memcache hasn't hashtable.
func (mc *MemcCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support HSet")
}

// HGet return value by key and field, memcache hasn't hashtable.
func (mc *MemcCache) HGet(key string, field string, val interface{}) (error, bool) {
	return errors.New("MemcCache: Memcache don't support HGet"), false
}

// HDel delete hashtable, memcache hasn't hashtable.
func (mc *MemcCache) HDel(key string, fields ...string) error {
	return errors.New("MemcCache: Memcache don't support HDel")
}

// HGetAll return all values by key, memcache hasn't hashtable.
func (mc *MemcCache) HGetAll(key string) (map[string]interface{}, error) {
	return nil, errors.New("MemcCache: Memcache don't support HGetAll")
}

// HMSet set multiple key-value pairs to hash tables, memcache hasn't hashtable.
func (c *MemcCache) HMSet(key string, fields map[string]interface{}, expire int32) error {
	return errors.New("MemcCache: Memcache don't support HMSet")
}

// HMGet return all values by key and fields., memcache hasn't hashtable.
func (mc *MemcCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	return nil, errors.New("MemcCache: Memcache don't support HMGet")
}

// HVals return all field and value by key., memcache hasn't hashtable.
func (mc *MemcCache) HVals(key string) ([]interface{}, error) {
	return nil, errors.New("MemcCache: Memcache don't support HVals")
}

// HIncr atomically increments key by delta, memcache hasn't hashtable.
func (mc *MemcCache) HIncr(key, fields string, delta ...uint64) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support HIncr")
}

// HDecr atomically decrements key by delta, memcache hasn't hashtable.
func (mc *MemcCache) HDecr(key, fields string, delta ...uint64) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support HDecr")
}

// ZSet add set, memcache hasn't set.
func (mc *MemcCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZSet")
}

// ZGet query set, memcache hasn't set.
func (mc *MemcCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	return nil, errors.New("MemcCache: Memcache don't support ZGet")
}

// ZDel delete set, memcache hasn't set.
func (rc *MemcCache) ZDel(key string, field ...string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZDel")
}

// ZRemRangeByRank delete set, memcache hasn't set.
func (rc *MemcCache) ZRemRangeByRank(key string, start, end int64) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZRemRangeByRank")
}

// ZRemRangeByScore delete set, memcache hasn't set.
func (rc *MemcCache) ZRemRangeByScore(key string, start, end string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZRemRangeByScore")
}

// ZRemRangeByLex delete set, memcache hasn't set.
func (rc *MemcCache) ZRemRangeByLex(key string, start, end string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZRemRangeByLex")
}

// ZCard return set, memcache hasn't set.
func (mc *MemcCache) ZCard(key string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZCard")
}

// Pipeline call pipeline command, memcache hasn't pipeline.
func (mc *MemcCache) Pipeline(isTx bool) cache.Pipeliner {
	return cache.Pipeliner{}
}
