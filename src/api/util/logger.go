package util

import (
	"errors"
	"strconv"
	"strings"

	"kamuy/infra"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// LoggingEncode -
	LoggingEncode string
	// LoggingLevel -
	LoggingLevel int
)

const (
	// LoggingEncodeJSON 運用時の指定を想定
	LoggingEncodeJSON LoggingEncode = "json"
	// LoggingEncodeConsole 開発時の指定を想定
	LoggingEncodeConsole LoggingEncode = "console"
)

const (
	// LoggingLevelDebug 本運用では出力されない
	LoggingLevelDebug LoggingLevel = iota - 1
	// LoggingLevelInfo 運用時の分析指標を提供
	LoggingLevelInfo
	// LoggingLevelWarn 運用担当者が定期的に集計して対応
	LoggingLevelWarn
	// LoggingLevelError bugsnag等に通知が飛ぶ
	LoggingLevelError
	// LoggingLevelFatal 使わない想定
	LoggingLevelFatal
)

// Logger -
var Logger *zap.Logger

// LoggerOption は logger設定用の関数です
type LoggerOption func(*zap.Config, *[]zap.Option) error

// WithLoggingLevel は loggingのlevelを制御します
func WithLoggingLevel(level LoggingLevel) LoggerOption {
	return func(cfg *zap.Config, _ *[]zap.Option) error {
		cfg.Level = toZapLevel(level)
		return nil
	}
}

// WithEncoded は loggingの出力方法を指定します
func WithEncoded(encode LoggingEncode) LoggerOption {
	return func(cfg *zap.Config, _ *[]zap.Option) error {
		cfg.Encoding = string(encode)
		return nil
	}
}

// WithColor は loggingのmessageに色をつけるかを制御します
// encodeにterminalを指定した際には見やすくなりますが、json時には制御コードが出力されてしまいます
func WithColor(color bool) LoggerOption {
	return func(cfg *zap.Config, _ *[]zap.Option) error {
		if color {
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		return nil
	}
}

// WithAddStacktrace はloggingにstacktrace情報を追加します
func WithAddStacktrace(level LoggingLevel) LoggerOption {
	return func(_ *zap.Config, opts *[]zap.Option) error {
		*opts = append(*opts, zap.AddStacktrace(toZapLevel(level)))
		return nil
	}
}

// WithAddCaller は loggingにcaller情報を追加します
// skip数だけcallerを遡って表示します
func WithAddCaller(skip int) LoggerOption {
	return func(_ *zap.Config, opts *[]zap.Option) error {
		*opts = append(*opts, zap.AddCaller())
		*opts = append(*opts, zap.AddCallerSkip(skip))
		return nil
	}
}

// WithBugsnagNotify は Error logging時にbugsnagへの通知hookを追加します
func WithBugsnagNotify(level LoggingLevel) LoggerOption {
	l := zapcore.Level(level)
	cb := func(entry zapcore.Entry) error {
		if entry.Level >= l {
			return notifier.Notify(
				errors.New(entry.Message),
				bugsnag.HandledState{
					SeverityReason:   bugsnag.SeverityReasonHandledError,
					OriginalSeverity: bugsnag.SeverityError,
					Unhandled:        false,
					Framework:        "Gin",
				})
		}
		return nil
	}

	return withCallback(cb)
}

func withCallback(f func(zapcore.Entry) error) LoggerOption {
	return func(_ *zap.Config, opts *[]zap.Option) error {
		*opts = append(*opts, zap.Hooks(f))
		return nil
	}
}

// NewLogger returns zap logger
func NewLogger(options ...LoggerOption) (*zap.Logger, error) {
	cfg := &zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.Level(int8(-1))),
		Development:      true,
		Encoding:         "console", // or json
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
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
		},
	}

	var buildOpts = []zap.Option{}
	for _, opt := range options {
		if err := opt(cfg, &buildOpts); err != nil {
			return nil, err
		}
	}

	logger, err := cfg.Build(buildOpts...)
	if err != nil {
		return nil, err
	}
	return logger, nil
}

// LoggingConfig returns basic logging configuration from app env
func LoggingConfig(appEnv string) (level LoggingLevel, encode LoggingEncode, color bool) {
	switch v := strings.ToLower(appEnv); v {
	case "dev", "development":
		level, encode, color = LoggingLevelDebug, LoggingEncodeConsole, true
	case "stg", "staging":
		level, encode, color = LoggingLevelDebug, LoggingEncodeJSON, false
	case "prod", "production":
		level, encode, color = LoggingLevelInfo, LoggingEncodeJSON, false
	default:
		level, encode, color = LoggingLevelDebug, LoggingEncodeJSON, false
	}
	return
}

func toZapLevel(level LoggingLevel) zap.AtomicLevel {
	return zap.NewAtomicLevelAt(zapcore.Level(int8(level)))
}

// fluentdのためのmeta情報を付加する
// https://github.com/howtv/gsskt_backend/issues/631
func initLogger() {
	Logger = Logger.With(zap.String("type", "app"))
}

func init() {
	level, encode, color := LoggingConfig(infra.EnvMan.AppEnv)

	// zap levelが明示的に指定されていればそちらに従う
	if v := infra.EnvMan.ZapLevel; v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			level = LoggingLevel(l)
		}
	}

	var err error
	Logger, err = NewLogger(
		WithLoggingLevel(level),
		WithEncoded(encode),
		WithColor(color),
		WithAddStacktrace(LoggingLevelError),
		//WithBugsnagNotify(LoggingLevelError),
		WithAddCaller(0))
	if err != nil {
		panic(err)
	}

	initLogger()
}
