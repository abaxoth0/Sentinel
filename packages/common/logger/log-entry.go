package logger

import (
	"time"
)

type logLevel int8

const TraceLogLevel logLevel = -1
const DebugLogLevel logLevel = 0
const InfoLogLevel logLevel = 1
const WarningLogLevel logLevel = 2
const ErrorLogLevel logLevel = 3
// Logs with this level will be handled immediately after calling Log().
// Also os.Exit(1) will be called after log creation.
const FatalLogLevel logLevel = 4
// Logs with this level will be handled immediately after calling Log().
// Also will cause panic after log creation.
const PanicLogLevel logLevel = 5

var logLevelToStrMap = map[logLevel]string{
    TraceLogLevel: "TRACE",
    DebugLogLevel: "DEBUG",
    InfoLogLevel: "INFO",
    WarningLogLevel: "WARNING",
    ErrorLogLevel: "ERROR",
    FatalLogLevel: "FATAL",
    PanicLogLevel: "PANIC",
}

func (s logLevel) String() string{
    return logLevelToStrMap[s]
}

type LogEntry struct {
    Timestamp time.Time `json:"ts"`
    Service   string    `json:"service"`
    Instance  string    `json:"instance"`
    rawLevel  logLevel
    Level     string    `json:"level"`
    Source    string    `json:"source,omitempty"`
    Message   string    `json:"msg"`
    Error     string    `json:"error,omitempty"`
	Meta	  Meta   	`json:"meta,omitempty"`
}

// Creates a new log entry. Timestamp is time.Now().
// If level is not error, fatal or panic, then Error will be empty, even if err specified.
func NewLogEntry(
	level logLevel,
	src string,
	msg string,
	err string,
	meta Meta,
) LogEntry {
    e := LogEntry{
        Timestamp: time.Now(),
        Service: "sentinel",
        rawLevel: level,
        Level: level.String(),
        Source: src,
        Message: msg,
		Meta: meta,
    }

    // error, fatal, panic
    if level >= ErrorLogLevel {
        e.Error = err
    }

    return e
}

