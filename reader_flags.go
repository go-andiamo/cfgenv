package cfgenv

import (
	"flag"
)

// FlagNameConverter is an interface that can be used with NewFlagReader
// and converts env var names to flag names or vice versa
type FlagNameConverter interface {
	ToFlagName(envKey string) string
	ToEnvName(flagName string) string
}

type flagReader struct {
	fs            *flag.FlagSet
	nameConverter FlagNameConverter
	useDefaults   bool
}

// NewFlagReader creates a new EnvReader that reads from flags (i.e. flag.CommandLine)
//
// If flag names differ from env var names pass a nameConverter arg
// or nil if no name conversion is needed
//
// The useDefaults args determines whether flag defaults are used
func NewFlagReader(nameConverter FlagNameConverter, useDefaults bool) EnvReader {
	return NewFlagSetReader(flag.CommandLine, nameConverter, useDefaults)
}

// NewFlagSetReader creates a new EnvReader that reads from flags from the provided *flag.FlagSet
//
// If flag names differ from env var names pass a nameConverter arg
// or nil if no name conversion is needed
//
// The useDefaults args determines whether flag defaults are used
func NewFlagSetReader(fs *flag.FlagSet, nameConverter FlagNameConverter, useDefaults bool) EnvReader {
	return &flagReader{
		fs:            fs,
		nameConverter: nameConverter,
		useDefaults:   useDefaults,
	}
}

func (f *flagReader) LookupEnv(key string) (string, bool) {
	if f.nameConverter != nil {
		key = f.nameConverter.ToFlagName(key)
	}
	if flg := f.fs.Lookup(key); flg != nil {
		if f.useDefaults {
			return flg.Value.String(), true
		}
		ok := false
		result := ""
		f.fs.Visit(func(flg *flag.Flag) {
			if flg.Name == key {
				ok = true
				result = flg.Value.String()
			}
		})
		return result, ok
	}
	return "", false
}

func (f *flagReader) Environ() []string {
	result := make([]string, 0)
	if f.useDefaults {
		f.fs.VisitAll(func(flg *flag.Flag) {
			key := flg.Name
			if f.nameConverter != nil {
				key = f.nameConverter.ToEnvName(key)
			}
			result = append(result, key+"="+flg.Value.String())
		})
	} else {
		f.fs.Visit(func(flg *flag.Flag) {
			key := flg.Name
			if f.nameConverter != nil {
				key = f.nameConverter.ToEnvName(key)
			}
			result = append(result, key+"="+flg.Value.String())
		})
	}
	return result
}
