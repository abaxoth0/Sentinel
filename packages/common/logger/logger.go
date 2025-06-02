package logger

import (
	"os"
	"sync/atomic"
)

var Debug atomic.Bool
var Trace atomic.Bool

type Logger interface {
    Log(entry *LogEntry)
    // Just logs entry, ignoring its content.
    // This method is mostly required for TransmittingLogger.
    // (e.g. entry with panic level won't cause panic)
    log(entry *LogEntry)
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

    // Removes existing transition.
    // Will return error if transmission to specified logger isn't exist.
    RemoveTransmission(logger Logger) error
}

type logHandler = func (*LogEntry)

func logPreprocessing(
    entry *LogEntry,
    transmissions []Logger,
    handler logHandler,
) bool {
    if entry.rawLevel == DebugLogLevel && !Debug.Load() {
        return false
    }

    if entry.rawLevel == TraceLogLevel && !Trace.Load() {
        return false
    }

    if transmissions != nil && len(transmissions) != 0 {
        defer func() {
            for _, transmission := range transmissions {
                // Must call log() not Log(), since log() just doing logging
                // without any additional side effects.
                // Also log() won't cause recursive transmissions.
                // (cuz transmissions handled at Log())
                transmission.log(entry)
            }
        }()
    }

    // Immediatly handle panic or fatal log
    if entry.rawLevel >= FatalLogLevel {
        handler(entry)

        if entry.rawLevel == PanicLogLevel {
            panic(entry.Message + "\n" + entry.Error)
        }

        // Fatal
        os.Exit(1)
    }

    return true
}

// TODO replace "default" with service id
var Default = NewFileLogger("default")

// Refers logger.Default
var Undefined = NewSource("UNDEFINED", Default)

var Stdout = newStdoutLogger()

var Stderr = newStderrLogger()

