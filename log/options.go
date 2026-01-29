package log

import (
	"fmt"
	"github.com/iWuxc/go-wit/sls"
	jsoniter "github.com/json-iterator/go"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

var (
	_defaultPath = "runtime/logs/"
	json         = jsoniter.ConfigCompatibleWithStandardLibrary
	fieldMap     = logrus.FieldMap{
		logrus.FieldKeyTime:  logFieldTimeStamp,
		logrus.FieldKeyLevel: logFieldLevel,
		logrus.FieldKeyMsg:   logFieldMessage,
	}
)

const (
	DefTime      = "2006-01-02 15:04:05.999"
	defaultLevel = "debug"
)

type Option func(log *logrus.Logger)

func newOptions() []Option {
	return []Option{
		SetOutPutLevel(defaultLevel),
		SetOutFormat(TextFormat()),
	}
}

func SetOutPath(dir string) Option {
	_defaultPath = strings.TrimRight(dir, "/") + "/"
	return func(log *logrus.Logger) {}
}

func SetOutput(name string, maxAgeDay uint32) Option {
	if logPath := os.Getenv("log_path"); len(logPath) > 0 {
		name = strings.TrimRight(logPath, "/") + "/" + name
	}
	if len(name) > 4 && name[len(name)-4:] == ".log" {
		name = name[:len(name)-4]
	}
	var writer *rotatelogs.RotateLogs
	if name[0] == '/' {
		writer, _ = rotatelogs.New(
			name+"_%Y-%m-%d.log",
			rotatelogs.WithMaxAge(time.Hour*24*time.Duration(int64(maxAgeDay))),
			rotatelogs.WithRotationTime(time.Hour*24),
		)
	} else {
		writer, _ = rotatelogs.New(
			_defaultPath+name+"_%Y-%m-%d.log",
			rotatelogs.WithMaxAge(time.Hour*24*time.Duration(int64(maxAgeDay))),
			rotatelogs.WithRotationTime(time.Hour*24),
		)
	}

	return func(log *logrus.Logger) {
		log.AddHook(lfshook.NewHook(lfshook.WriterMap{
			logrus.PanicLevel: writer,
			logrus.FatalLevel: writer,
			logrus.ErrorLevel: writer,
			logrus.WarnLevel:  writer,
			logrus.InfoLevel:  writer,
			logrus.TraceLevel: writer,
			logrus.DebugLevel: writer,
		}, JsonFormat()))
	}
}

func SetOutputWithRotationTime(name string, maxAgeHour uint32, rotationTime time.Duration) Option {
	if logPath := os.Getenv("log_path"); len(logPath) > 0 {
		name = strings.TrimRight(logPath, "/") + "/" + name
	}
	if len(name) > 4 && name[len(name)-4:] == ".log" {
		name = name[:len(name)-4]
	}
	var writer *rotatelogs.RotateLogs
	if name[0] == '/' {
		writer, _ = rotatelogs.New(
			name+"_%Y-%m-%d-%H.log",
			rotatelogs.WithMaxAge(time.Hour*24*time.Duration(int64(maxAgeHour))),
			rotatelogs.WithRotationTime(rotationTime),
		)
	} else {
		writer, _ = rotatelogs.New(
			_defaultPath+name+"_%Y-%m-%d-%H.log",
			rotatelogs.WithMaxAge(time.Hour*time.Duration(int64(maxAgeHour))),
			rotatelogs.WithRotationTime(rotationTime),
		)
	}

	return func(log *logrus.Logger) {
		log.AddHook(lfshook.NewHook(lfshook.WriterMap{
			logrus.PanicLevel: writer,
			logrus.FatalLevel: writer,
			logrus.ErrorLevel: writer,
			logrus.WarnLevel:  writer,
			logrus.InfoLevel:  writer,
			logrus.TraceLevel: writer,
			logrus.DebugLevel: writer,
		}, JsonFormat()))
	}
}

func SetOutPutLevel(level string) Option {
	return func(log *logrus.Logger) {
		log.SetLevel(toLogrusLevel(toLogLevel(level)))
	}
}

func SetOutFormat(format logrus.Formatter) Option {
	return func(log *logrus.Logger) {
		log.SetFormatter(format)
	}
}

// SetOutAliLog 设置日志输出到阿里云日志
func SetOutAliLog() Option {
	return func(log *logrus.Logger) {
		if service := sls.GetAliLogService(); service != nil {
			log.AddHook(service)
		}
	}
}

func JsonFormat() logrus.Formatter {
	format := new(logrus.JSONFormatter)
	format.TimestampFormat = DefTime
	format.FieldMap = fieldMap
	return format
}

func TextFormat() logrus.Formatter {
	format := new(logrus.TextFormatter)
	format.TimestampFormat = DefTime
	format.FieldMap = fieldMap
	return format
}

func CustomFormat() logrus.Formatter {
	format := new(CustomLogFormatter)
	return format
}

type CustomLogFormatter struct{}

func (s *CustomLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Format(DefTime)
	content, _ := json.Marshal(entry.Data)
	msg := fmt.Sprintf("[%s] %s %v\n", timestamp, entry.Message, string(content))
	return []byte(msg), nil
}
