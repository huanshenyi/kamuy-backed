package app

type EnvType string

const (
	// EnvDevelopment -
	EnvDevelopment EnvType = "development"
	// EnvProto -
	EnvProto EnvType = "proto"
)

func (ev EnvType) String() string {
	return string(ev)
}

var env = EnvDevelopment

// Env -
func Env() EnvType {
	return env
}

func SetEnv(e EnvType) {
	env = e
}
