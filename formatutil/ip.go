package formatutil

import (
	"bytes"
	"log"
	"net"
)

// CheckIPInRange check input `ip` is between `start` and `end` IP range
func CheckIPInRange(ip string, start net.IP, end net.IP) bool {
	//sanity check
	input := net.ParseIP(ip)
	if input.To4() == nil {
		log.Printf("%v is not a valid IPv4 address", input)
		return false
	}

	if bytes.Compare(input, start) >= 0 && bytes.Compare(input, end) <= 0 {
		return true
	}

	log.Printf("%v is NOT between %v and %v", input, start, end)
	return false
}
