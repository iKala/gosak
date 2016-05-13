package gosak

import (
	"time"
)

// MultipleConfirm return true if `isTenable` functor holds for X times continuously, else return false
func MultipleConfirm(times int, isTenable func() bool, retryInterval time.Duration) bool {
	for i := 0; i < times; i++ {
		if !isTenable() {
			return false
		}
		time.Sleep(retryInterval)
	}

	return true
}

// RunUntilSuccess tries to run `f` function most `n` times until `f` return true, else return false
func RunUntilSuccess(n int, f func() bool, retryInterval time.Duration) bool {
	for i := 0; i < n; i++ {
		if f() {
			return true
		}

		time.Sleep(retryInterval)
	}

	return false
}
