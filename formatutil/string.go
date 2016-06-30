package formatutil

import (
	"math/rand"
	"strings"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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

// StringValidate return true if the input string pass all validator checking
func StringValidate(body string, filters ...func(body string) bool) bool {
	for _, filter := range filters {
		if !filter(body) {
			return false
		}
	}
	return true
}
