package zap

import (
	"errors"
	"fmt"
	"io"
	"kamuy/app"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	encodeJSON    = "json"
	encodeConsole = "console"
)

func New(c *app.LoggerConfig) (app.Logger, error) {
	return newLogger(c)
}

func newLogger(c *app.LoggerConfig) (*Logger, error) {
	return newConfig(c).build()
}

func ToAppLogger(z *zap.Logger) *Logger {
	return &Logger{z: z}
}

type Logger struct {
	z *zap.Logger
}

// Debug -
func (l *Logger) Debug(msg string) { l.z.Debug(msg) }

// Info -
func (l *Logger) Info(msg string) { l.z.Info(msg) }

// Warn -
func (l *Logger) Warn(msg string) { l.z.Warn(msg) }

// Error -
func (l *Logger) Error(msg string) { l.z.Error(msg) }

// Fatal -
func (l *Logger) Fatal(msg string) { l.z.Fatal(msg) }

// Close -
func (l *Logger) Close() error {
	return l.z.Sync()
}

// nolint: gocyclo
func newConfig(appCfg *app.LoggerConfig) *config {

	c := &config{}

	// logging levelの設定
	switch appCfg.LoggingLevel {
	case app.LoggingLvlDebug:
		c.level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case app.LoggingLvlInfo:
		c.level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case app.LoggingLvlError:
		c.level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		c.err = fmt.Errorf("unexpected logging level %v", app.LoggingLvlDebug)
	}

	// encodeの設定
	c.encoderCfg = zapcore.EncoderConfig{
		// 以下の値はbigquery側のschemaと対応しているので変更する際はbigquery側のschemaも変更する
		TimeKey:        "timelocal",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	switch appCfg.LoggingEncode {
	case app.LoggingEncJSON:
		c.encode = encodeJSON
	case app.LoggingEncText:
		c.encode = encodeConsole
	case app.LoggingEncColorText:
		c.encode = encodeConsole
		c.encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		c.err = fmt.Errorf("unexpected logging encode %v", appCfg.LoggingEncode)
	}

	// timestampを出力されるとtestできないので
	if appCfg.NoTimestamp {
		c.encoderCfg.TimeKey = ""
	}

	c.out = appCfg.Out
	if c.out == nil {
		c.err = errors.New("app.LoggerConfig.Out should be not nil")
	}

	return c
}

func (c *config) build() (*Logger, error) {
	if c.err != nil {
		return nil, c.err
	}

	encoder, err := c.buildEncoder()
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(c.out), c.level)
	z := zap.New(core, c.options()...)
	return &Logger{
		z: z,
	}, nil
}

type config struct {
	level      zap.AtomicLevel
	encode     string
	encoderCfg zapcore.EncoderConfig

	out io.Writer
	err error
}

func (c *config) buildEncoder() (zapcore.Encoder, error) {
	switch c.encode {
	case "json":
		return zapcore.NewJSONEncoder(c.encoderCfg), nil
	case "console":
		return zapcore.NewConsoleEncoder(c.encoderCfg), nil
	}
	return nil, fmt.Errorf("unexpected logging encode %v", c.encode)
}

func (c *config) options() []zap.Option {
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}
	return options
}
