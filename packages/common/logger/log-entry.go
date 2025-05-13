package logger

import (
	"time"
)

type logLevel uint8

const InfoLogLevel logLevel = 0
const WarningLogLevel logLevel = 1
const ErrorLogLevel logLevel = 2

var logLevelToStrMap = map[logLevel]string{
    InfoLogLevel: "INF",
    WarningLogLevel: "WRN",
    ErrorLogLevel: "ERR",
}

func (s logLevel) String() string{
    return logLevelToStrMap[s]
}

type LogEntry struct {
    Timestamp time.Time `json:"ts"`
    Service   string    `json:"service"`
    Instance  string    `json:"instance"`
    Level     string    `json:"level"`
    Source    string    `json:"source,omitempty"`
    Message   string    `json:"msg"`
    Error     string    `json:"error,omitempty"`
}

// Creates a new log entry. Timestamp is time.Now()
func NewLogEntry(level logLevel, src string, msg string, err string) LogEntry {
    return LogEntry{
        Timestamp: time.Now(),
        Service: "sentinel",
        Instance: "default", // TODO replace "default" with service id
        Level: level.String(),
        Source: src,
        Message: msg,
        Error: err,
    }
}

