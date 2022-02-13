package tftp

import (
	"go.uber.org/zap"
)

type Logger interface {
	Infof(tmpl string, args ...interface{})
	Debugf(tmpl string, args ...interface{})
	Errorf(tmpl string, args ...interface{})
}

func NewZapLogger() Logger {
	logger, _ := zap.NewProduction()

	return &ZapLogger{
		SugaredLogger: logger.Sugar(),
	}
}

type ZapLogger struct {
	*zap.SugaredLogger
}

func (l *ZapLogger) Infof(tmpl string, args ...interface{}) {
	l.SugaredLogger.Infof(tmpl, args...)
}

func (l *ZapLogger) Debugf(tmpl string, args ...interface{}) {
	l.SugaredLogger.Debugf(tmpl, args...)
}

func (l *ZapLogger) Errorf(tmpl string, args ...interface{}) {
	l.SugaredLogger.Errorf(tmpl, args...)
}
