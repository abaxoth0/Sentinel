package logger

type Logger interface {
    Log(entry *LogEntry)
}

type ConcurrentLogger interface {
    Logger

    Start() error
    Stop()  error
}

// Logger that can transmit logs to another loggers.
type TransmittingLogger interface {
    Logger

    // Binds another logger to this logger.
    // On calling Log() it also will be called on all binded loggers
    // (entry will be the same for all loggers)
    //
    // Can't bind to self. Can't bind to one logger more then once.
    NewTransmission(logger Logger) error
}

// TODO replace "default" with service id
var Default = NewFileLogger("default")

var Undefined = NewSource("UNDEFINED", Default)

var Stdout = newStdoutLogger()

var Stderr = newStderrLogger()

