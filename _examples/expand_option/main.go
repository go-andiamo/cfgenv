package main

import (
	"fmt"
	"github.com/go-andiamo/cfgenv"
)

type Config struct {
	Example string
}

func main() {
	cfg := &Config{}
	err := cfgenv.Load(cfg, cfgenv.Expand())
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%+v\n", cfg)
	}
}
