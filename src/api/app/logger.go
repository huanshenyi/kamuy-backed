package app

import "io"

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Close() error
}

// LoggingLevel はloggingのsuppressするlevelを表します.
// 指定されたlevel未満のlogは出力されません.
type LoggingLevel int

const (
	// LoggingLvlDebug 開発時のdebug等を想定
	LoggingLvlDebug LoggingLevel = iota - 1
	// LoggingLvlInfo stagingで出力されることを想定
	LoggingLvlInfo
	// LoggingLvlError productionで出力されることを想定
	LoggingLvlError
)

// LoggingEncode はloggingのencode方法を表します.
type LoggingEncode string

const (
	// LoggingEncJSON はjsonのencodeを表します. staging/productionではjsonを指定してください.
	LoggingEncJSON LoggingEncode = "json"
	// LoggingEncText はtextのencodeを表します.
	LoggingEncText LoggingEncode = "text"
	// LoggingEncColorText はtextのencodeを表します. escape sequenceで色情報を付加します.
	LoggingEncColorText LoggingEncode = "color"
)

// LoggerConfig はlogger生成のための設定を表します.
type LoggerConfig struct {
	LoggingLevel  LoggingLevel
	LoggingEncode LoggingEncode
	// logの出力先
	Out io.Writer
	// Timestampの出力制御
	NoTimestamp bool
}
