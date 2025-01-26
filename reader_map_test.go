package cfgenv

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMapEnvReader(t *testing.T) {
	menv := MapEnvReader{
		"FOO": "foo",
		"BAR": "bar",
	}
	require.Equal(t, 2, len(menv.Environ()))
	v, ok := menv.LookupEnv("FOO")
	require.True(t, ok)
	require.Equal(t, "foo", v)
	v, ok = menv.LookupEnv("BAR")
	require.True(t, ok)
	require.Equal(t, "bar", v)
	_, ok = menv.LookupEnv("XXX")
	require.False(t, ok)
}
