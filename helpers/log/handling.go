package log

import (
	"github.com/sirupsen/logrus"
	"time"
)

// logging handler

type ILoggingHandler interface {
	Tracef(fmt string, i ...interface{})
	Debugf(fmt string, i ...interface{})
	Infof(fmt string, i ...interface{})
	Warningf(fmt string, i ...interface{})
	Errorf(fmt string, i ...interface{})
	Fatalf(fmt string, i ...interface{})
	Panicf(fmt string, i ...interface{})
	Exceptionf(fmt string, i ...interface{})
}

type LoggingHandler logrus.Entry

func (l *LoggingHandler) Exceptionf(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Errorf(fmt, i...)
	panic("logging handler exception")
}

func (l *LoggingHandler) Tracef(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Tracef(fmt, i...)
}

func (l *LoggingHandler) Debugf(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Debugf(fmt, i...)
}

func (l *LoggingHandler) Infof(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Infof(fmt, i...)
}

func (l *LoggingHandler) Warningf(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Warningf(fmt, i...)
}

func (l *LoggingHandler) Errorf(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Errorf(fmt, i...)
}

func (l *LoggingHandler) Fatalf(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Fatalf(fmt, i...)
}

func (l *LoggingHandler) Panicf(fmt string, i ...interface{}) {
	entry := (logrus.Entry)(*l)
	(&entry).Panicf(fmt, i...)
}

func Field(key string, value interface{}) ILoggingHandler {
	return (*LoggingHandler)(logrus.WithField(key, value))
}
func Fields(m map[string]interface{}) ILoggingHandler {
	return (*LoggingHandler)(logrus.WithFields(m))
}
func Now() ILoggingHandler {
	return (*LoggingHandler)(logrus.WithTime(time.Now()))
}

var _ ILoggingHandler = &LoggingHandler{}
