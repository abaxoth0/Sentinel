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

// Used for intenal logging (mostly for errors).
// Send logs in stderr
var fileLog = NewSource("FILE LOGGER", Stderr)

const (
	fallbackBatchSize = 500
	fallbackWorkers   = 5
)

// Satisfies Logger and LoggerBinder interfaces
type FileLogger struct {
	name 				 string
	instance 			 string
	isInit 				 bool
    done                 chan struct{}
    isRunning            atomic.Bool
    disruptor            *structs.Disruptor[*LogEntry]
    fallback             *structs.WorkerPool
    logger               *log.Logger
    logFile              *os.File
    transmissions        []Logger
    taskProducer         func(entry *LogEntry) *logTask
	streamPool			 sync.Pool
}

func NewFileLogger(name string) *FileLogger {
    if err := os.MkdirAll("/var/log/sentinel", 0755); err != nil {
        panic("Failed to create log directory: " + err.Error())
    }

	logger := &FileLogger{
		name: name,
        done: make(chan struct{}),
        disruptor: structs.NewDisruptor[*LogEntry](),
        fallback: structs.NewWorkerPool(context.Background(), fallbackBatchSize),
        transmissions: []Logger{},
		streamPool: sync.Pool{
			New: func() any {
				return jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
			},
		},
    }
	logger.taskProducer = newTaskProducer(logger)

	return logger
}

func (l *FileLogger) Init(instance string) {
	// file with logs have following format: <logger name>:<app instance>:<session start date>
	fileName := l.name+":"+instance+":"+time.Now().Format(time.RFC3339)+".log"

    f, err := os.OpenFile(
		"/var/log/sentinel/" + fileName,
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

	l.instance = instance
	l.logger = logger
	l.logFile = f
	l.taskProducer = newTaskProducer(l)
	l.isInit = true
}

func (l *FileLogger) Start(debug bool) error {
	if !l.isInit {
		return errors.New("logger isn't initialized")
	}

    if l.isRunning.Load() {
        return errors.New("logger already started")
    }

    // canceled WorkerPool can't be started
    if l.fallback.IsCanceled() {
        l.fallback = structs.NewWorkerPool(context.Background(), fallbackBatchSize)
    }

    l.isRunning.Store(true)

    go l.disruptor.Consume(l.handler)
    go l.fallback.Start(fallbackWorkers)

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

func (l *FileLogger) handler(entry *LogEntry) {
	stream := l.streamPool.Get().(*jsoniter.Stream)
	defer l.streamPool.Put(stream)

	stream.Reset(nil)
	stream.Error = nil

	entry.Instance = l.instance

	stream.WriteVal(entry)
	if stream.Error != nil {
		fileLog.Error("failed to write log", stream.Error.Error(), nil)
		return
	}

	if stream.Buffered() > 0 {
		// Without this all logs will be written in single line
		stream.WriteRaw("\n")
	}

	// NOTE: log.Logger use mutex and atomic operations under the hood,
	//       so it's thread safe by default
	l.logger.Writer().Write(stream.Buffer())
}

func (l *FileLogger) log(entry *LogEntry) {
    // if ok is false, that means disruptor's buffer is overflowed
    if ok := l.disruptor.Publish(entry); ok {
        return
    }

    l.fallback.Push(l.taskProducer(entry))
}

func (l *FileLogger) Log(entry *LogEntry) {
    if !preprocess(entry, l.transmissions) {
		return
    }

    l.log(entry)

	if entry.rawLevel >= FatalLogLevel {
		handleCritical(entry)
	}
}

func (l *FileLogger) NewTransmission(logger Logger) error {
    if logger == nil {
        return errors.New("received nil instead of logger")
    }

    if l == logger {
        return errors.New("can't create transmission for self")
    }

    if slices.Contains(l.transmissions, logger) {
        return errors.New("this logger already has transmission")
    }

    l.transmissions = append(l.transmissions, logger)

    return nil
}

func (l *FileLogger) RemoveTransmission(logger Logger) error {
    if logger == nil {
        return errors.New("received nil instead of logger")
    }

    for idx, transmission := range l.transmissions {
        if transmission == logger {
            l.transmissions = slices.Delete(l.transmissions, idx, idx+1)
            return nil
        }
    }

    return errors.New("transmission now found")
}

