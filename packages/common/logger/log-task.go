package logger

// Designed to be used by worker pool
type logTask struct {
    entry 	*LogEntry
    logger  *FileLogger
	handler logHandler
}

func (t logTask) Process() {
	t.handler(t.entry)
}

func newTaskProducer(logger *FileLogger, handler logHandler) func(*LogEntry) *logTask {
    return func (entry *LogEntry) *logTask {
        return &logTask{entry, logger, handler}
    }
}

