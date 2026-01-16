package logger

import (
	"os"
	"sync/atomic"
)

var Debug atomic.Bool
var Trace atomic.Bool

type Meta map[string]any

func (m Meta) stringSuffix() string {
	if m == nil {
		return ""
	}

	return createSuffixFromWellKnownMeta(m)
}

var wellKnownMetaProperties = []string{
	"addr",
	"method",
	"path",
	"user_agent",
	"request_id",
}

func createSuffixFromWellKnownMeta(m Meta) string {
	s := ""
	for _, prop := range wellKnownMetaProperties {
		if v, ok := m[prop].(string); ok {
			s += v + " "
		}
	}
	if s == "" {
		return ""
	}
	return " ("+s+")"
}


type Logger interface {
	Log(entry *LogEntry)
	// Just logs specified entry.
	// This method mustn't cause any side effects and mostly required for ForwardingLogger.
	// e.g. entry with panic level won't cause panic when
	// forwarded to another logger, only when main logger will handle it
	log(entry *LogEntry)
}

type ConcurrentLogger interface {
	Logger

	Start() error
	Stop() error
}

// Logger that can forward logs to another loggers.
type ForwardingLogger interface {
	Logger

	// Binds another logger to this logger.
	// On calling Log() it also will be called on all binded loggers
	// (entry will be the same for all loggers)
	//
	// Can't bind to self. Can't bind to one logger more then once.
	NewForwarding(logger Logger) error

	// Removes existing forwarding.
	// Will return error if forwading to specified logger isn't exist.
	RemoveForwarding(logger Logger) error
}

// Returns false if log must not be processed
func preprocess(entry *LogEntry, forwardings []Logger) bool {
	if entry.rawLevel == DebugLogLevel && !Debug.Load() {
		return false
	}

	if entry.rawLevel == TraceLogLevel && !Trace.Load() {
		return false
	}

	if forwardings != nil && len(forwardings) != 0 {
		for _, forwardings := range forwardings {
			// Must call log() not Log(), since log() just doing logging
			// without any additional side effects.
			forwardings.log(entry)
		}
	}

	return true
}

// If log entry rawLevel is:
//   - FatalLogLevel: will call os.Exit(1)
//   - PanicLogLevel: will cause panic with entry.Message and entry.Error
func handleCritical(entry *LogEntry) {
	if entry.rawLevel == PanicLogLevel {
		panic(entry.Message + "\n" + entry.Error)
	}
	os.Exit(1)
}

var Default = NewFileLogger("default")

// Refers logger.Default
var Undefined = NewSource("UNDEFINED", Default)

var Stdout = newStdoutLogger()

var Stderr = newStderrLogger()
