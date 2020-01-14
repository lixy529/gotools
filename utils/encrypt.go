package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	MODE_RSA_PUB = 1 // public key encode or public key decode
	MODE_RSA_PRI = 2 // private key encode or private key decode
)

// Md5 returns a 32-bit md5 string.
func Md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Sha1 returns sha1 string.
func Sha1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// HmacSha1 returns hash_hmac string.
// Base64 encoding if "isBase64" is true, defautl is false.
func HmacSha1(s, k string, isBase64 ...bool) string {
	h := hmac.New(sha1.New, []byte(k))
	h.Write([]byte(s))

	if len(isBase64) > 0 && isBase64[0] {
		return Base64Encode(h.Sum(nil))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// GobEncode returns encode string by gob.
func GobEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobDecode returns decode string by gob.
func GobDecode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

// ZlibEncode Zlib compression.
func ZlibEncode(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		w.Close()
		return nil, err
	}
	w.Close()

	return b.Bytes(), nil
}

// ZlibDecode Zlib uncompression.
func ZlibDecode(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.Bytes(), err
}

// RsaEncode returns encode string by rsa.
// "pemFile" paramater is the pem certificate.
// public key encode when "mode" is MODE_RSA_PUB, private key encode when "mode" is MODE_RSA_PRI.
func RsaEncode(data []byte, pemFile string, mode int) ([]byte, error) {
	if mode == MODE_RSA_PRI {
		// private key encode
		privateKey, err := ioutil.ReadFile(pemFile)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(privateKey)
		if block == nil {
			return nil, errors.New("private key error!")
		}
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		return EncPKCS1v15(rand.Reader, priv, data)
	}

	// public key encode
	publicKey, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

// RsaDecode returns decode string by rsa.
// Reference resources https://github.com/dgkang/rsa .
// "pemFile" paramater is the pem certificate.
// public key decode when "mode" is MODE_RSA_PUB, private key decode when "mode" is MODE_RSA_PRI.
func RsaDecode(data []byte, pemFile string, mode int) ([]byte, error) {
	if mode == MODE_RSA_PUB {
		// private decode
		publicKey, err := ioutil.ReadFile(pemFile)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(publicKey)
		if block == nil {
			return nil, errors.New("public key error")
		}

		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		pub := pubInterface.(*rsa.PublicKey)

		return DecPKCS1v15(pub, data)
	}

	// private decode
	privateKey, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, data)
}

// Base64Encode returns encode string by base64.
// "replaces" is the data to be replaced.
func Base64Encode(data []byte, replaces ...string) string {
	str := base64.StdEncoding.EncodeToString(data)
	n := len(replaces)
	if n > 1 {
		var olds []string
		var news []string
		for i := 0; i < n-1; i += 2 {
			olds = append(olds, replaces[i])
			news = append(news, replaces[i+1])
		}

		str = Replace(str, olds, news, -1)
	}

	return str
}

// Base64Decode returns decode string by base64
// "replaces" is the data to be replaced.
func Base64Decode(data string, replaces ...string) ([]byte, error) {
	n := len(replaces)
	if n > 1 {
		var olds []string
		var news []string
		for i := 0; i < n-1; i += 2 {
			olds = append(olds, replaces[i])
			news = append(news, replaces[i+1])
		}

		data = Replace(data, olds, news, -1)
	}

	// Handling non-standard encrypted strings
	if m := len(data) % 4; m != 0 {
		data += strings.Repeat("=", 4-m)
	}

	return base64.StdEncoding.DecodeString(data)
}

// Padding add text to an integer multiple of blockSize.
// If the text length is exactly an integer multiple of blockSize, then blockSize long data will be added.
func Padding(text []byte, blockSize int) []byte {
	padding := blockSize - len(text)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(text, padtext...)
}

// UnPadding delete Padding filled data.
func UnPadding(text []byte) []byte {
	length := len(text)
	unpadding := int(text[length-1])

	if length-unpadding < 0 || unpadding < 0 {
		return []byte("")
	}

	return text[:(length - unpadding)]
}

// DesEncode returns encode string by des.
// The key must be 8 digits and the excess will be truncated.
func DesEncode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l > 8 {
		key = key[:8]
	} else if l < 8 {
		padtext := bytes.Repeat([]byte("x"), 8-l)
		key = append(key, padtext...)
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	data = Padding(data, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	encode := make([]byte, len(data))
	blockMode.CryptBlocks(encode, data)
	return encode, nil
}

// DesDecode returns decode string by des.
// The key must be 8 digits and the excess will be truncated.
func DesDecode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l > 8 {
		key = key[:8]
	} else if l < 8 {
		padtext := bytes.Repeat([]byte("x"), 8-l)
		key = append(key, padtext...)
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, key)
	decode := data
	blockMode.CryptBlocks(decode, data)
	decode = UnPadding(decode)
	return decode, nil
}

// AesEncode returns encode string by aes.
// The key must be a multiple of 16, and the excess will be truncated.
func AesEncode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l < 16 {
		padtext := bytes.Repeat([]byte("x"), 16-l)
		key = append(key, padtext...)
	} else if l%16 != 0 {
		var n int = l / 16 * 16
		key = key[:n]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	data = Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encode := make([]byte, len(data))
	blockMode.CryptBlocks(encode, data)
	return encode, nil
}

// AesDecode returns decode string by aes.
// The key must be a multiple of 16, and the excess will be truncated.
func AesDecode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l < 16 {
		padtext := bytes.Repeat([]byte("x"), 16-l)
		key = append(key, padtext...)
	} else if l%16 != 0 {
		var n int = l / 16 * 16
		key = key[:n]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	decode := make([]byte, len(data))
	blockMode.CryptBlocks(decode, data)
	decode = UnPadding(decode)

	return decode, nil

}

// UnicodeEncode Encoding the wide characters in the JSON string.
func UnicodeEncode(b []byte) string {
	var buf bytes.Buffer
	strLen := len(b)
	for i := 0; i < strLen; i++ {
		c1 := int64(b[i])
		// Single byte
		if c1 < 128 {
			if c1 > 31 {
				buf.WriteByte(b[i])
			} else {
				buf.WriteString(fmt.Sprintf("\\u%04s", strconv.FormatInt(c1, 16)))
			}
			continue
		}

		// Double byte
		i++
		if i >= strLen {
			break
		}
		c2 := int64(b[i])
		if c1&32 == 0 {
			buf.WriteString(fmt.Sprintf("\\u%04s", strconv.FormatInt((c1-192)*64+c2-128, 16)))
			continue
		}

		// Triple
		i++
		if i >= strLen {
			break
		}
		c3 := int64(b[i])
		if c1&16 == 0 {
			buf.WriteString(fmt.Sprintf("\\u%04s", strconv.FormatInt(((c1-224)<<12)+((c2-128)<<6)+(c3-128), 16)))
			continue
		}

		// Quadruple
		i++
		if i >= strLen {
			break
		}
		c4 := int64(b[i])
		if c1&8 == 0 {
			var u int64 = ((c1 & 15) << 2) + ((c2 >> 4) & 3) - 1

			var w1 int64 = (54 << 10) + (u << 6) + ((c2 & 15) << 2) + ((c3 >> 4) & 3)
			var w2 int64 = (55 << 10) + ((c3 & 15) << 6) + (c4 - 128)
			buf.WriteString(fmt.Sprintf("\\u%04s\\u%04s", strconv.FormatInt(w1, 16), strconv.FormatInt(w2, 16)))
		}
	}

	return buf.String()
}
