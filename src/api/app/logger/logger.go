package logger

import (
	"kamuy/app"
	"kamuy/app/logger/zap"
)

// New はapp.Loggerを実装したLoggerを返す
func New(c *app.LoggerConfig) (app.Logger, error) {
	return zap.New(c)
}

// NewNop はapp.Loggerを実装した何も出力しない(No Operation)Loggerを返します.
// 主にtestで使うことを想定しています.
func NewNop() app.Logger {
	return newNopLogger()
}

func newNopLogger() *nopLogger {
	return &nopLogger{}
}

type nopLogger struct{}

func (n *nopLogger) Debug(_ string) {}
func (n *nopLogger) Info(_ string)  {}
func (n *nopLogger) Error(_ string) {}
func (n *nopLogger) Warn(_ string)  {}
func (n *nopLogger) Fatal(_ string) {}
func (n *nopLogger) Close() error   { return nil }
