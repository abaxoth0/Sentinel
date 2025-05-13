package logger

import (
	"log"
	"os"
)

// Satisfies Logger interface
type stdoutLogger struct {
    logger *log.Logger
}

func newStdoutLogger() stdoutLogger {
    return stdoutLogger{
        logger: log.New(os.Stdout, "INFO: ", log.Ldate | log.Ltime),
    }
}

func (l stdoutLogger) Log(entry *LogEntry) {
    l.logger.Println("["+entry.Source+": "+entry.Level+"] " + entry.Message)
}

