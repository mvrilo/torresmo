package log

import "context"

type discardLogger struct{}

func (discardLogger) Info(v ...interface{})                    {}
func (discardLogger) Error(v ...interface{})                   {}
func (d discardLogger) With(v ...interface{}) Logger           { return d }
func (d discardLogger) WithContext(ctx context.Context) Logger { return d }

func DiscardLogger() Logger {
	return discardLogger{}
}
