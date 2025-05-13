package logger

import "log"

// Designed to be used by worker pool
type logTask struct {
    entry *LogEntry
    logger *log.Logger
}

func (t logTask) Process() {
    newLogEntryHandlerProducer(t.logger)(t.entry)
}

func newTaskProducer(logger *log.Logger) func(*LogEntry) *logTask {
    return func (entry *LogEntry) *logTask {
        return &logTask{entry, logger}
    }
}

