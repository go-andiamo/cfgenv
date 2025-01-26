package cfgenv

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewEnvReader(t *testing.T) {
	er := NewEnvReader()
	require.NotNil(t, er)
}

func TestEnvReader_LookupEnv(t *testing.T) {
	os.Clearenv()
	err := os.Setenv("FOO", "foo")
	require.NoError(t, err)

	er := NewEnvReader()
	require.NotNil(t, er)
	v, ok := er.LookupEnv("FOO")
	require.True(t, ok)
	require.Equal(t, "foo", v)
	_, ok = er.LookupEnv("BAR")
	require.False(t, ok)
}

func TestEnvReader_Environ(t *testing.T) {
	os.Clearenv()
	err := os.Setenv("FOO", "foo")
	require.NoError(t, err)

	er := NewEnvReader()
	require.NotNil(t, er)
	env := er.Environ()
	require.Len(t, env, 1)
	require.Equal(t, "FOO=foo", env[0])
}
