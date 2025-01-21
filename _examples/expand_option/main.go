package main

import (
	"fmt"
	"github.com/go-andiamo/cfgenv"
	"os"
)

type Config struct {
	Example string
}

func main() {
	f, err := os.Open("_examples/expand_option/example.env")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	cfg := &Config{}
	err = cfgenv.Load(cfg, cfgenv.Expand(), cfgenv.NewEnvFileReader(f, nil))
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%+v\n", cfg)
	}
}
