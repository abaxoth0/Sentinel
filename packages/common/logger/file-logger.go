package logger

import (
	"context"
	"errors"
	"log"
	"os"
	"sentinel/packages/common/structs"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var errLogger = NewSource("LOG", Stderr)

// Satisfies Logger and LoggerBinder interfaces
type FileLogger struct {
    debug         bool
    logger        *log.Logger
    logFile       *os.File
    isRunning     atomic.Bool
    transmissions []Logger
    disruptor     *structs.Disruptor[*LogEntry]
    fallback      *structs.WorkerPool
    taskProducer  func(entry *LogEntry) *logTask
    done          chan struct{}
}

func NewFileLogger(name string) *FileLogger {
    if err := os.MkdirAll("/var/log/sentinel", 0755); err != nil {
        panic("Failed to create log directory: " + err.Error())
    }

    f, err := os.OpenFile(
        "/var/log/sentinel/"+name+".log",
        os.O_APPEND | os.O_CREATE | os.O_WRONLY,
        0644, // -rw-r--r--
    )
    if err != nil {
        panic(err)
    }

    logger := log.New(
        f,
        "",
        log.LstdFlags | log.Lmicroseconds,
    )

    return &FileLogger{
        logger: logger,
        logFile: f,
        transmissions: []Logger{},
        disruptor: structs.NewDisruptor[*LogEntry](),
        fallback: structs.NewWorkerPool(
            context.Background(),
            structs.NewCondWaiter(new(sync.Mutex)),
        ),
        taskProducer: newTaskProducer(logger),
        done: make(chan struct{}),
    }
}

func (l *FileLogger) Start(debug bool) error {
    if l.isRunning.Load() {
        return errors.New("logger already started")
    }

    l.debug = debug

    // canceled WorkerPool can't be started
    if l.fallback.IsCanceled() {
        l.fallback = structs.NewWorkerPool(
            context.Background(),
            structs.NewCondWaiter(new(sync.Mutex)),
        )
    }

    l.isRunning.Store(true)

    go l.disruptor.Consume(newLogEntryHandlerProducer(l.logger))
    go l.fallback.Start(true)

    for {
        select {
        case <-l.done:
            return nil
        default:
            time.Sleep(time.Millisecond * 50)
        }
    }
}

func (l *FileLogger) Stop() error {
    if !l.isRunning.Load() {
        return errors.New("logger isn't started, hence can't be stopped")
    }

    l.isRunning.Store(false)

    l.disruptor.Close()
    if err := l.fallback.Cancel(); err != nil {
        return err
    }
    if err := l.logFile.Close(); err != nil {
        return err
    }

    close(l.done)

    return nil
}

// Creates producer wich will return function that handles log saving.
func newLogEntryHandlerProducer(logger *log.Logger) func(*LogEntry) {
    pool := sync.Pool{
        New: func() any {
            return jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
        },
    }

    return func(entry *LogEntry) {
        stream := pool.Get().(*jsoniter.Stream)
        defer pool.Put(stream)

        stream.Reset(nil)
        stream.Error = nil

        stream.WriteVal(entry)
        if stream.Error != nil {
            errLogger.Error("failed to write log", stream.Error.Error())
            return
        }

        if stream.Buffered() > 0 {
            // Without this all logs will be written in single line
            stream.WriteRaw("\n")
        }

        // NOTE: log.Logger use mutex and atomic operations under the hood,
        //       so it's thread safe by default
        logger.Writer().Write(stream.Buffer())
    }
}

func (l *FileLogger) Log(entry *LogEntry) {
    if entry.rawLevel == DebugLogLevel && l.debug {
        return
    }

    if len(l.transmissions) != 0 {
        defer func() {
            for _, transmission := range l.transmissions {
                transmission.Log(entry)
            }
        }()
    }

    // Immediatly handle panic or fatal log
    if entry.rawLevel >= FatalLogLevel {
        newLogEntryHandlerProducer(l.logger)(entry)

        if entry.rawLevel == PanicLogLevel {
            panic(entry.Message + "\n" + entry.Error)
        }

        // Fatal
        os.Exit(1)
    }

    // if ok is false, that means disruptor's buffer is overflowed
    if ok := l.disruptor.Publish(entry); ok {
        return
    }

    l.fallback.Push(l.taskProducer(entry))
}

func (l *FileLogger) NewTransmission(logger Logger) error {
    if logger == nil {
        return errors.New("received nil instead of logger")
    }

    if logger, ok := logger.(*FileLogger); ok {
        if l == logger {
            return errors.New("can't create transmission for self")
        }
    }

    if slices.Contains(l.transmissions, logger) {
        return errors.New("this logger already has transmission")
    }

    l.transmissions = append(l.transmissions, logger)

    return nil
}

