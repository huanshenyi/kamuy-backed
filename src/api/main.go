package main

import (
	"fmt"
	"kamuy/app"
	"kamuy/app/logger"
	"kamuy/infra"
	"os"

	"kamuy/util"
)

func main() {
	checkInit()
	setEnv()

	logger := getLogger()
	defer syncLogger(logger)
	logger.Info("Welcome to Kamui API v.01")
}

func checkInit() {
	if infra.EnvMan.Err != nil {
		fmt.Printf("failed to initialize EnvManager: %v", infra.EnvMan.Err)
		os.Exit(1)
	}
}

func setEnv() {
	var env app.EnvType
	switch infra.EnvMan.AppEnv {
	case "dev", "development":
		env = app.EnvDevelopment
	case "proto":
		env = app.EnvProto
	}
	app.SetEnv(env)
}

func getLogger() app.Logger {
	cfg := &app.LoggerConfig{}
	switch app.Env() {

	case app.EnvDevelopment:
		cfg.LoggingLevel = app.LoggingLvlDebug
		cfg.LoggingEncode = app.LoggingEncColorText
	default:
		cfg.LoggingLevel = app.LoggingLvlInfo
		cfg.LoggingEncode = app.LoggingEncJSON
	}
	cfg.Out = os.Stdout
	logger, err := logger.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger", err) // nolint
		os.Exit(2)
	}
	return logger
}

func syncLogger(logger app.Logger) {
	if err := util.Logger.Sync(); err != nil {
		fmt.Println(err)
	}
	if err := logger.Close(); err != nil {
		fmt.Println(err)
	}
}
