package util

import (
	"time"
)

func TimestampSinceNow(t time.Duration) int64 {
	return time.Now().Add(t).UTC().UnixMilli()
}

func UnixTimeNow() int64 {
	return time.Now().UTC().UnixMilli()
}

// Ternary operator.
// If 'cond' is true then returns 'a', otherwise returns 'b'
func Ternary[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}

	return b
}

