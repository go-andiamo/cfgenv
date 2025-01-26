package main

import (
	"flag"
	"fmt"
	"github.com/go-andiamo/cfgenv"
	"log"
	"strings"
)

type Config struct {
	Foo int               `env:"optional"`
	All map[string]string `env:"match=.*"`
}

func main() {
	flag.Int("foo", -1, "usage")
	flag.String("bar", "", "usage")
	flag.Parse()

	cfg, err := cfgenv.LoadAs[Config](cfgenv.NewFlagReader(nameConverter{}, true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", cfg)
}

type nameConverter struct{}

func (n nameConverter) ToFlagName(envKey string) string {
	return strings.ToLower(strings.ReplaceAll(envKey, "_", "-"))
}

func (n nameConverter) ToEnvName(flagName string) string {
	return strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}
