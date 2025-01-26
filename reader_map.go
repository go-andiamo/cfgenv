package cfgenv

// MapEnvReader is a map[string]string that implements the EnvReader interface
type MapEnvReader map[string]string

var _ EnvReader = MapEnvReader{}

func (m MapEnvReader) LookupEnv(key string) (string, bool) {
	v, ok := m[key]
	return v, ok
}

func (m MapEnvReader) Environ() []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, k+"="+v)
	}
	return result
}
