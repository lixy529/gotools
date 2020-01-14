package utils

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	OPT_PROXY = iota
	OPT_HTTPHEADER
	OPT_SSLCERT
)

// Curl Execute HTTP requests.
// return body string, http status and error message.
// "data" parameter is request data, JSON or URL query format, but only POST methods can use JSON format.
// "method" parameter contains POST and GET, default is GET.
// "timeout" parameter is the timeout time, the unit is seconds, default is 5 seconds.
// "params" supports the following parameters:
//   OPT_PROXY:      proxy server, eg: http://10.12.34.53:2443.
//   OPT_SSLCERT:    ssl certificate, map[string]string type, certFile、keyFile（key certificate, use the cert certificate when it is empty）、caFile（ca certificate, can be empty）.
//   OPT_HTTPHEADER: http header, map[string]string type.
func Curl(urlAddr, data, method string, timeout time.Duration, params ...map[int]interface{}) (string, int, error) {
	if timeout <= 0 {
		timeout = 5
	}
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	if strings.ToUpper(method) == "POST" {
		method = "POST"

		if data != "" && data[0] == '{' { // json: {"a":"1", "b":"2", "c":"3"}
			headers["Content-Type"] = "application/json; charset=utf-8"
		}
	} else {
		method = "GET"

		if data != "" {
			if strings.Contains(urlAddr, "?") {
				urlAddr = urlAddr + "&" + data
			} else {
				urlAddr = urlAddr + "?" + data
			}
			data = ""
		}
	}

	tpFlag := false
	tp := http.Transport{}
	if len(params) > 0 {
		param := params[0]

		// set proxy
		if v, ok := param[OPT_PROXY]; ok {
			if proxyAddr, ok := v.(string); ok {
				proxy := func(_ *http.Request) (*url.URL, error) {
					return url.Parse(proxyAddr)
				}
				tp.Proxy = proxy
				tpFlag = true
			}
		}

		// set certificate
		if v, ok := param[OPT_SSLCERT]; ok {
			if t, ok := v.(map[string]string); ok {
				if certFile, ok := t["certFile"]; ok && certFile != "" {
					keyFile := ""
					if keyFile, ok = t["keyFile"]; !ok || keyFile == "" {
						keyFile = certFile
					}
					caFile, _ := t["caFile"]

					tlsCfg, err := parseTLSConfig(certFile, keyFile, caFile)
					if err == nil {
						tp.TLSClientConfig = tlsCfg
						tpFlag = true
					}
				}
			}
		}

		// set header
		if v, ok := param[OPT_HTTPHEADER]; ok {
			if t, ok := v.(map[string]string); ok {
				for key, val := range t {
					headers[key] = val
				}
			}
		}
	}

	req, err := http.NewRequest(method, urlAddr, strings.NewReader(data))
	if err != nil {
		return "", -1, err
	}

	// set header
	for key, val := range headers {
		if strings.ToLower(key) == "host" {
			req.Host = val
		} else {
			req.Header.Set(key, val)
		}
	}

	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	// set Transport
	if tpFlag {
		client.Transport = &tp
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", -1, err
	}

	return string(body), resp.StatusCode, nil
}

// parseTLSConfig parse TLS certificate file.
func parseTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// load cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsCfg := tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}

	// load root ca
	if caFile != "" {
		caData, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		tlsCfg.RootCAs = pool
	}

	return &tlsCfg, nil
}

// PostFile http post a file.
// Currently only one file submission is supported.
// return body string, http status and error message.
// "fileKey" parameter is the key for file.
// "filePath" parameter is the path for file.
// "fields" parameter is the other parameter.
// "timeout" parameter is the timeout time, the unit is seconds, default is 5 seconds.
// "params" supports the following parameters:
//   OPT_PROXY:      proxy server, eg: http://10.12.34.53:2443.
//   OPT_SSLCERT:    ssl certificate, map[string]string type, certFile、keyFile（key certificate, use the cert certificate when it is empty）、caFile（ca certificate, can be empty）.
//   OPT_HTTPHEADER: http header, map[string]string type.
func PostFile(urlAddr, fileKey, filePath string, fields map[string]string, timeout time.Duration, params ...map[int]interface{}) (string, int, error) {
	if timeout <= 0 {
		timeout = 5
	}
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// 添加文件
	fileWriter, err := bodyWriter.CreateFormFile(fileKey, filePath)
	if err != nil {
		return "", -1, err
	}

	// 打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		return "", -1, err
	}
	defer fh.Close()

	// iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return "", -1, err
	}

	// add other paramater
	for k, v := range fields {
		bodyWriter.WriteField(k, v)
	}

	// set Content-Type
	headers := map[string]string{
		"Content-Type": bodyWriter.FormDataContentType(),
	}
	bodyWriter.Close()

	tpFlag := false
	tp := http.Transport{}
	if len(params) > 0 {
		param := params[0]

		// set proxy
		if v, ok := param[OPT_PROXY]; ok {
			if proxyAddr, ok := v.(string); ok {
				proxy := func(_ *http.Request) (*url.URL, error) {
					return url.Parse(proxyAddr)
				}
				tp.Proxy = proxy
				tpFlag = true
			}
		}

		// set certificate
		if v, ok := param[OPT_SSLCERT]; ok {
			if t, ok := v.(map[string]string); ok {
				if certFile, ok := t["certFile"]; ok && certFile != "" {
					keyFile := ""
					if keyFile, ok = t["keyFile"]; !ok || keyFile == "" {
						keyFile = certFile
					}
					caFile, _ := t["caFile"]

					tlsCfg, err := parseTLSConfig(certFile, keyFile, caFile)
					if err == nil {
						tp.TLSClientConfig = tlsCfg
						tpFlag = true
					}
				}
			}
		}

		// set header
		if v, ok := param[OPT_HTTPHEADER]; ok {
			if t, ok := v.(map[string]string); ok {
				for key, val := range t {
					headers[key] = val
				}
			}
		}
	}

	// post data
	req, err := http.NewRequest("POST", urlAddr, bodyBuf)
	if err != nil {
		return "", -1, err
	}

	// set header
	for key, val := range headers {
		if strings.ToLower(key) == "host" {
			req.Host = val
		} else {
			req.Header.Set(key, val)
		}
	}

	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	// set Transport
	if tpFlag {
		client.Transport = &tp
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", -1, err
	}

	return string(body), resp.StatusCode, nil
}
