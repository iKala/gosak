package gosak

import (
	"log"
	"testing"
)

func TestIsIP(t *testing.T) {
	testedIP := "1.3.1.2"
	log.Printf("%s is ip address: %t", testedIP, IsIP(testedIP))

	testedIP = "MISSING"
	log.Printf("%s is ip address: %t", testedIP, IsIP(testedIP))
}
