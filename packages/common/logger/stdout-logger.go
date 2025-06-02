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
        logger: log.New(os.Stdout, "", log.Ldate | log.Ltime),
    }
}

func (l stdoutLogger) log(entry *LogEntry) {
    l.logger.Println("["+entry.Source+": "+entry.Level+"] " + entry.Message)
}

func (l stdoutLogger) Log(entry *LogEntry) {
    if ok := logPreprocessing(entry, nil, l.log); !ok {
        return
    }

    l.log(entry)
}

