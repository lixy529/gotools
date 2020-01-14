package utils

import (
	"encoding/json"
	"fmt"
	"testing"
)

// TestSha1 test Sha1
func TestSha1(t *testing.T) {
	str := "Hello World!"
	sh := Sha1(str)
	t.Log(sh)
}

// TestHmacSha1 test Sha1 function.
func TestHmacSha1(t *testing.T) {
	str := "Hello World!"
	key := "123456"
	sh := HmacSha1(str, key)
	fmt.Println(sh)

	sh = HmacSha1(str, key, true)
	t.Log(sh)
}

// TestGobEncode test GobEncode and GobDecode function.
func TestGobEncode(t *testing.T) {
	str := "Hello World!"
	// encode
	e, err := GobEncode(str)
	if err != nil {
		t.Errorf("GobEncode failed. err: %s.", err.Error())
		return
	}

	// decode
	infoNew := new(string)
	err = GobDecode(e, infoNew)
	if err != nil {
		t.Errorf("GobDecode failed. err: %s.", err.Error())
		return
	} else if *infoNew != str {
		t.Errorf("GobDecode failed. Got %s, expected %s.", *infoNew, str)
		return
	}
}

// TestZlibEncode test ZlibEncode and ZlibDecode function.
func TestZlibEncode(t *testing.T) {
	s1 := "Hello World!"
	// compression
	b, err := ZlibEncode([]byte(s1))
	if err != nil {
		t.Errorf("ZlibEncode failed. err: %s.", err.Error())
		return
	}

	// uncompression
	s2, err := ZlibDecode(b)
	if err != nil {
		t.Errorf("ZlibDecode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("ZlibDecode failed. Got %s, expected %s.", s2, s1)
		return
	}
}

// TestZlibEncode test RsaEncode and RsaDecode function.
func TestRsaEncode(t *testing.T) {
	s1 := "Hello World!"
	pubPem := "./data/public.pem"
	priPem := "./data/private.pem"

	// public key encode
	b, err := RsaEncode([]byte(s1), pubPem, MODE_RSA_PUB)
	if err != nil {
		t.Errorf("RsaEncode failed. err: %s.", err.Error())
		return
	}

	// private key decode
	s2, err := RsaDecode(b, priPem, MODE_RSA_PRI)
	if err != nil {
		t.Errorf("RsaDecode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("RsaDecode failed. Got %s, expected %s.", s2, s1)
		return
	}

	// private key endoce
	b, err = RsaEncode([]byte(s1), priPem, MODE_RSA_PRI)
	if err != nil {
		t.Errorf("RsaEncode failed. err: %s.", err.Error())
		return
	}

	// public key decode
	s2, err = RsaDecode(b, pubPem, MODE_RSA_PUB)
	if err != nil {
		t.Errorf("RsaDecode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("RsaDecode failed. Got %s, expected %s.", s2, s1)
		return
	}
}

// TestBase64Encode test Base64Encode and Base64Decode function.
func TestBase64Encode(t *testing.T) {
	s1 := "Hello World!"

	// encode
	str := Base64Encode([]byte(s1), "m", "m1", "+", "m2", "/", "m3")

	// decode
	s2, err := Base64Decode(str, "m3", "/", "m2", "+", "m1", "m")
	if err != nil {
		t.Errorf("Base64Decode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("Base64Decode failed. Got %s, expected %s.", s2, s1)
		return
	}
}

// TestPadding test Padding and UnPadding function.
func TestPadding(t *testing.T) {
	src := []byte("12345")
	dst := Padding(src, 16)
	if len(dst) != 16 {
		t.Errorf("Padding failed. Got %d, expected %d.", len(dst), 16)
		return
	}
	src2 := UnPadding(dst)
	if string(src) != string(src2) {
		t.Errorf("UnPadding failed. Got %s, expected %s.", src2, src)
		return
	}
}

// TestDesEncode test DesEncode and DesDecode function.
func TestDesEncode(t *testing.T) {
	key := "12345678"
	src := "HelloWorld!"
	dst, err := DesEncode([]byte(src), []byte(key))
	if err != nil {
		t.Errorf("DesEncode error. %s.", err.Error())
		return
	}

	src2, err := DesDecode(dst, []byte(key))
	if err != nil {
		t.Errorf("DesDecode error. %s.", err.Error())
		return
	} else if src != string(src2) {
		t.Errorf("DesDecode failed. Got %s, expected %s.", src2, src)
		return
	}
}

// TestAesEncode test AesEncode and AesDecode function.
func TestAesEncode(t *testing.T) {
	key := "12345678901234561234567890123456"
	src := "HelloWorld!"
	dst, err := AesEncode([]byte(src), []byte(key))
	if err != nil {
		t.Errorf("AesEncode error. %s.", err.Error())
		return
	}

	src2, err := AesDecode(dst, []byte(key))
	if err != nil {
		t.Errorf("AesDecode error. %s.", err.Error())
		return
	} else if src != string(src2) {
		t.Errorf("AesDecode failed. Got %s, expected %s.", src2, src)
		return
	}
}

// TestStrToJSON test StrToJSON function.
func TestUnicodeEncode(t *testing.T) {
	info := `{"aa":"(✪ω✪)"}`
	strJson := UnicodeEncode([]byte(info))
	t.Log(strJson)
	arr := make(map[string]string)
	err := json.Unmarshal([]byte(strJson), &arr)
	if err != nil {
		t.Errorf("Unmarshal err: %s", err.Error())
		return
	}
	t.Log("00:", arr)

	info = `{"aa":"\ud83d\ude02"}`
	json.Unmarshal([]byte(info), &arr)
	t.Log("11:", arr)
	b, _ := json.Marshal(arr)
	strJson = UnicodeEncode(b)
	t.Log(strJson)
	err = json.Unmarshal([]byte(strJson), &arr)
	if err != nil {
		t.Errorf("Unmarshal err: %s", err.Error())
		return
	}
	t.Log(arr)
}
