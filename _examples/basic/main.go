package main

import (
	"github.com/go-andiamo/cfgenv"
	"os"
)

type Config struct {
	Connection string
}

func main() {
	cfg := &Config{}

	err := cfgenv.Load(cfg)
	if err != nil {
		println("Error:", err.Error())
	} else {
		println("Connection: ", cfg.Connection)
	}

	setupEnv(map[string]string{"CONNECTION": "localhost:3601"})
	err = cfgenv.Load(cfg)
	if err != nil {
		println("Error:", err.Error())
	} else {
		println("Connection: ", cfg.Connection)
	}
}

func setupEnv(envs map[string]string) {
	os.Clearenv()
	for k, v := range envs {
		_ = os.Setenv(k, v)
	}
}
