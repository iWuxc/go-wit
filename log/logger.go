package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
)

const (
	logFieldTimeStamp = "time"
	logFieldLevel     = "level"
	logFieldMessage   = "msg"
)

// LogLevel .
type LogLevel string

const (
	// DebugLevel has verbose message.
	DebugLevel LogLevel = "debug"
	// InfoLevel is default log level.
	InfoLevel LogLevel = "info"
	// WarnLevel is for logging messages about possible issues.
	WarnLevel LogLevel = "warn"
	// ErrorLevel is for logging errors.
	ErrorLevel LogLevel = "error"
	// FatalLevel is for logging fatal messages. The system shuts down after logging the message.
	FatalLevel LogLevel = "fatal"
)

var (
	globalLoggers     = map[string]Logger{}
	globalLoggersLock = sync.RWMutex{}
)

// Logger includes the logging api sets.
type Logger interface {
	// Out .
	Out() io.Writer
	// WithFields .
	WithFields(map[string]interface{}) Logger
	// WithField .
	WithField(key string, value interface{}) Logger
	// WithError .
	WithError(err error) Logger
	// WithContext .
	WithContext(ctx context.Context) Logger
	// WithLevel .
	WithLevel(level LogLevel) Logger
	// Printf logs a message at level Info.
	Printf(format string, args ...interface{})
	// Info logs a message at level Info.
	Info(args ...interface{})
	// Infof logs a message at level Info.
	Infof(format string, args ...interface{})
	// Debug logs a message at level Debug.
	Debug(args ...interface{})
	// Debugf logs a message at level Debug.
	Debugf(format string, args ...interface{})
	// Warn logs a message at level Warn.
	Warn(args ...interface{})
	// Warnf logs a message at level Warn.
	Warnf(format string, args ...interface{})
	// Error logs a message at level Error.
	Error(args ...interface{})
	// Errorf logs a message at level Error.
	Errorf(format string, args ...interface{})
	// Fatal logs a message at level Fatal then the process will exit with status set to 1.
	Fatal(args ...interface{})
	// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
	Fatalf(format string, args ...interface{})
}

func toLogrusLevel(lvl LogLevel) logrus.Level {
	// ignore error because it will never happen.
	l, _ := logrus.ParseLevel(string(lvl))
	return l
}

// toLogLevel converts to LogLevel.
func toLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// NewLogger creates new Logger instance.
func NewLogger(name string, opts ...Option) Logger {
	globalLoggersLock.Lock()
	defer globalLoggersLock.Unlock()

	logger, ok := globalLoggers[name]
	if !ok {
		logger = newLog(name, opts...)
		globalLoggers[name] = logger
	}

	return logger
}

// SetLogger replace Logger instance.
func SetLogger(name string, logger Logger) Logger {
	globalLoggersLock.Lock()
	defer globalLoggersLock.Unlock()

	l, ok := globalLoggers[name]
	if ok {
		globalLoggers[name] = logger
	}

	return l
}

func GetLogger(name string) Logger {
	return globalLoggers[name]
}

func getLoggers() map[string]Logger {
	globalLoggersLock.RLock()
	defer globalLoggersLock.RUnlock()

	l := map[string]Logger{}
	for k, v := range globalLoggers {
		l[k] = v
	}

	return l
}
