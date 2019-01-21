package utils

import (
	"regexp"
	"strings"
)

const (
	IPV4 = 1
	IPV6 = 2
	IPVX = 3
)

// CheckIp check the "ip" is an IP address format.
// The range of "ipv" is IPV4、IPV6、IPVX, default is IPV4
func CheckIp(ip string, ipv ...int) bool {
	ipv = append(ipv, IPV4)

	if ipv[0] & IPV4 == IPV4 && IsIpv4(ip) {
		return true
	}

	if ipv[0] & IPV6 == IPV6 && IsIpv6(ip) {
		return true
	}

	return false
}

// IsIpv4 check the "ip" is an IPv4 address format.
func IsIpv4(ip string) bool {
	pattern := `^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	return false
}

// IsIpv6 check the "ip" is an IPv6 address format.
func IsIpv6(ip string) bool {
	// CDCD:910A:2222:5498:8475:1111:3900:2020
	pattern := `^([0-9a-fA-Z]{1,4}:){7}[0-9a-fA-Z]{1,4}$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// F:F:F::1:1 F:F:F:F:F::1 F::F:F:F:F:1
	pattern = `^(([0-9a-fA-Z]{1,4}:){0,6})((:[0-9a-fA-Z]{1,4}){0,6})$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		t := strings.Split(ip, ":")
		if len(t) > 0 && len(t) <= 8 {
			return true
		}
	}

	// F:F:10F::
	pattern = `^([0-9a-fA-F]{1,4}:){1,7}:$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// ::F:F:10F
	pattern = `^:(:[0-9a-fA-F]{1,4}){1,7}$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// F:0:0:0:0:0:10.0.0.1
	pattern = `^([0-9a-fA-F]{1,4}:){6}((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// F::10.0.0.1
	pattern = `^([0-9a-fA-F]{1,4}:){1,5}:((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// ::10.0.0.1
	pattern = `^::((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	return false
}

// CheckEmail check mailbox format.
func CheckEmail(email string) bool {
	//pattern := "[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?"
	pattern := `^[A-Za-z\d]+([-_.][A-Za-z\d]+)*@([A-Za-z\d]+[-.])+[A-Za-z\d]{2,4}$`
	if m, _ := regexp.MatchString(pattern, email); m {
		return true
	}

	return false
}

// CheckMobile check mobile format.
func CheckMobile(mobile string) bool {
	pattern := `^(1[3|4|5|7|8])\d{9}$`
	if m, _ := regexp.MatchString(pattern, mobile); m {
		return true
	}

	return false
}
