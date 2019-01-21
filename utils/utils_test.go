package utils

import (
	"fmt"
	"testing"
)

// TestUniqid test Uniqid function.
func TestUniqid(t *testing.T) {
	fmt.Println("uniqid1: " + Uniqid("le"))
	fmt.Println("uniqid2: " + Uniqid("le", true))
}

// TestGuid test Guid function.
func TestGuid(t *testing.T) {
	fmt.Println("guid: " + Guid())
}

// TestGetTopDomain test GetTopDomain function.
func TestGetTopDomain(t *testing.T) {
	domain := "http://www.lixy.com:9090?aa=11&b=22"
	topDomain := GetTopDomain(domain)
	if topDomain != "lixy.com" {
		t.Errorf("GetString failed. Got %s, expected lixy.com.", topDomain)
		return
	}

	domain = "localhost"
	topDomain = GetTopDomain(domain)
	if topDomain != "localhost" {
		t.Errorf("GetString failed. Got %s, expected localhost.", topDomain)
		return
	}

	domain = "us.sso.le.com"
	topDomain = GetTopDomain(domain)
	if topDomain != "le.com" {
		t.Errorf("GetString failed. Got %s, expected le.com.", topDomain)
		return
	}

	domain = "http://www.lixy.com:9090?aa=bb&ccc=dd"
	topDomain = GetTopDomain(domain)
	if topDomain != "lixy.com" {
		t.Errorf("GetString failed. Got %s, expected lixy.com.", topDomain)
		return
	}

	domain = "www.lixy.com?aa=bb&ccc=dd"
	topDomain = GetTopDomain(domain)
	if topDomain != "lixy.com" {
		t.Errorf("GetString failed. Got %s, expected lixy.com.", topDomain)
		return
	}

	domain = "127.0.0.1?aa=bb&ccc=dd"
	topDomain = GetTopDomain(domain)
	if topDomain != "127.0.0.1" {
		t.Errorf("GetString failed. Got %s, expected 127.0.0.1.", topDomain)
		return
	}

	domain = "www.lixy.com"
	topDomain = GetTopDomain(domain)
	if topDomain != "lixy.com" {
		t.Errorf("GetString failed. Got %s, expected lixy.com.", topDomain)
		return
	}
}

// TestGetLocalIp test GetLocalIp function.
func TestGetLocalIp(t *testing.T) {
	addr := GetLocalIp()
	if addr == "" {
		t.Errorf("GetLocalIp failed.")
		return
	}
	t.Log(addr)
}

// TestStack test Stack function.
func TestStack(t *testing.T) {
	stack := Stack(1, 6)
	fmt.Println(stack)
}

// TestKrand test Krand function.
func TestKrand(t *testing.T) {
	fmt.Println("num:    " + Krand(16, RAND_KIND_NUM))
	fmt.Println("lower:  " + Krand(16, RAND_KIND_LOWER))
	fmt.Println("upper:  " + Krand(16, RAND_KIND_UPPER))
	fmt.Println("letter: " + Krand(16, RAND_KIND_LETTER))
	fmt.Println("all:    " + Krand(16, RAND_KIND_ALL))
}

// TestIrand test Irand function.
func TestIrand(t *testing.T) {
	for i := 0; i <= 100; i++ {
		n := Irand(100, 100)
		fmt.Println(n)
		if n < 100 || n > 300 {
			t.Errorf("Irand failed. Got %d, expected 100-300.", n)
			return
		}
	}
}

// TestRangeInt test RangeInt function.
func TestRangeInt(t *testing.T) {
	res := RangeInt(0, 255)
	fmt.Println(res)
}

// TestGetTerminal test GetTerminal function.
func TestGetTerminal(t *testing.T) {
	tType, osType := GetTerminal("aaiPadbb")
	if tType != "pad" || osType != "ios" {
		t.Errorf("GetString failed. Got %s-%s, expected pad-ios.", tType, osType)
		return
	}

	tType, osType = GetTerminal("aaiPhonebb")
	if tType != "phone" || osType != "ios" {
		t.Errorf("GetString failed. Got %s-%s, expected phone-ios.", tType, osType)
		return
	}

	tType, osType = GetTerminal("aaWindows Phonebb")
	if tType != "phone" || osType != "win" {
		t.Errorf("GetString failed. Got %s-%s, expected pad-win.", tType, osType)
		return
	}

	tType, osType = GetTerminal("aaandroidbb")
	if tType != "phone" || osType != "android" {
		t.Errorf("GetString failed. Got %s-%s, expected pad-android.", tType, osType)
		return
	}

	tType, osType = GetTerminal("aamacbb")
	if tType != "pc" || osType != "mac" {
		t.Errorf("GetString failed. Got %s-%s, expected pc-mac.", tType, osType)
		return
	}

	tType, osType = GetTerminal("aawin ntbb")
	if tType != "pc" || osType != "win" {
		t.Errorf("GetString failed. Got %s-%s, expected pc-win.", tType, osType)
		return
	}
}

// TestSelStrVal test SelStrVal function.
func TestSelStrVal(t *testing.T) {
	opt1 := "aaa"
	opt2 := "bbb"
	opt := SelStrVal(true, opt1, opt2)
	if opt != opt1 {
		t.Errorf("SelStrVal err, Got:%s expected:%s", opt, opt1)
	}

	opt = SelStrVal(false, opt1, opt2)
	if opt != opt2 {
		t.Errorf("SelStrVal err, Got:%s expected:%s", opt, opt2)
	}
}

// TestSelIntVal test SelIntVal function.
func TestSelIntVal(t *testing.T) {
	opt1 := 111
	opt2 := 222
	opt := SelIntVal(true, opt1, opt2)
	if opt != opt1 {
		t.Errorf("SelIntVal err, Got:%d expected:%d", opt, opt1)
	}

	opt = SelIntVal(false, opt1, opt2)
	if opt != opt2 {
		t.Errorf("SelIntVal err, Got:%d expected:%d", opt, opt2)
	}
}
