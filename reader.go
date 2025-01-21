package cfgenv

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type EnvReader interface {
	LookupEnv(key string) (string, bool)
	Environ() []string
}

type envReader struct{}

var defaultReader EnvReader = &envReader{}

func (e *envReader) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (e *envReader) Environ() []string {
	return os.Environ()
}

func defaultErrHandler(err error) {
	panic(err)
}

type envFileReader struct {
	f          io.Reader
	errHandler func(err error)
	read       bool
	vars       map[string]string
}

var _ EnvReader = (*envFileReader)(nil)

func NewEnvFileReader(f io.Reader, errHandler func(err error)) EnvReader {
	eh := errHandler
	if eh == nil {
		eh = defaultErrHandler
	}
	return &envFileReader{
		f:          f,
		errHandler: eh,
		vars:       make(map[string]string),
	}
}

func (e *envFileReader) LookupEnv(key string) (string, bool) {
	if e.readFile() {
		if v, ok := e.vars[key]; ok {
			return v, true
		}
	}
	return "", false
}

func (e *envFileReader) Environ() []string {
	result := make([]string, 0, len(e.vars))
	if e.readFile() {
		for k, v := range e.vars {
			result = append(result, k+"="+v)
		}
	}
	return result
}

func (e *envFileReader) readFile() bool {
	if !e.read {
		e.read = true
		scanner := bufio.NewScanner(e.f)
		for scanner.Scan() {
			e.readLine(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			e.errHandler(err)
			return false
		}
	}
	return true
}

func (e *envFileReader) readLine(line string) {
	if trimmed := strings.Trim(line, "\t "); trimmed != "" {
		if strings.HasPrefix(trimmed, "#") {
			return
		}
		if parts := strings.SplitN(trimmed, "=", 2); len(parts) == 2 {
			if (strings.HasPrefix(parts[1], `"`) && strings.HasSuffix(parts[1], `"`)) ||
				(strings.HasPrefix(parts[1], `'`) && strings.HasSuffix(parts[1], `'`)) {
				e.vars[parts[0]] = parts[1][1 : len(parts[1])-1]
			} else {
				e.vars[parts[0]] = parts[1]
			}
		} else {
			e.vars[trimmed] = ""
		}
	}
	return
}
