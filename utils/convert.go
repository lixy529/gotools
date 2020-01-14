package utils

import (
	"bytes"
	"encoding/binary"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// IpItoa IP integer to string.
// eg: "10.58.1.29" => 171573533.
func IpItoa(iIp int64) string {
	var bs [4]byte
	bs[0] = byte(iIp & 0xFF)
	bs[1] = byte((iIp >> 8) & 0xFF)
	bs[2] = byte((iIp >> 16) & 0xFF)
	bs[3] = byte((iIp >> 24) & 0xFF)

	return net.IPv4(bs[3], bs[2], bs[1], bs[0]).String()
}

// IpAtoi IP string to integer.
// eg: 171573533 => "10.58.1.29".
func IpAtoi(sIp string) int64 {
	if !IsIpv4(sIp) {
		return -1
	}

	bits := strings.Split(sIp, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

// Int32ToByte int32 to []byte.
// "isBig" is true if big endian is required, otherwise is false.
func Int32ToByte(i int32, isBig bool) []byte {
	bBuf := bytes.NewBuffer([]byte{})
	if isBig {
		binary.Write(bBuf, binary.BigEndian, i)
	} else {
		binary.Write(bBuf, binary.LittleEndian, i)
	}

	return bBuf.Bytes()
}

// ByteToInt32 []byte to int32.
// "isBig" is true if big endian is required, otherwise is false.
func ByteToInt32(b []byte, isBig bool) int32 {
	bBuf := bytes.NewBuffer(b)
	var i int32
	if isBig {
		binary.Read(bBuf, binary.BigEndian, &i)
	} else {
		binary.Read(bBuf, binary.LittleEndian, &i)
	}

	return i
}

// Int64ToByte int64 to []byte.
func Int64ToByte(i int64, isBig bool) []byte {
	bBuf := bytes.NewBuffer([]byte{})
	if isBig {
		binary.Write(bBuf, binary.BigEndian, i)
	} else {
		binary.Write(bBuf, binary.LittleEndian, i)
	}

	return bBuf.Bytes()
}

// ByteToInt64 []byte to int64.
func ByteToInt64(b []byte, isBig bool) int64 {
	bBuf := bytes.NewBuffer(b)
	var i int64
	if isBig {
		binary.Read(bBuf, binary.BigEndian, &i)
	} else {
		binary.Read(bBuf, binary.LittleEndian, &i)
	}

	return i
}

// Float32ToByte foat32 to []byte.
func Float32ToByte(f float32, isBig bool) []byte {
	bits := math.Float32bits(f)
	bs := make([]byte, 4)
	if isBig {
		binary.BigEndian.PutUint32(bs, bits)
	} else {
		binary.LittleEndian.PutUint32(bs, bits)
	}

	return bs
}

// ByteToFloat32 []byte to int32.
func ByteToFloat32(b []byte, isBig bool) float32 {
	var bits uint32
	if isBig {
		bits = binary.BigEndian.Uint32(b)
	} else {
		bits = binary.LittleEndian.Uint32(b)
	}

	return math.Float32frombits(bits)
}

// Float64ToByte foat64 to []byte.
// "isBig" is true if big endian is required, otherwise is false.
func Float64ToByte(f float64, isBig bool) []byte {
	bits := math.Float64bits(f)
	bs := make([]byte, 8)
	if isBig {
		binary.BigEndian.PutUint64(bs, bits)
	} else {
		binary.LittleEndian.PutUint64(bs, bits)
	}

	return bs
}

// ByteToFloat64 []byte to int64.
func ByteToFloat64(b []byte, isBig bool) float64 {
	var bits uint64
	if isBig {
		bits = binary.BigEndian.Uint64(b)
	} else {
		bits = binary.LittleEndian.Uint64(b)
	}

	return math.Float64frombits(bits)
}

// MapToUrlQuery Convert map to URL query format.
// URL transcoding must be performed.
// eg: a=11&b=22&c=33
func MapToUrlQuery(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	var buf bytes.Buffer
	i := 0
	for key, val := range m {
		if i == 0 {
			buf.WriteString(url.QueryEscape(key) + "=")
		} else {
			buf.WriteByte('&')
			buf.WriteString(url.QueryEscape(key) + "=")
		}
		buf.WriteString(url.QueryEscape(val))

		i++
	}

	return buf.String()
}

// UrlQueryToMap Convert URL query to map format.
func UrlQueryToMap(s string) (map[string]string, error) {
	m := make(map[string]string)
	if s == "" {
		return m, nil
	}

	l := strings.Split(s, "&")
	for _, v := range l {
		t := strings.Split(v, "=")
		key, err := url.QueryUnescape(t[0])
		if err != nil {
			return m, err
		}

		if len(t) == 1 {
			m[key] = ""
		} else if len(t) >= 2 {
			val, err := url.QueryUnescape(t[1])
			if err != nil {
				return m, err
			}
			m[key] = val
		}
	}

	return m, nil
}
