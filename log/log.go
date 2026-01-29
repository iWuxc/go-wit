package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

const globalName = "common"

var global = NewLogger(globalName, newOptions()...)

// Log is the implementation for logrus.
type Log struct {
	// name is the name of logger .
	name *string
	// logger.
	logger *logrus.Logger
	entry  *logrus.Entry
}

func newLog(name string, opts ...Option) Logger {
	newLogger := logrus.New()
	newLogger.SetOutput(os.Stdout)
	for _, o := range opts {
		o(newLogger)
	}

	dl := &Log{
		name:   &name,
		logger: newLogger,
	}
	return dl
}

func Printf(format string, args ...interface{}) {
	global.Infof(format, args...)
}

func Info(args ...interface{}) {
	global.Info(args...)
}

func (l *Log) Printf(format string, args ...interface{}) {
	if l.entry == nil {
		l.logger.Printf(format, args...)
	} else {
		l.entry.Printf(format, args...)
	}
}

func (l *Log) Info(args ...interface{}) {
	if l.entry == nil {
		l.logger.Info(args...)
	} else {
		l.entry.Info(args...)
	}
}

func (l *Log) Out() io.Writer {
	return l.logger.Out
}

func Infof(format string, args ...interface{}) {
	global.Infof(format, args...)
}

func (l *Log) Infof(format string, args ...interface{}) {
	if l.entry == nil {
		l.logger.Infof(format, args...)
	} else {
		l.entry.Infof(format, args...)
	}
}

func Debug(args ...interface{}) {
	global.Debug(args...)
}

func (l *Log) Debug(args ...interface{}) {
	if l.entry == nil {
		l.logger.Debug(args...)
	} else {
		l.entry.Debug(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	global.Debugf(format, args...)
}

func (l *Log) Debugf(format string, args ...interface{}) {
	if l.entry == nil {
		l.logger.Debugf(format, args...)
	} else {
		l.entry.Debugf(format, args...)
	}
}

func Warn(args ...interface{}) {
	global.Warn(args...)
}

func (l *Log) Warn(args ...interface{}) {
	if l.entry == nil {
		l.logger.Warn(args...)
	} else {
		l.entry.Warn(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	global.Warnf(format, args...)
}

func (l *Log) Warnf(format string, args ...interface{}) {
	if l.entry == nil {
		l.logger.Warnf(format, args...)
	} else {
		l.entry.Warnf(format, args...)
	}
}

func Error(args ...interface{}) {
	global.Error(args...)
}

func (l *Log) Error(args ...interface{}) {
	if l.entry == nil {
		l.logger.Error(args...)
	} else {
		l.entry.Error(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	global.Errorf(format, args...)
}

func (l *Log) Errorf(format string, args ...interface{}) {
	if l.entry == nil {
		l.logger.Errorf(format, args...)
	} else {
		l.entry.Errorf(format, args...)
	}
}

func Fatal(args ...interface{}) {
	global.Fatal(args...)
}

func (l *Log) Fatal(args ...interface{}) {
	if l.entry == nil {
		l.logger.Fatal(args...)
	} else {
		l.entry.Fatal(args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	global.Fatalf(format, args...)
}

func (l *Log) Fatalf(format string, args ...interface{}) {
	if l.entry == nil {
		l.logger.Fatalf(format, args...)
	} else {
		l.entry.Fatalf(format, args...)
	}
}

func WithFields(fields map[string]interface{}) Logger {
	return global.WithFields(fields)
}

func (l *Log) WithFields(fields map[string]interface{}) Logger {
	log := &Log{
		logger: l.logger,
		name:   l.name,
	}
	if l.entry == nil {
		log.entry = log.logger.WithFields(fields)
	} else {
		log.entry = l.entry.WithFields(fields)
	}
	return log
}

func WithField(key string, value interface{}) Logger {
	return global.WithField(key, value)
}

func (l *Log) WithField(key string, value interface{}) Logger {
	log := &Log{
		logger: l.logger,
		name:   l.name,
	}
	if l.entry == nil {
		log.entry = log.logger.WithField(key, value)
	} else {
		log.entry = l.entry.WithField(key, value)
	}
	return log
}

func WithContext(ctx context.Context) Logger {
	return global.WithContext(ctx)
}

func (l *Log) WithContext(ctx context.Context) Logger {

	log := &Log{
		logger: l.logger,
		name:   l.name,
	}

	if l.entry == nil {
		log.entry = log.logger.WithContext(ctx)
	} else {
		log.entry = l.entry.WithContext(ctx)
	}
	return log
}

func WithError(err error) Logger {
	return global.WithError(err)
}

func (l *Log) WithError(err error) Logger {
	log := &Log{
		logger: l.logger,
		name:   l.name,
	}
	if l.entry == nil {
		log.entry = log.logger.WithError(err)
	} else {
		log.entry = l.entry.WithError(err)
	}
	return log
}

func (l *Log) WithHook(hook logrus.Hook) Logger {
	if hook != nil {
		l.logger.AddHook(hook)
	}
	return l
}

func WithLevel(level LogLevel) Logger {
	return global.WithLevel(level)
}

func (l *Log) WithLevel(level LogLevel) Logger {
	lev, _ := logrus.ParseLevel(string(level))
	l.logger.SetLevel(lev)
	return l
}

func GetInstance() Logger {
	return global
}

func ReplaceLogger(logger Logger) Logger {
	global = logger
	SetLogger(globalName, logger)
	return logger
}
