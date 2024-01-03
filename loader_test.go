package cfgenv

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadAs(t *testing.T) {
	type cfg struct {
		Test string `env:"optional,default=foo"`
	}
	c, err := LoadAs[cfg]()
	assert.NoError(t, err)
	assert.Equal(t, "foo", c.Test)
}

func TestLoadAs_ErrorOnNonStruct(t *testing.T) {
	_, err := LoadAs[string]()
	assert.Error(t, err)
	assert.Equal(t, "cfg not a struct", err.Error())
}

func TestLoad_ErrorsOnNonPtr(t *testing.T) {
	type cfg struct{}
	err := Load(cfg{})
	assert.Error(t, err)
	assert.Equal(t, "cfg not a pointer", err.Error())
}

func TestLoad_ErrorsOnNonStruct(t *testing.T) {
	cfg := ""
	err := Load(&cfg)
	assert.Error(t, err)
	assert.Equal(t, "cfg not a struct", err.Error())
}

func TestLoad_ErrorsOnBadOption(t *testing.T) {
	type cfg struct{}
	err := Load(&cfg{}, "")
	assert.Error(t, err)
	assert.Equal(t, "invalid option", err.Error())

	err = Load(&cfg{}, nil)
	assert.NoError(t, err)
}

type customLowercaseNaming struct{}

func (c *customLowercaseNaming) BuildName(prefix string, separator string, fld reflect.StructField, tagName string) string {
	name := tagName
	if name == "" {
		name = toSnakeCase(fld.Name)
	}
	if prefix != "" {
		name = prefix + separator + name
	}
	return strings.ToLower(name)
}

var _ NamingOption = &customLowercaseNaming{}

func TestLoad(t *testing.T) {
	testCases := []struct {
		cfg         any
		env         map[string]string
		options     []any
		expectError string
		expect      string
	}{
		{
			cfg: &struct {
				BadTag string `env:"unknown=foo"`
			}{},
			expectError: "invalid tag 'unknown=foo' on field 'BadTag'",
		},
		{
			cfg: &struct {
				BadTag string `env:"unbalanced '"`
			}{},
			expectError: "invalid tag 'unbalanced '' on field 'BadTag'",
		},
		{
			cfg: &struct {
				Test string
			}{},
			expectError: "missing env var 'TEST'",
		},
		{
			cfg: &struct {
				Test string
			}{},
			options:     []any{&customLowercaseNaming{}},
			expectError: "missing env var 'test'",
		},
		{
			cfg: &struct {
				Test string
			}{},
			options:     []any{&customLowercaseNaming{}, &customLowercaseNaming{}},
			expectError: "multiple naming options",
		},
		{
			cfg: &struct {
				Test string `env:"FOO"`
			}{},
			expectError: "missing env var 'FOO'",
		},
		{
			cfg: &struct {
				Test string `env:"optional"`
			}{},
		},
		{
			cfg: &struct {
				Test string `env:"'optional'"`
			}{},
			expectError: "missing env var 'optional'",
		},
		{
			cfg: &struct {
				Test string
			}{},
			options:     []any{NewPrefix("APP")},
			expectError: "missing env var 'APP_TEST'",
		},
		{
			cfg: &struct {
				Test string
			}{},
			options:     []any{NewPrefix("APP"), NewPrefix("APP")},
			expectError: "multiple prefix options",
		},
		{
			cfg: &struct {
				Test string
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expect: `{"Test":"foo"}`,
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			expectError: "missing env var 'SUB_TEST'",
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			env: map[string]string{
				"SUB_TEST": "foo",
			},
			expect: `{"Inner":{"Test":"foo"}}`,
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			options:     []any{NewSeparator(".")},
			expectError: "missing env var 'SUB.TEST'",
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			options:     []any{NewSeparator("."), NewSeparator(".")},
			expectError: "multiple separator options",
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				} `env:"prefix=SUB"`
			}{},
			options:     []any{NewPrefix("APP")},
			expectError: "missing env var 'APP_SUB_TEST'",
		},
		{
			cfg: &struct {
				Inner struct {
					Test string
				}
			}{},
			options:     []any{NewPrefix("APP")},
			expectError: "missing env var 'APP_TEST'",
		},
		{
			cfg: &struct {
				Test bool
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not a bool",
		},
		{
			cfg: &struct {
				Test bool
			}{},
			env: map[string]string{
				"TEST": "true",
			},
			expect: `{"Test":true}`,
		},
		{
			cfg: &struct {
				Test int
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test int
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test int8
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test int8
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test int16
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test int16
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test int32
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test int32
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test int64
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test int64
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test time.Duration
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test time.Duration
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test uint
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an uint",
		},
		{
			cfg: &struct {
				Test uint
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test uint8
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an uint",
		},
		{
			cfg: &struct {
				Test uint8
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test uint16
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an uint",
		},
		{
			cfg: &struct {
				Test uint16
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test uint32
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an uint",
		},
		{
			cfg: &struct {
				Test uint32
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test uint64
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an uint",
		},
		{
			cfg: &struct {
				Test uint64
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test float32
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not a float",
		},
		{
			cfg: &struct {
				Test float32
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test float64
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not a float",
		},
		{
			cfg: &struct {
				Test float64
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test []int
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test []int
			}{},
			env: map[string]string{
				"TEST": "1,2,3",
			},
			expect: `{"Test":[1,2,3]}`,
		},
		{
			cfg: &struct {
				Test []int `env:"optional,default='1,2,3'"`
			}{},
			expect: `{"Test":[1,2,3]}`,
		},
		{
			cfg: &struct {
				Test map[string]int
			}{},
			env: map[string]string{
				"TEST": "foo:foo",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test map[int]int
			}{},
			env: map[string]string{
				"TEST": "foo:1",
			},
			expectError: "env var 'TEST' is not an int",
		},
		{
			cfg: &struct {
				Test map[int]int
			}{},
			env: map[string]string{
				"TEST": "1:10,2:20",
			},
			expect: `{"Test":{"1":10,"2":20}}`,
		},
		{
			cfg: &struct {
				Test map[string]int
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "env var 'TEST' contains invalid key/value pair - foo",
		},
		{
			cfg: &struct {
				Test error
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expectError: "field 'Test' has unsupported type - error",
		},
		{
			cfg: &struct {
				Test *string
			}{},
			expect: `{"Test":null}`,
		},
		{
			cfg: &struct {
				Test *string `env:"default=foo"`
			}{},
			expect: `{"Test":"foo"}`,
		},
		{
			cfg: &struct {
				Test *string
			}{},
			env: map[string]string{
				"TEST": "foo",
			},
			expect: `{"Test":"foo"}`,
		},
		{
			cfg: &struct {
				Test *bool
			}{},
			expect: `{"Test":null}`,
		},
		{
			cfg: &struct {
				Test *bool
			}{},
			env: map[string]string{
				"TEST": "true",
			},
			expect: `{"Test":true}`,
		},
		{
			cfg: &struct {
				Test *int
			}{},
			expect: `{"Test":null}`,
		},
		{
			cfg: &struct {
				Test *int
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test *uint
			}{},
			expect: `{"Test":null}`,
		},
		{
			cfg: &struct {
				Test *uint
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test *float32
			}{},
			expect: `{"Test":null}`,
		},
		{
			cfg: &struct {
				Test *float32
			}{},
			env: map[string]string{
				"TEST": "1",
			},
			expect: `{"Test":1}`,
		},
		{
			cfg: &struct {
				Test *[]string
			}{},
			expectError: "field 'Test' has unsupported type - *[]string",
		},
		{
			cfg: &struct {
				Test *map[string]string
			}{},
			expectError: "field 'Test' has unsupported type - *map[string]string",
		},
		{
			cfg: &struct {
				Test *struct{}
			}{},
			expectError: "field 'Test' has unsupported embedded struct ptr",
		},
		{
			cfg: &struct {
				Test []struct{ Foo string }
			}{},
			expectError: "field 'Test' has unsupported slice item type",
		},
		{
			cfg: &struct {
				Test []*struct{ Foo string }
			}{},
			expectError: "field 'Test' has unsupported slice item type",
		},
		{
			cfg: &struct {
				Test map[string]struct{ Foo string }
			}{},
			expectError: "field 'Test' has unsupported map item type",
		},
		{
			cfg: &struct {
				Test map[string]*struct{ Foo string }
			}{},
			expectError: "field 'Test' has unsupported map item type",
		},
		{
			cfg: &struct {
				Test map[*string]string
			}{},
			expectError: "field 'Test' has unsupported map key type",
		},
		{
			cfg: &struct {
				Test map[string]string `env:"prefix=STUFF_"`
			}{},
			env: map[string]string{
				"STUFF_FOO": "foo",
			},
			expect: `{"Test":{"FOO":"foo"}}`,
		},
		{
			cfg: &struct {
				Test map[string]string `env:"prefix="`
			}{},
			env: map[string]string{
				"STUFF_FOO": "foo",
			},
			expect: `{"Test":{"STUFF_FOO":"foo"}}`,
		},
		{
			cfg: &struct {
				Inner struct {
					Envs map[string]string `env:"prefix=STUFF_"`
				} `env:"prefix=SUB"`
			}{},
			options: []any{NewPrefix("APP")},
			env: map[string]string{
				"APP_SUB_STUFF_FOO": "foo",
			},
			expect: `{"Inner":{"Envs":{"FOO":"foo"}}}`,
		},
		{
			cfg: &struct {
				Test string `env:"prefix=STUFF_"`
			}{},
			expectError: "cannot use env tag 'prefix' on field 'Test' (only for structs or map[string]string)",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			os.Clearenv()
			for k, v := range tc.env {
				require.NoError(t, os.Setenv(k, v))
			}
			err := Load(tc.cfg, tc.options...)
			if tc.expectError == "" {
				assert.NoError(t, err)
				if tc.expect != "" {
					data, err := json.Marshal(tc.cfg)
					require.NoError(t, err)
					assert.Equal(t, tc.expect, string(data))
				}
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectError, err.Error())
			}
		})
	}
}

type testCustomSetter struct {
	err error
}
type custom []byte

var customType = reflect.TypeOf(custom{})

func (ct *testCustomSetter) IsApplicable(fld reflect.StructField) bool {
	return fld.Type == customType
}

func (ct *testCustomSetter) Set(fld reflect.StructField, v reflect.Value, raw string) error {
	if ct.err != nil {
		return ct.err
	}
	val := custom(raw)
	v.Set(reflect.ValueOf(val))
	return nil
}

var _ CustomSetterOption = &testCustomSetter{}

func TestLoad_WithCustomSetter(t *testing.T) {
	os.Clearenv()

	type MyConfig struct {
		Test1 []byte `env:"optional,default='1,2,3'"`
		Test2 custom `env:"optional,default=foo"`
	}
	cfg := &MyConfig{}
	err := Load(cfg, &testCustomSetter{})
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, cfg.Test1)
	assert.Equal(t, custom("foo"), cfg.Test2)

	err = Load(cfg, &testCustomSetter{err: errors.New("fooey")})
	assert.Error(t, err)
	assert.Equal(t, "fooey", err.Error())

	type MyConfig2 struct {
		Test custom
	}
	cfg2 := &MyConfig2{}
	err = Load(cfg2, &testCustomSetter{})
	assert.Error(t, err)
	assert.Equal(t, "missing env var 'TEST'", err.Error())
}
