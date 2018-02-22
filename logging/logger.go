package logging

import (
	"github.com/sirupsen/logrus"
	"runtime"
)

var baseLogger = logrus.New()

type AnonymousMessagingLogger struct {
	*logrus.Entry
}

func (l *AnonymousMessagingLogger) WithField(key string, value interface{}) *AnonymousMessagingLogger {
	return &AnonymousMessagingLogger{l.Entry.WithField(key, value)}
}

func (l *AnonymousMessagingLogger) WithFields(fields logrus.Fields) *AnonymousMessagingLogger {
	return &AnonymousMessagingLogger{l.Entry.WithFields(fields)}
}

func PackageLogger() *AnonymousMessagingLogger {
	_, filename, _, _ := runtime.Caller(1)
	return &AnonymousMessagingLogger{baseLogger.WithField("prefix", filename)}
}

func PackageLoggerWithField(key, value string) *AnonymousMessagingLogger {
	return &AnonymousMessagingLogger{baseLogger.WithField(key, value)}
}
