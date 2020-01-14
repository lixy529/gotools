package utils

import (
	"testing"
)

// TestIpItoa test IpItoa function
func TestIpConvert(t *testing.T) {
	strIp1 := "10.58.1.29"
	var intIp1 int64 = 171573533
	intIp2 := IpAtoi(strIp1)
	if intIp1 != intIp2 {
		t.Errorf("IpAtoi failed. Got: %d, expected %d.", intIp1, intIp2)
		return
	}

	strIp2 := IpItoa(intIp1)
	if strIp1 != strIp2 {
		t.Errorf("IpAtoi failed. Got: %s, expected %s.", strIp1, strIp2)
		return
	}
}

// TestConvert Type Transcoding Function Testing
func TestConvert(t *testing.T) {
	var i32 int32
	i32 = 123
	// big endian
	b := Int32ToByte(i32, true)
	i32_2 := ByteToInt32(b, true)
	if i32_2 != i32 {
		t.Errorf("GetString failed. Got %d, expected %d.", i32_2, i32)
		return
	}
	// little endian
	b = Int32ToByte(i32, false)
	i32_2 = ByteToInt32(b, false)
	if i32_2 != i32 {
		t.Errorf("GetString failed. Got %d, expected %d.", i32_2, i32)
		return
	}

	var i64 int64
	i64 = 123456789
	// big endian
	b = Int64ToByte(i64, true)
	i64_2 := ByteToInt64(b, true)
	if i64_2 != i64 {
		t.Errorf("GetString failed. Got %d, expected %d.", i64_2, i64)
		return
	}
	// little endian
	b = Int64ToByte(i64, true)
	i64_2 = ByteToInt64(b, true)
	if i64_2 != i64 {
		t.Errorf("GetString failed. Got %d, expected %d.", i64_2, i64)
		return
	}

	var f32 float32
	f32 = 123.45
	// big endian
	b = Float32ToByte(f32, true)
	f32_2 := ByteToFloat32(b, true)
	if f32_2 != f32 {
		t.Errorf("GetString failed. Got %f, expected %f.", f32_2, f32)
		return
	}
	// little endian
	b = Float32ToByte(f32, false)
	f32_2 = ByteToFloat32(b, false)
	if f32_2 != f32 {
		t.Errorf("GetString failed. Got %f, expected %f.", f32_2, f32)
		return
	}

	var f64 float64
	f64 = 123.45
	// big endian
	b = Float64ToByte(f64, true)
	f64_2 := ByteToFloat64(b, true)
	if f64_2 != f64 {
		t.Errorf("GetString failed. Got %f, expected %f.", f64_2, f64)
		return
	}
	// little endian
	b = Float64ToByte(f64, false)
	f64_2 = ByteToFloat64(b, false)
	if f64_2 != f64 {
		t.Errorf("GetString failed. Got %f, expected %f.", f64_2, f64)
		return
	}
}

// TestMapToUrlQuery test MapToUrlQuery and  UrlQueryToMap function
func TestMapToUrlQuery(t *testing.T) {
	m1 := map[string]string{
		"aa": "100",
		"bb": "200",
		"cc": "中国",
	}

	// MapToUrlQuery
	str := MapToUrlQuery(m1)
	t.Log(str)

	// UrlQueryToMap
	m2, err := UrlQueryToMap(str)
	t.Log(m2)
	if err != nil {
		t.Errorf("GetString error. %s.", err.Error())
		return
	} else if m1["aa"] != m2["aa"] || m1["bb"] != m2["bb"] || m1["cc"] != m2["cc"] {
		t.Errorf("HttpQueryToMap failed. Got %s-%s-%s, expected %s-%s-%s.", m2["aa"], m2["bb"], m2["cc"], m1["aa"], m1["bb"], m1["cc"])
		return
	}
}
