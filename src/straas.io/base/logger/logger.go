package logger

import (
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	logger = newLogger()
)

// Get return logger
func Get() Logger {
	return logger
}

// SetLevel set the log level
func SetLevel(level string) error {
	lv, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logger.Level = lv
	return nil
}

func newLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		// DisableColors: true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	}
	return logger
}

// func AddPrefix(prefix string, Logger) Logger {
//
// }

// Logger defines an interface for log
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}
