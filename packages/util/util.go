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
