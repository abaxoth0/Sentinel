package logger

// Wrapper for logger L.
// Strictly bound to the single logger's source.
// Provides more convenient and readable methods for creating logs.
type Source[L Logger] struct {
    logger L
    src    string
}

// Creates a new Source with specified source.
// Will use specified logger to create logs.
func NewSource[L Logger](src string, logger L) *Source[L] {
    return &Source[L]{
        src: src,
        logger: logger,
    }
}

func (s *Source[L]) log(status logLevel, msg string, err string) {
    entry := NewLogEntry(status, s.src, msg, err)
    s.logger.Log(&entry)
}

// Same as L.Log(), but sets status to the InfoLogLevel
func (s *Source[L]) Info(msg string) {
    s.log(InfoLogLevel, msg, "")
}

// Same as L.Log(), but sets status to the WarningLogLevel
func (s *Source[L]) Warning(msg string) {
    s.log(WarningLogLevel, msg, "")
}

// Same as L.Log(), but sets status to the ErrorLogLevel
func (s *Source[L]) Error(msg string, err string) {
    s.log(ErrorLogLevel, msg, err)
}

