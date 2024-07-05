package util

import (
	"os"
	"os/exec"
	"time"
)

func ClearTerminal() {
	// Currently program will run only on Linux
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func TimestampSinceNow(t time.Duration) int64 {
	return time.Now().Add(t).UTC().UnixMilli()
}

func UnixTimeNow() int64 {
	return time.Now().UTC().UnixMilli()
}

// Ternary operator.
// Returns `a` if `c` is true, `b` otherwise
func Ternary[T any](c bool, a T, b T) T {
	if c {
		return a
	}

	return b
}
