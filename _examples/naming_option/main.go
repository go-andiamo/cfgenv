package main

import (
	"fmt"
	"github.com/go-andiamo/cfgenv"
	"reflect"
	"strings"
)

type DbConfig struct {
	Host     string `env:"optional,default=localhost"`
	Port     uint   `env:"optional,default=3601"`
	Username string
	Password string
}

type Config struct {
	ServiceName string
	Database    DbConfig `env:"prefix=DB"`
}

func main() {
	cfg := &Config{}
	err := cfgenv.Load(cfg, &LowercaseFieldNames{}, cfgenv.NewSeparator("."))
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%+v\n", cfg)
	}
}

type LowercaseFieldNames struct{}

func (l *LowercaseFieldNames) BuildName(prefix string, separator string, fld reflect.StructField, overrideName string) string {
	name := overrideName
	if name == "" {
		name = strings.ToLower(fld.Name)
	}
	if prefix != "" {
		name = prefix + separator + name
	}
	return name
}
