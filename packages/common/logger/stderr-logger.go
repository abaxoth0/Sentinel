package logger

import (
	"log"
	"os"
)

// Satisfies Logger interface
type stderrLogger struct {
    logger *log.Logger
}

func newStderrLogger() stderrLogger {
    return stderrLogger{
        // btw log package sends logs into stderr by default (blew my mind)
        // but i want to add prefix to logs and possibility to adjust flags
        logger: log.New(os.Stderr, "ERROR: ", log.Ldate | log.Ltime),
    }
}

func (l stderrLogger) Log(entry *LogEntry) {
    log.Println("["+entry.Source+": "+entry.Level+"] " + entry.Message)
}


