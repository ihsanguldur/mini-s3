package s3api

import (
	"net"
	"strings"
)

func validBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !isAlnum(c) && c != '-' && c != '.' {
			return false
		}
	}
	if !isAlnum(name[0]) || !isAlnum(name[len(name)-1]) {
		return false
	}
	if strings.Contains(name, "..") {
		return false
	}
	if net.ParseIP(name) != nil {
		return false
	}
	if strings.HasPrefix(name, "xn--") || strings.HasSuffix(name, "-s3alias") {
		return false
	}
	return true
}

func isAlnum(c byte) bool {
	return 'a' <= c && c <= 'z' || '0' <= c && c <= '9'
}
