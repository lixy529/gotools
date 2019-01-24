package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lixy529/gotools/utils"
	"github.com/go-redis/redis"
)

const (
	ENCODE_FLAG = "ENC1_"
	ENCODE_LEN  = 5
)

// Cache all cache interface.
type Cache interface {
	Init(config string) error
	Set(key string, val interface{}, expire int32, encode ...bool) error
	Get(key string, val interface{}) (error, bool)
	Del(key string) error

	MSet(mList map[string]interface{}, expire int32, encode ...bool) error
	MGet(keys ...string) (map[string]interface{}, error)
	MDel(keys ...string) error

	Incr(key string, delta ...uint64) (int64, error)
	Decr(key string, delta ...uint64) (int64, error)
	IsExist(key string) (bool, error)
	ClearAll() error

	// Hash table operation(only redis support)
	HSet(key string, field string, val interface{}, expire int32) (int64, error)
	HGet(key string, field string, val interface{}) (error, bool)
	HDel(key string, fields ...string) error
	HGetAll(key string) (map[string]interface{}, error)
	HMSet(key string, fields map[string]interface{}, expire int32) error
	HMGet(key string, fields ...string) (map[string]interface{}, error)
	HVals(key string) ([]interface{}, error)
	HIncr(key, fields string, delta ...uint64) (int64, error)
	HDecr(key, fields string, delta ...uint64) (int64, error)

	// Set operation (only redis support)
	ZSet(key string, expire int32, val ...interface{}) (int64, error)
	ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error)
	ZDel(key string, field ...string) (int64, error)
	ZCard(key string) (int64, error)
	ZRemRangeByRank(key string, start, end int64) (int64, error)
	ZRemRangeByScore(key string, start, end string) (int64, error)
	ZRemRangeByLex(key string, start, end string) (int64, error)

	// pipeline
	Pipeline(isTx bool) Pipeliner
}

// IJson marshal and unmarshal json.
type IJson interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

// Pipeliner redis pipeline.
type Pipeliner struct {
	Pipe redis.Pipeliner
}

var Adapters = make(map[string]Cache)

// GetCache return cache adapter.
// "adapterName" can be empty when there is only one adapter.
func GetCache(adapterName ...string) (Cache, error) {
	if len(adapterName) == 0 {
		for _, adapter := range Adapters {
			return adapter, nil
		}
		return nil, errors.New("Cache: Adapter is empty")
	}

	name := adapterName[0]
	adapter, ok := Adapters[name]
	if !ok {
		return nil, fmt.Errorf("Cache: unknown adapter name %q", name)
	}

	return adapter, nil
}

// Encode encode data.
func Encode(data, key []byte) ([]byte, error) {
	encodeText, err := utils.AesEncode(data, key)
	if err != nil {
		return nil, fmt.Errorf("Cache: Encode error, %s", err.Error())
	}
	encode := []byte(ENCODE_FLAG)
	encode = append(encode, encodeText...)
	return encode, nil
}

// Decode decode data.
func Decode(data, key []byte) ([]byte, error) {
	if len(data) < ENCODE_LEN {
		return data, nil
	}

	encodeFlag := data[:ENCODE_LEN]
	if string(encodeFlag) == ENCODE_FLAG {
		decode, err := utils.AesDecode(data[ENCODE_LEN:], []byte(key))
		if err != nil {
			return nil, fmt.Errorf("Cache: Decode error, %s", err.Error())
		}
		return decode, nil
	}
	return data, nil
}

// InterToByte convert the interface{} to []byte.
func InterToByte(src interface{}) ([]byte, error) {
	var data []byte
	var err error
	if str, ok := src.(string); ok {
		data = []byte(str)
	} else if str, ok := src.(*string); ok {
		data = []byte(*str)
	} else if inter, ok := src.(IJson); ok {
		data, err = inter.MarshalJSON()
		if err != nil {
			return nil, err
		}
	} else {
		data, err = json.Marshal(src)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// ByteToInter convert the []byte to interface{}.
func ByteToInter(src []byte, dst interface{}) error {
	if str, ok := dst.(*string); ok {
		*str = string(src)
		return nil
	} else if inter, ok := dst.(IJson); ok {
		return inter.UnmarshalJSON(src)
	}

	return json.Unmarshal(src, dst)
}
