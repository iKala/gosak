package gosak

import (
	"regexp"
)

const (
	ipRegexp = "^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$"
)

// IsIP check the input is IP or not
func IsIP(ip string) bool {
	if m, _ := regexp.MatchString(ipRegexp, ip); !m {
		return false
	}

	return true
}
