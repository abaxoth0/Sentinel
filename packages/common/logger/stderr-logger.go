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
        // BTW: log package sends logs into stderr by default
        // but i want to add prefix to logs and possibility to adjust flags
        logger: log.New(os.Stderr, "ERROR: ", log.Ldate | log.Ltime),
    }
}

func (l *stderrLogger) log(entry *LogEntry) {
	msg := "["+entry.Source+": "+entry.Level+"] " + entry.Message
	if entry.rawLevel >= ErrorLogLevel {
		msg += ": " + entry.Error
	}
    l.logger.Println(msg + entry.Meta.stringSuffix())
}

func (l *stderrLogger) Log(entry *LogEntry) {
    if ok := preprocess(entry, nil); !ok {
        return
    }

    l.log(entry)

	if entry.rawLevel >= FatalLogLevel {
		handleCritical(entry)
	}
}
