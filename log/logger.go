package log

import "context"

type Logger interface {
	WithContext(ctx context.Context) Logger
	With(v ...interface{}) Logger
	Info(v ...interface{})
	Error(v ...interface{})
}

var FieldsContextKey struct{}

var DefaultLogger = NewLogger()

func NewLogger() Logger {
	return newZapLogger()
}

func WithContext(ctx context.Context) Logger {
	return DefaultLogger.WithContext(ctx)
}

func With(v ...interface{}) Logger {
	return DefaultLogger.With(v...)
}

func Info(v ...interface{}) {
	DefaultLogger.Info(v...)
}

func Error(v ...interface{}) {
	DefaultLogger.Error(v...)
}
