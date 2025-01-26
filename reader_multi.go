package cfgenv

import "strings"

// NewMultiEnvReader creates an EnvReader that reads environment vars from multiple EnvReader's
// e.g. MapEnvReader, NewEnvReader, NewEnvFileReader, NewFlagReader
//
// when calling LookupEnv, it successively tries all the provided readers - returning the first found
//
// When call Environ, it merges all environs into one
func NewMultiEnvReader(readers ...EnvReader) EnvReader {
	result := &multiEnvReader{
		readers: make([]EnvReader, 0, len(readers)),
	}
	seenOs := false
	for _, reader := range readers {
		_, isOs := reader.(*envReader)
		if reader != nil && (!seenOs || !isOs) {
			result.readers = append(result.readers, reader)
			seenOs = seenOs || isOs
		}
	}
	return result
}

type multiEnvReader struct {
	readers []EnvReader
}

func (m *multiEnvReader) LookupEnv(key string) (string, bool) {
	for _, reader := range m.readers {
		if env, ok := reader.LookupEnv(key); ok {
			return env, ok
		}
	}
	return "", false
}

func (m *multiEnvReader) Environ() []string {
	result := make([]string, 0)
	keys := map[string]struct{}{}
	for _, reader := range m.readers {
		for _, v := range reader.Environ() {
			nv := strings.SplitN(v, "=", 2)
			if _, ok := keys[nv[0]]; !ok {
				keys[nv[0]] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}
