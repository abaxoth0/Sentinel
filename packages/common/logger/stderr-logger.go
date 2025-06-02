package logger

import (
	"log"
	"os"
)

// Satisfies Logger interface
type stderrLogger struct {
    logger *log.Logger
}

func newStderrLogger() *stderrLogger {
    return &stderrLogger{
        // btw log package sends logs into stderr by default (blew my mind)
        // but i want to add prefix to logs and possibility to adjust flags
        logger: log.New(os.Stderr, "ERROR: ", log.Ldate | log.Ltime),
    }
}

func (l *stderrLogger) log(entry *LogEntry) {
    // Must not change behaviour based on log level
    log.Println("["+entry.Source+": "+entry.Level+"] " + entry.Message)
}

func (l *stderrLogger) Log(entry *LogEntry) {
    if ok := logPreprocessing(entry, nil, l.log); !ok {
        return
    }

    l.log(entry)
}

