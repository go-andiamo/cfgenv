package cfgenv

import (
	"os"
)

// EnvReader is an option that can be passed to Load or LoadAs
// an abstraction around reading environment variables fom various sources
type EnvReader interface {
	// LookupEnv see os.LookupEnv
	LookupEnv(key string) (string, bool)
	// Environ see os.Environ
	Environ() []string
}

type envReader struct{}

var defaultReader EnvReader = &envReader{}

// NewEnvReader creates a new EnvReader that reads from environment vars using os.LookupEnv and os.Environ
func NewEnvReader() EnvReader {
	return defaultReader
}

func (e *envReader) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (e *envReader) Environ() []string {
	return os.Environ()
}
