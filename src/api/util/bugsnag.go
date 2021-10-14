package util

import (
	"fmt"

	"kamuy/infra"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"go.uber.org/zap"
)

const (
	appEnvDev = "development"
)

var notifier *bugsnag.Notifier

// NewBugsnagConfiguration はbugsnag通知設定を返します
func NewBugsnagConfiguration() bugsnag.Configuration {
	return newBugsnagConfiguration(infra.EnvMan.BugsnagKey, infra.EnvMan.AppEnv)
}

func newBugsnagConfiguration(apiKey, releaseStage string) bugsnag.Configuration {
	return bugsnag.Configuration{
		// Your Bugsnag API key, e.g. "c9d60ae4c7e70c4b6c4ebd3e8056d2b8". You can
		// find this by clicking Settings on https://bugsnag.com/.
		APIKey:  apiKey,
		AppType: "backend",
		// The import paths for the Go packages containing your source files
		ProjectPackages: []string{"main", "gsskt_api/**"},
		ReleaseStage:    releaseStage,
		Hostname:        hostname(),
		// 管理画面に表示しないurl parameter
		ParamsFilters: []string{},
		// bigqueryが期待するformatでloggingするためにcustome loggerを生成する
		Logger: newBugsnagLogger(releaseStage),
		// 通知するReleaseStageを指定する
		NotifyReleaseStages: []string{
			"production",
			"staging",
			"testing",
		},
	}
}

// bugsnag管理画面のerror event>deviceにこの値が利用される
// 各人の開発環境の場合はlocalhostとだすようにして、それ以外はbugsnag側に委ねる
func hostname() string {
	if infra.EnvMan.AppEnv == appEnvDev {
		return "localhost"
	}
	return ""
}

// bugsnag packageが要求するlogger
type bugsnagLogger struct {
	*zap.Logger
}

func newBugsnagLogger(appEnv string) logger {
	if appEnv == appEnvDev {
		return &nopBugsnagLogger{}
	}
	_, encode, color := LoggingConfig(appEnv)
	logger, err := NewLogger(
		WithLoggingLevel(1),
		WithEncoded(encode),
		WithColor(color),
		WithAddCaller(0))
	if err != nil {
		panic("failed to initialize middleware bugsnag logger " + err.Error())
	}
	logger = logger.With(zap.String("type", "bugsnag"))
	return &bugsnagLogger{logger}
}

func (l *bugsnagLogger) Printf(format string, v ...interface{}) {
	l.Logger.Warn(fmt.Sprintf(format, v...))
}

type nopBugsnagLogger struct{}

func (n *nopBugsnagLogger) Printf(_ string, _ ...interface{}) {
	// do nothing
}

// bugsnagが要求するloggerのinterface
type logger interface {
	Printf(f string, v ...interface{})
}

func init() {
	notifier = bugsnag.New(NewBugsnagConfiguration())
	if notifier == nil {
		panic("failed to initialize bugsnag notifier")
	}
	notifier.Config.Synchronous = false

}
