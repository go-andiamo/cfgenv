package cfgenv

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExampleOf(t *testing.T) {
	type cfg struct {
		Test string `env:"optional,default=foo"`
	}
	var w bytes.Buffer
	err := ExampleOf[cfg](&w)
	assert.NoError(t, err)
	assert.Equal(t, "TEST=foo\n", w.String())
}

func TestExampleOf_ErrorOnNonStruct(t *testing.T) {
	err := ExampleOf[string](nil)
	assert.Error(t, err)
	assert.Equal(t, "cfg not a struct", err.Error())
}

func TestExample_ErrorsOnNonPtr(t *testing.T) {
	type cfg struct{}
	err := Example(nil, cfg{})
	assert.Error(t, err)
	assert.Equal(t, "cfg not a pointer", err.Error())
}

func TestExample_ErrorsOnNonStruct(t *testing.T) {
	cfg := ""
	err := Example(nil, &cfg)
	assert.Error(t, err)
	assert.Equal(t, "cfg not a struct", err.Error())
}

func TestExample_ErrorsOnBadOption(t *testing.T) {
	type cfg struct{}
	err := Example(nil, &cfg{}, "")
	assert.Error(t, err)
	assert.Equal(t, "invalid option", err.Error())
}

func TestWrite_ErrorsOnNonPtr(t *testing.T) {
	type cfg struct{}
	err := Write(nil, cfg{})
	assert.Error(t, err)
	assert.Equal(t, "cfg not a pointer", err.Error())
}

func TestWrite_ErrorsOnNonStruct(t *testing.T) {
	cfg := ""
	err := Write(nil, &cfg)
	assert.Error(t, err)
	assert.Equal(t, "cfg not a struct", err.Error())
}

func TestWrite_ErrorsOnBadOption(t *testing.T) {
	type cfg struct{}
	err := Write(nil, &cfg{}, "")
	assert.Error(t, err)
	assert.Equal(t, "invalid option", err.Error())
}

func TestExample(t *testing.T) {
	s := "foo"
	i := 1
	b := true
	f := 1.1
	type inner struct {
		Test string
	}
	testCases := []struct {
		cfg         any
		actual      bool
		options     []any
		expectError string
		expect      string
	}{
		{
			cfg: &struct {
				Test string
			}{},
			expect: `TEST=<string>
`,
		},
		{
			cfg: &struct {
				Test string `env:"default=foo"`
			}{},
			expect: `TEST=foo
`,
		},
		{
			cfg: &struct {
				Test bool
			}{},
			expect: `TEST=true|false
`,
		},
		{
			cfg: &struct {
				Test int
			}{},
			expect: `TEST=0
`,
		},
		{
			cfg: &struct {
				Test float32
			}{},
			expect: `TEST=0.0
`,
		},
		{
			cfg: &struct {
				Test []string
			}{},
			expect: `TEST=value,value,...
`,
		},
		{
			cfg: &struct {
				Test map[string]string
			}{},
			expect: `TEST=key:value,key:value,...
`,
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				}
			}{},
			expect: `TEST=<string>
`,
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			expect: `SUB_TEST=<string>
`,
		},
		{
			cfg: &struct {
				Test map[string]string `env:"prefix="`
			}{},
			expect: ``,
		},
		{
			cfg: &struct {
				Test error
			}{},
			expectError: `field 'Test' has unsupported type - error`,
		},
		{
			cfg: &struct {
				Inner struct {
					Test error
				}
			}{},
			expectError: `field 'Test' has unsupported type - error`,
		},
		{
			cfg: &struct {
				Test string
			}{},
			options: []any{NewPrefix("APP")},
			expect: `APP_TEST=<string>
`,
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			options: []any{NewPrefix("APP")},
			expect: `APP_SUB_TEST=<string>
`,
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			options: []any{NewPrefix("APP"), NewSeparator(".")},
			expect: `APP.SUB.TEST=<string>
`,
		},
		{
			cfg: &struct {
				Test   string
				Mapped map[string]string `env:"prefix=APP_"`
			}{},
			expect: `TEST=<string>
`,
		},
		{
			cfg: &struct {
				Mapped map[string]string `env:"prefix=APP_"`
			}{
				Mapped: map[string]string{
					"APP_FOO": "foo",
				},
			},
			actual: true,
			expect: `APP_FOO=foo
`,
		},
		{
			cfg: &struct {
				Mapped map[string]string `env:"prefix=APP_"`
			}{},
			actual: true,
		},
		{
			cfg: &struct {
				Test string
			}{
				Test: "foo",
			},
			actual: true,
			expect: `TEST=foo
`,
		},
		{
			cfg: &struct {
				Test *string
			}{},
			actual: true,
			expect: ``,
		},
		{
			cfg: &struct {
				Test *string
			}{
				Test: &s,
			},
			actual: true,
			expect: `TEST=foo
`,
		},
		{
			cfg: &struct {
				Test bool
			}{
				Test: true,
			},
			actual: true,
			expect: `TEST=true
`,
		},
		{
			cfg: &struct {
				Test bool
			}{
				Test: false,
			},
			actual: true,
			expect: `TEST=false
`,
		},
		{
			cfg: &struct {
				Test *bool
			}{
				Test: &b,
			},
			actual: true,
			expect: `TEST=true
`,
		},
		{
			cfg: &struct {
				Test *bool
			}{},
			actual: true,
			expect: ``,
		},
		{
			cfg: &struct {
				Test int
			}{
				Test: 10,
			},
			actual: true,
			expect: `TEST=10
`,
		},
		{
			cfg: &struct {
				Test uint
			}{
				Test: 10,
			},
			actual: true,
			expect: `TEST=10
`,
		},
		{
			cfg: &struct {
				Test *int
			}{
				Test: &i,
			},
			actual: true,
			expect: `TEST=1
`,
		},
		{
			cfg: &struct {
				Test *int
			}{},
			actual: true,
			expect: ``,
		},
		{
			cfg: &struct {
				Test float32
			}{
				Test: 10.2,
			},
			actual: true,
			expect: `TEST=10.2
`,
		},
		{
			cfg: &struct {
				Test *float64
			}{
				Test: &f,
			},
			actual: true,
			expect: `TEST=1.1
`,
		},
		{
			cfg: &struct {
				Test *float64
			}{},
			actual: true,
			expect: ``,
		},
		{
			cfg: &struct {
				Test []int
			}{},
			actual: true,
			expect: `TEST=
`,
		},
		{
			cfg: &struct {
				Test []int
			}{
				Test: []int{1, 2, 3},
			},
			actual: true,
			expect: `TEST=1,2,3
`,
		},
		{
			cfg: &struct {
				Test []float32
			}{
				Test: []float32{1.1, 2.2, 3.3},
			},
			actual: true,
			expect: `TEST=1.1,2.2,3.3
`,
		},
		{
			cfg: &struct {
				Test map[string]float32
			}{},
			actual: true,
			expect: `TEST=
`,
		},
		{
			cfg: &struct {
				Test map[string]float32
			}{
				Test: map[string]float32{
					"foo": 1.1,
				},
			},
			actual: true,
			expect: `TEST=foo:1.1
`,
		},
		{
			cfg: &struct {
				Test1 string `env:"TEST"`
				Test2 string `env:"TEST"`
			}{},
			expect: `TEST=<string>
`,
		},
		{
			cfg: &struct {
				Test1 string `env:"TEST"`
				Test2 string `env:"TEST"`
			}{
				Test1: "foo",
				Test2: "bar",
			},
			actual: true,
			expect: `TEST=foo
`,
		},
		{
			cfg: &struct {
				Test   string
				Mapped map[string]string `env:"prefix="`
			}{
				Test: "foo",
				Mapped: map[string]string{
					"TEST": "bar",
				},
			},
			actual: true,
			expect: `TEST=foo
`,
		},
		{
			cfg: &struct {
				Inner *inner `env:"prefix=SUB"`
			}{},
			actual: true,
		},
		{
			cfg: &struct {
				Inner *inner `env:"prefix=SUB"`
			}{},
			expect: `SUB_TEST=<string>
`,
		},
		{
			cfg: &struct {
				Inner *inner `env:"prefix=SUB"`
			}{
				Inner: &inner{
					Test: "foo",
				},
			},
			actual: true,
			expect: `SUB_TEST=foo
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var w bytes.Buffer
			var err error
			if tc.actual {
				err = Write(&w, tc.cfg, tc.options...)
			} else {
				err = Example(&w, tc.cfg, tc.options...)
			}
			if tc.expectError != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, w.String())
			}
		})
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("fooey")
}

func TestExample_ErrorsWithWriteError(t *testing.T) {
	testCases := []struct {
		cfg    any
		actual bool
	}{
		{
			cfg: &struct {
				Test string
			}{},
		},
		{
			cfg: &struct {
				Test string
			}{},
			actual: true,
		},
		{
			cfg: &struct {
				Mapped map[string]string `env:"prefix=APP_"`
			}{
				Mapped: map[string]string{
					"APP_FOO": "foo",
				},
			},
			actual: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := &errorWriter{}
			var err error
			if tc.actual {
				err = Write(w, tc.cfg)
			} else {
				err = Example(w, tc.cfg)
			}
			assert.Error(t, err)
			assert.Equal(t, "fooey", err.Error())
		})
	}
}
