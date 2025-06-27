package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerHost string `env:"SERVER_HOST,localhost"`
	ServerPort string `env:"SERVER_PORT,8000"`
	SecretKey  string `env:"SECRET_KEY,required"`
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

func (c Config) LoadFromEnv() error {
	s := reflect.ValueOf(&c)
	st := reflect.TypeOf(c)
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

func MustLoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	return &Config{
		ServerHost: os.Getenv("SERVER_HOST"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}
}
