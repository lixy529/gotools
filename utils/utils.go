package utils

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"path"
	"regexp"
	"net/url"
)

const (
	RAND_KIND_NUM    = 0 // Numbers
	RAND_KIND_LOWER  = 1 // Lowercase letters
	RAND_KIND_UPPER  = 2 // Uppercase letter
	RAND_KIND_LETTER = 3 // Lowercase and Uppercase letters
	RAND_KIND_ALL    = 4 // Numbers, Lowercase and Uppercase letters
)

// Uniqid return unique string.
// r: Whether to add a random string, eg: xx_5c4530a0634c877.66605056
func Uniqid(prefix string, r ...bool) string {
	t := time.Now()
	str := fmt.Sprintf("%s%x%x", prefix, t.Unix(), t.UnixNano() - t.Unix() * 1000000000)
	if len(r) > 0 && r[0] {
		str += "." + Krand(8, RAND_KIND_NUM)
	}
	return str
}

// Guid return guid.
func Guid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(crand.Reader, b); err != nil {
		return ""
	}
	return Md5(base64.URLEncoding.EncodeToString(b) + Uniqid(""))
}

// Krand returns a random string.
// size is String length.
// kind values are RAND_KIND_NUM, RAND_KIND_LOWER, RAND_KIND_UPPER, RAND_KIND_LETTER and RAND_KIND_ALL.
func Krand(size int, kind int) string {
	ikind, bases, scopes, result := kind, []int{48, 97, 65}, []int{10, 26, 26}, make([]byte, size)
	is_all := kind > 3 || kind < 0
	mrand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if is_all {
			ikind = mrand.Intn(3)
		} else if kind == RAND_KIND_LETTER {
			ikind = RAND_KIND_LOWER + mrand.Intn(2)
		}

		base, scope := bases[ikind], scopes[ikind]
		result[i] = uint8(base + mrand.Intn(scope))
	}
	return string(result)
}

// Irand Returns a random number of the specified range [start, end].
func Irand(start, end int) int {
	if start >= end {
		return end
	}
	mrand.Seed(time.Now().UnixNano())
	ikind := mrand.Intn(end - start + 1) + start
	return ikind
}

// RangeInt return the slice from start to end, range [start, end].
// eg: [0 1 2 3 4 5 6 7 8 9 10]
func RangeInt(start, end int) []int {
	res := make([]int, end - start + 1)
	for i := 0; i <= end - start; i++ {
		res[i] = start + i
	}

	return res
}

// GetTopDomain return the top domain.
func GetTopDomain(domain string) string {
	if domain == "" {
		return ""
	}

	// parse url
	domain = strings.ToLower(domain)
	urlAddr := domain
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		urlAddr = "http://" + domain
	}
	urlObj, err := url.Parse(urlAddr)
	if err != nil {
		return domain
	}
	urlHost := urlObj.Host
	if strings.Contains(urlHost, ":") {
		urlList := strings.Split(urlHost, ":")
		urlHost = urlList[0]
	}
	if urlHost == "" {
		return domain
	}

	// ip
	if CheckIp(urlHost) {
		return urlHost
	}

	// top domain
	domainParts := strings.Split(urlHost, ".")
	l := len(domainParts)
	if l > 1 {
		urlHost = domainParts[l-2] + "." + domainParts[l-1]
	}

	return urlHost
}

// GetLocalIp return local IP
func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// Stack returns the function name, file name, and number of lines of the calling code.
// depth is the start depth and end depth of the stack, default 1 to 10
// Format is "function name:file name:number of lines"
func Stack(depth ...int) string {
	var stack string
	var start int = 1
	var end int = 20
	if len(depth) > 0 {
		start = depth[0]
	}
	if len(depth) > 1 {
		end = depth[1]
	}

	for i := start; i < end; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if ok {
			funcName := runtime.FuncForPC(pc).Name()
			names := strings.Split(funcName, "/")
			if len(names) > 0 {
				funcName = names[len(names)-1]
			}

			if stack == "" {
				stack = fmt.Sprintf("\n%v:%v:%v", funcName, file, line)
			} else {
				stack = stack + "\n" + fmt.Sprintf("%v:%v:%v", funcName, file, line)
			}
		}
	}

	return stack
}

// GetCall 获取调用代码文件名和行数 returns the function name and number of lines of the calling code.
// depth is the start depth and end depth of the stack.
func GetCall(depth int) (string, int) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "unknown"
		line = 0
	}

	list := strings.Split(file, "/")
	n := len(list)
	if n > 1 {
		file = path.Join(list[n-2], list[n-1])
	}

	return file, line
}

// HandleSignals Returns the captured signal number and name.
func HandleSignals() (os.Signal, string) {
	var sig os.Signal
	signalChan := make(chan os.Signal)

	signal.Notify(
		signalChan,
		syscall.SIGTERM,
		syscall.SIGUSR2,
		syscall.SIGHUP,
	)

	for {
		sig = <-signalChan

		switch sig {
		case syscall.SIGTERM:
			return sig, "SIGTERM"

		case syscall.SIGHUP:
			return sig, "SIGHUP"

		case syscall.SIGUSR2:
			return sig, "SIGUSR2"

		default:
			return sig, "unknown"
		}
	}
}

// GetTerminal return client terminal information.
// terminal type: pc, phone, pad.
// terminal os: win, unix, linux, mac, ios, android.
func GetTerminal(userAgent string) (string, string) {
	userAgent = strings.ToLower(userAgent)

	if m, _ := regexp.MatchString("ipad", userAgent); m {
		return "pad", "ios"
	} else if m, _ := regexp.MatchString("jakarta|iphone|ipod", userAgent); m {
		return "phone", "ios"
	} else if m, _ := regexp.MatchString("windows phone", userAgent); m {
		return "phone", "win"
	} else if m, _ := regexp.MatchString("resty|android", userAgent); m {
		return "phone", "android"
	} else if m, _ := regexp.MatchString("mac", userAgent); m {
		return "pc", "mac"
	} else {
		return "pc", "win"
	}

	return "pc", "win"
}

// SelStrVal return options based on conditions.
// Options is string.
// If the condition is true, return opt1, otherwise return opt2.
func SelStrVal(con bool, opt1, opt2 string) string {
	if con {
		return opt1
	}

	return opt2
}

// SelIntVal return options based on conditions.
// Options is int.
// If the condition is true, return opt1, otherwise return opt2.
func SelIntVal(con bool, opt1, opt2 int) int {
	if con {
		return opt1
	}

	return opt2
}
