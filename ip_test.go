package gosak

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckIPInRange(t *testing.T) {
	start := net.ParseIP("10.240.0.0")
	end := net.ParseIP("10.240.255.255")

	assert.Equal(t, true, CheckIPInRange("10.240.176.197", start, end))
	assert.Equal(t, false, CheckIPInRange("220.128.223.100", start, end))
}
