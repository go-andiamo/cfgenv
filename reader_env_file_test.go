package cfgenv

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
)

func TestNewEnvFileReader(t *testing.T) {
	efr := NewEnvFileReader(nil, nil)
	require.NotNil(t, efr)
}

func TestEnvFileReader_LookupEnv(t *testing.T) {
	r := strings.NewReader(`FOO=bar`)
	efr := NewEnvFileReader(r, nil)
	s, ok := efr.LookupEnv("FOO")
	require.True(t, ok)
	require.Equal(t, "bar", s)
}

func TestEnvFileReader_Environ(t *testing.T) {
	r := strings.NewReader(`FOO=bar`)
	efr := NewEnvFileReader(r, nil)
	vals := efr.Environ()
	require.Len(t, vals, 1)
	require.Equal(t, "FOO=bar", vals[0])
}

func TestEnvFileReader(t *testing.T) {
	testCases := []struct {
		env      string
		expect   string
		expectOk bool
	}{
		{
			env:      ``,
			expect:   ``,
			expectOk: false,
		},
		{
			env:      `FOO=`,
			expect:   ``,
			expectOk: true,
		},
		{
			env:      `FOO`,
			expect:   ``,
			expectOk: true,
		},
		{
			env:      `FOO=bar`,
			expect:   `bar`,
			expectOk: true,
		},
		{
			env:      " \tFOO=bar \t",
			expect:   `bar`,
			expectOk: true,
		},
		{
			env: `FOO=foo
FOO=bar`,
			expect:   `bar`,
			expectOk: true,
		},
		{
			env:      `FOO="bar"`,
			expect:   `bar`,
			expectOk: true,
		},
		{
			env:      `FOO='bar'`,
			expect:   `bar`,
			expectOk: true,
		},
		{
			env: `# this is a comment
FOO=bar`,
			expect:   `bar`,
			expectOk: true,
		},
		{
			env: `  	# this is a comment
FOO=bar`,
			expect:   `bar`,
			expectOk: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			efr := NewEnvFileReader(strings.NewReader(tc.env), nil)
			v, ok := efr.LookupEnv("FOO")
			assert.Equal(t, tc.expectOk, ok)
			assert.Equal(t, tc.expect, v)
		})
	}
}

func TestEnvFileReader_ScanErrors(t *testing.T) {
	require.Panics(t, func() {
		efr := NewEnvFileReader(nil, nil)
		_ = efr.Environ()
	})
	require.Panics(t, func() {
		efr := NewEnvFileReader(&erroringReader{}, nil)
		_ = efr.Environ()
	})
	called := false
	errHandler := func(err error) {
		called = true
	}
	efr := NewEnvFileReader(&erroringReader{}, errHandler)
	_ = efr.Environ()
	require.True(t, called)
}

type erroringReader struct{}

var _ io.Reader = (*erroringReader)(nil)

func (e *erroringReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
