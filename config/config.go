package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// server configs
	ServerHost string `env:"SERVER_HOST,localhost"`
	ServerPort string `env:"SERVER_PORT,8000"`

	// secrets
	SecretKey string `env:"SECRET_KEY,required"` // a key used to hashing and encryption tasks

	// logging configs
	LogLevel string `env:"LOG_LEVEL,info"` // the level of logging
	LogOut   string `env:"LOG_OUT,stdout"` // the output of logging
}

func (c Config) String() (vars string) {
	st := reflect.TypeOf(c)
	for i := range st.NumField() {
		f := st.Field(i)
		values := strings.Split(f.Tag.Get("env"), ",")
		fn, fv := values[0], values[1]
		vars += fmt.Sprintf("%s - %s\n", fn, fv)
	}
	return
}

// LoadFromEnv populates the Config struct fields with environment variables.
// It reads the 'env' tag of each field to determine the environment variable name
// and its default value. If an environment variable is not set and the default value
// is marked as "required", it collects an error message.
func (c *Config) LoadFromEnv() error {
	s := reflect.ValueOf(c)
	st := reflect.TypeOf(*c)
	var errorMsgs []string

	for i := range st.NumField() {
		f := st.Field(i)
		values := strings.Split(f.Tag.Get("env"), ",")
		varName, varDefault := values[0], values[1]

		envValue := os.Getenv(varName)
		if envValue != "" {
			s.Elem().Field(i).SetString(envValue)
			continue
		}
		if varDefault == "required" {
			errorMsgs = append(errorMsgs, fmt.Sprintf("%s is required", varName))
			continue
		}
		s.Elem().Field(i).SetString(varDefault)
	}
	if len(errorMsgs) > 0 {
		return errors.New(strings.Join(errorMsgs, "\n"))
	}
	return nil
}

func (c Config) LoggerLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (c Config) LoggerOut() io.Writer {
	switch c.LogOut {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		return os.Stdout
	}
}

func MustLoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error("couldn't load env vars", "error", err.Error())
		panic(err)
	}
	conf := &Config{}
	if err := conf.LoadFromEnv(); err != nil {
		slog.Error("couldn't load settings", "error", err.Error())
		panic(err)
	}
	return conf
}
