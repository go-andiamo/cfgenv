package main

import (
	"fmt"
	"github.com/go-andiamo/cfgenv"
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
	err := cfgenv.Load(cfg, cfgenv.NewPrefix("MYAPP"), cfgenv.NewSeparator("."))
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%+v\n", cfg)
	}
}
