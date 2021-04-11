package log

import (
	"context"

	"go.uber.org/zap"
)

type zapLogger struct {
	*zap.SugaredLogger
}

func newZapLogger() zapLogger {
	zlog, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return zapLogger{
		SugaredLogger: zlog.Sugar(),
	}
}

func (l zapLogger) With(v ...interface{}) Logger {
	return zapLogger{
		SugaredLogger: l.SugaredLogger.With(v...),
	}
}

func (l zapLogger) WithContext(ctx context.Context) Logger {
	fields, ok := ctx.Value(FieldsContextKey).([]interface{})
	if !ok {
		return l
	}

	return l.With(fields...)
}
