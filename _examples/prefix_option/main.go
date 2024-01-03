package main

import (
	"fmt"
	"github.com/go-andiamo/cfgenv"
)

type Config struct {
	ServiceName string
}

func main() {
	cfg := &Config{}
	err := cfgenv.Load(cfg, cfgenv.NewPrefix("MYAPP"))
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%+v\n", cfg)
	}
}
