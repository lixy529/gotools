package utils

import (
	"testing"
)

// TestGet test Curl function
func TestGet(t *testing.T) {
	url := "http://www.baidu.com"
	data := ""
	flag := "get"
	params := make(map[int]interface{})

	res, status, err := Curl(url, data, flag, 1, params)
	if err != nil {
		t.Errorf("Curl failed. err: %s.", err.Error())
		return
	}
	t.Log(res, status)
}

// TestPost test Curl function
func TestPost(t *testing.T) {
	url := "http://test.lixy.com/lixy/postTest"
	data := "name=lixiaoya&age=50"
	flag := "post"
	params := make(map[int]interface{})

	res, status, err := Curl(url, data, flag, 1, params)
	if err != nil {
		t.Errorf("Curl failed. err: %s.", err.Error())
		return
	}
	t.Log(res, status)
}

// TestPost test Curl function
func TestProxy(t *testing.T) {
	url := "https://timgsa.baidu.com/timg?image&quality=80&size=b9999_10000&sec=1489061813047&di=f142d8e974ce848bafc4320536279f18&imgtype=0&src=http%3A%2F%2Fwww.th7.cn%2FArticle%2FUploadFiles%2F200907%2F20090710075404170.jpg"
	data := ""
	flag := "get"
	params := make(map[int]interface{})
	params[OPT_PROXY] = "http://10.135.108.235:2443"

	res, status, err := Curl(url, data, flag, 5, params)
	if err != nil {
		t.Errorf("Curl failed. err: %s.", err.Error())
		return
	}
	t.Log(res, status)
}

// TestPost test Curl function
func TestProxy1(t *testing.T) {
	url := "http://test.lixy.com/api/getUserByID?uid=147703777"
	data := ""
	flag := "get"
	params := make(map[int]interface{})
	res, status, err := Curl(url, data, flag, 1, params)
	if err != nil {
		t.Errorf("Curl failed. err: %s.", err.Error())
		return
	}
	t.Log(res, status)
}

// TestPost test Curl function
func TestCurl3(t *testing.T) {
	url := "https://aaa.com/aaa/bbb"
	data := `{"appId":"10000","authCode":"ssss","sign":"9cc49c85ad574be0f6d93c63505670e0"}`
	method := "post"

	params := make(map[int]interface{})
	params[OPT_SSLCERT] = map[string]string{
		"certFile": "./cert.pem",
		"keyFile":  "./key.pem",
	}

	res, status, err := Curl(url, data, method, 5, params)
	if err != nil {
		t.Errorf("Curl error: %s.", err.Error())
		return
	}
	t.Log(res, status)
}

// TestPostFile test PostFile function
func TestPostFile(t *testing.T) {
	url := "https://aaa.com/aaa/bbb"
	fileKey := "image"
	filePath := "./data/time.jpeg"
	fields := map[string]string{
		"token":  "xxx101xxx",
		"mobile": "15812345678",
	}
	res, status, err := PostFile(url, fileKey, filePath, fields, 5)
	if err != nil {
		t.Errorf("PostFile error: %s.", err.Error())
		return
	}
	t.Log(res, status)
}
