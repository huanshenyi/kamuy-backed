package infra

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var EnvMan envManager

type envManager struct {
	AppEnv     string `validate:"required"`
	ZapLevel   string
	BugsnagKey string

	Err error
}

func init() {
	EnvMan = newEnvManager()
}

func newEnvManager() envManager {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}
	var em envManager
	em = envManager{
		AppEnv:     os.Getenv("APP_ENV"),
		ZapLevel:   os.Getenv("ZAP_LEVEL"),
		BugsnagKey: os.Getenv("BUGSNAG_KEY"),
	}
	em.Err = validator.New().Struct(&em)
	return em
}
