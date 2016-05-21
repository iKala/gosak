package gosak

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLastSplit(t *testing.T) {
	src := "https://www.googleapis.com/compute/v1"

	assert.Equal(t, "v1", GetLastSplit(src, "/"))
}

func TestRandSeq(t *testing.T) {
	s := RandSeq(5)

	assert.Equal(t, 5, len(s))
}

func TestStringValidate(t *testing.T) {
	lenValidator := func(s string) bool {
		if len(s) > 3 {
			return true
		}
		return false
	}
	containValidator := func(s string) bool {
		if strings.Contains(s, "abc") {
			return true
		}
		return false
	}
	assert.True(t, StringValidate("abcde", lenValidator, containValidator))

	emptyValidator := func(s string) bool {
		if s == "" {
			return true
		}
		return false
	}
	assert.False(t, StringValidate("abcde", lenValidator, containValidator, emptyValidator))
}
