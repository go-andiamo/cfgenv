package cfgenv

import (
	"flag"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewFlagReader(t *testing.T) {
	er := NewFlagReader(nil, false)
	require.NotNil(t, er)
}

func TestFlagReader_LookupEnv(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Bool("test", false, "")
	err := fs.Parse([]string{"-test"})
	require.NoError(t, err)
	er := NewFlagSetReader(fs, nil, true)
	v, ok := er.LookupEnv("test")
	require.True(t, ok)
	require.Equal(t, "true", v)

	_, ok = er.LookupEnv("foo")
	require.False(t, ok)
}

func TestFlagReader_LookupEnv_NoDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Bool("test", false, "")
	err := fs.Parse([]string{})
	require.NoError(t, err)
	er := NewFlagSetReader(fs, nil, false)
	_, ok := er.LookupEnv("test")
	require.False(t, ok)

	err = fs.Parse([]string{"-test"})
	require.NoError(t, err)
	v, ok := er.LookupEnv("test")
	require.True(t, ok)
	require.Equal(t, "true", v)
}

func TestFlagReader_LookupEnv_WithNameConverter(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Bool("test", false, "")
	err := fs.Parse([]string{})
	require.NoError(t, err)
	er := NewFlagSetReader(fs, testNameConverter{}, false)
	_, ok := er.LookupEnv("TEST")
	require.False(t, ok)

	err = fs.Parse([]string{"-test"})
	require.NoError(t, err)
	v, ok := er.LookupEnv("TEST")
	require.True(t, ok)
	require.Equal(t, "true", v)
}

func TestFlagReader_Environ(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Bool("test", false, "")
	fs.Bool("test2", false, "")
	err := fs.Parse([]string{"-test"})
	require.NoError(t, err)
	er := NewFlagSetReader(fs, nil, true)
	env := er.Environ()
	require.Len(t, env, 2)
	require.Equal(t, "test=true", env[0])
}

func TestFlagReader_Environ_WithNameConverter(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Bool("test", false, "")
	fs.Bool("test2", false, "")
	err := fs.Parse([]string{"-test"})
	require.NoError(t, err)
	er := NewFlagSetReader(fs, testNameConverter{}, true)
	env := er.Environ()
	require.Len(t, env, 2)
	require.Equal(t, "TEST=true", env[0])
}

func TestFlagReader_Environ_NoDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Bool("test", false, "")
	fs.Bool("test2", false, "")
	err := fs.Parse([]string{"-test"})
	require.NoError(t, err)
	er := NewFlagSetReader(fs, nil, false)
	env := er.Environ()
	require.Len(t, env, 1)
	require.Equal(t, "test=true", env[0])
}

func TestFlagReader_Environ_NoDefaults_WithNameConverter(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Bool("test", false, "")
	fs.Bool("test2", false, "")
	err := fs.Parse([]string{"-test"})
	require.NoError(t, err)
	er := NewFlagSetReader(fs, testNameConverter{}, false)
	env := er.Environ()
	require.Len(t, env, 1)
	require.Equal(t, "TEST=true", env[0])
}

type testNameConverter struct{}

func (n testNameConverter) ToFlagName(envKey string) string {
	return strings.ToLower(strings.ReplaceAll(envKey, "_", "-"))
}

func (n testNameConverter) ToEnvName(flagName string) string {
	return strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}
