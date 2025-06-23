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
	msg := "["+entry.Source+": "+entry.Level+"] " + entry.Message
	if entry.rawLevel >= ErrorLogLevel {
		msg += ": " + entry.Error
	}
    l.logger.Println(msg + entry.Meta.stringSuffix())
}

func (l stdoutLogger) Log(entry *LogEntry) {
    if ok := logPreprocessing(entry, nil); !ok {
        return
    }

    l.log(entry)

	if entry.rawLevel >= FatalLogLevel {
		throwError(entry)
	}
}

