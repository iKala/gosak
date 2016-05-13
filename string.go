package gosak

import (
	"math/rand"
	"strings"
	"time"
)

// GetLastSplit get the last part of splits
func GetLastSplit(src, separator string) string {
	split := strings.Split(src, separator)
	return split[len(split)-1]
}

// RandSeq generates the random string of length n
func RandSeq(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
