package gosak

import (
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
