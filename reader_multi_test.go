package cfgenv

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewMultiEnvReader(t *testing.T) {
	er := NewMultiEnvReader(nil, NewEnvReader(), NewEnvFileReader(nil, nil), MapEnvReader{}, NewEnvReader(), NewFlagReader(nil, false))
	rer := er.(*multiEnvReader)
	require.Equal(t, 4, len(rer.readers))
}

func TestMapEnvReader_LookupEnv(t *testing.T) {
	er1 := MapEnvReader{
		"TEST": "${FOO}-${BAR}",
	}
	er2 := MapEnvReader{
		"FOO": "foo",
	}
	er3 := NewEnvFileReader(strings.NewReader(`BAR=bar
FOO=xxx`), nil)
	er := NewMultiEnvReader(er1, er2, er3)
	v, ok := er.LookupEnv("TEST")
	require.True(t, ok)
	require.Equal(t, "${FOO}-${BAR}", v)
	v, ok = er.LookupEnv("FOO")
	require.True(t, ok)
	require.Equal(t, "foo", v)
	v, ok = er.LookupEnv("BAR")
	require.True(t, ok)
	require.Equal(t, "bar", v)
	v, ok = er.LookupEnv("XXX")
	require.False(t, ok)

	exp := Expand()
	s := exp.Expand("${TEST}", er)
	require.Equal(t, "foo-bar", s)
}

func TestMultiEnvReader_Environ(t *testing.T) {
	er1 := MapEnvReader{
		"TEST": "${FOO}-${BAR}",
	}
	er2 := MapEnvReader{
		"FOO": "foo",
	}
	er3 := NewEnvFileReader(strings.NewReader(`BAR=bar
FOO=xxx`), nil)
	er := NewMultiEnvReader(er1, er2, er3)
	env := er.Environ()
	require.Len(t, env, 3)
	require.Equal(t, "TEST=${FOO}-${BAR}", env[0])
	require.Equal(t, "FOO=foo", env[1])
	require.Equal(t, "BAR=bar", env[2])
}
