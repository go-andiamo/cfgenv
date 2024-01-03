package main

import (
	"fmt"
	"github.com/go-andiamo/cfgenv"
	"os"
)

type Config struct {
	CustomValue Custom
}

func main() {
	_ = os.Setenv("CUSTOM_VALUE", "this will be decoded as Custom (i.e. []byte)")
	cfg := &Config{}
	err := cfgenv.Load(cfg, &CustomSetter{})
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%+v\n", cfg)
		fmt.Printf("CustomValue: %s\n", string(cfg.CustomValue))
	}
}
