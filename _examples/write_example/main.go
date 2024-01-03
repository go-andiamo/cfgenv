package main

import (
	"bytes"
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

	// show example...
	w := &bytes.Buffer{}
	if err := cfgenv.Example(w, cfg, cfgenv.NewPrefix("MYAPP")); err == nil {
		println("Example...")
		println(w.String())
	} else {
		panic(err)
	}

	// load the config...
	if err := cfgenv.Load(cfg, cfgenv.NewPrefix("MYAPP")); err != nil {
		panic(err)
	}
	// write current config...
	w = &bytes.Buffer{}
	if err := cfgenv.Write(w, cfg, cfgenv.NewPrefix("MYAPP")); err == nil {
		println("Current...")
		println(w.String())
	} else {
		panic(err)
	}
}
