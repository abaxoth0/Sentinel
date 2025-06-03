package logger

// Wrapper for logger L.
// Strictly bound to the single logger's source.
// Provides more convenient and readable methods for creating logs.
type Source[L Logger] struct {
    logger L
    src    string
}

// Creates a new Source.
// Will use specified logger for logs.
func NewSource[L Logger](src string, logger L) *Source[L] {
    return &Source[L]{
        src: src,
        logger: logger,
    }
}

func (s *Source[L]) log(level logLevel, msg string, err string) {
    entry := NewLogEntry(level, s.src, msg, err)
    s.logger.Log(&entry)
}

// Will create log only if trace logs are enabled,
// Same as L.Log(), but sets level to the TraceLogLevel.
func (s *Source[L]) Trace(msg string) {
    s.log(TraceLogLevel, msg, "")
}

// Will create log only if app running in debug mode,
// Same as L.Log(), but sets level to the DebugLogLevel.
func (s *Source[L]) Debug(msg string) {
    s.log(DebugLogLevel, msg, "")
}

// Same as L.Log(), but sets level to the InfoLogLevel.
func (s *Source[L]) Info(msg string) {
    s.log(InfoLogLevel, msg, "")
}

// Same as L.Log(), but sets level to the WarningLogLevel.
func (s *Source[L]) Warning(msg string) {
    s.log(WarningLogLevel, msg, "")
}

// Same as L.Log(), but sets level to the ErrorLogLevel.
func (s *Source[L]) Error(msg string, err string) {
    s.log(ErrorLogLevel, msg, err)
}

// Same as L.Log(), but sets level to the FatalLogLevel
func (s *Source[L]) Fatal(msg string, err string) {
    s.log(FatalLogLevel, msg, err)
}

// Same as L.Log(), but sets level to the PanicLogLevel
func (s *Source[L]) Panic(msg string, err string) {
    s.log(PanicLogLevel, msg, err)
}

