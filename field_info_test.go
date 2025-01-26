package cfgenv

import (
	"fmt"
	"github.com/go-andiamo/gopt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestGetFieldInfo(t *testing.T) {
	type config struct {
		Foo string
	}
	fld := reflect.TypeOf(config{}).Field(0)
	fi, err := getFieldInfo(fld, &opts{})
	require.NoError(t, err)
	require.NotNil(t, fi)
	assert.Equal(t, "", fi.name)
	assert.False(t, fi.optional)
	assert.False(t, fi.pointer)
	assert.False(t, fi.hasDefault)
	assert.Equal(t, "", fi.defaultValue)
	assert.Equal(t, "", fi.prefix)
	assert.False(t, fi.isStruct)
	assert.False(t, fi.isPrefixedMap)
	assert.False(t, fi.isMatchedMap)
	assert.Nil(t, fi.matchRegex)
	assert.Nil(t, fi.customSetter)
	assert.Nil(t, fi.optionalSetter)
	assert.Nil(t, fi.decoder)
	assert.Equal(t, ":", fi.separator)
	assert.Equal(t, ",", fi.delimiter)
	assert.False(t, fi.expand)
	assert.False(t, fi.noExpand)
}

func TestGetFieldInfo_Tags(t *testing.T) {
	testCases := []struct {
		cfg                  any
		options              []any
		expectErr            bool
		expectName           string
		expectOptional       bool
		expectPointer        bool
		expectOptionalSetter bool
		expectDefault        string
		expectPrefix         string
		expectStruct         bool
		expectPrefixedMap    bool
		expectMatchedMap     bool
		expectMatchRegexp    bool
		expectSeparator      string
		expectDelimiter      string
		expectDecoder        bool
		expectExpand         bool
		expectNoExpand       bool
		expectCustomSetter   bool
	}{
		{
			cfg: struct {
				Test string `env:"TEST_ME"`
			}{},
			expectName: "TEST_ME",
		},
		{
			cfg: struct {
				Test string `env:"'TEST_ME'"`
			}{},
			expectName: "TEST_ME",
		},
		{
			cfg: struct {
				Test string `env:"name=TEST_ME"`
			}{},
			expectName: "TEST_ME",
		},
		{
			cfg: struct {
				Test string `env:"name='TEST_ME'"`
			}{},
			expectName: "TEST_ME",
		},
		{
			cfg: struct {
				Test string `env:"unknown=foo"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test *string
			}{},
			expectOptional: true,
			expectPointer:  true,
		},
		{
			cfg: struct {
				Test error
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test gopt.Optional[string]
			}{},
			expectOptional:       true,
			expectOptionalSetter: true,
		},
		{
			cfg: struct {
				Test string `env:",unbalanced '"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"default=foo"`
			}{},
			expectDefault: "foo",
		},
		{
			cfg: struct {
				Test string `env:"default='foo'"`
			}{},
			expectDefault: "foo",
		},
		{
			cfg: struct {
				Test string `env:"prefix=foo"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test struct{ Foo string } `env:"prefix=foo"`
			}{},
			expectPrefix: "foo",
			expectStruct: true,
		},
		{
			cfg: struct {
				Test struct{ Foo string } `env:"prefix='foo'"`
			}{},
			expectPrefix: "foo",
			expectStruct: true,
		},
		{
			cfg: struct {
				Test map[string]string `env:"prefix=foo"`
			}{},
			expectPrefix:      "foo",
			expectPrefixedMap: true,
		},
		{
			cfg: struct {
				Test map[string]string `env:"prefix='foo'"`
			}{},
			expectPrefix:      "foo",
			expectPrefixedMap: true,
		},
		{
			cfg: struct {
				Test string `env:"match=*"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test map[string]string `env:"match=*"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test map[string]string `env:"match=.*"`
			}{},
			expectMatchedMap:  true,
			expectMatchRegexp: true,
		},
		{
			cfg: struct {
				Test map[string]int `env:"separator=>"`
			}{},
			expectSeparator: ">",
		},
		{
			cfg: struct {
				Test map[string]int `env:"separator='>'"`
			}{},
			expectSeparator: ">",
		},
		{
			cfg: struct {
				Test map[string]int `env:"sep=>"`
			}{},
			expectSeparator: ">",
		},
		{
			cfg: struct {
				Test map[string]int `env:"sep='>'"`
			}{},
			expectSeparator: ">",
		},
		{
			cfg: struct {
				Test map[string]int `env:"delimiter=>"`
			}{},
			expectDelimiter: ">",
		},
		{
			cfg: struct {
				Test map[string]int `env:"delimiter='>'"`
			}{},
			expectDelimiter: ">",
		},
		{
			cfg: struct {
				Test string `env:"encoding=unknown"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"encoding=base64"`
			}{},
			expectDecoder: true,
		},
		{
			cfg: struct {
				Test string `env:"encoding='base64'"`
			}{},
			expectDecoder: true,
		},
		{
			cfg: struct {
				Test string `env:"optional"`
			}{},
			expectOptional: true,
		},
		{
			cfg: struct {
				Test string `env:"expand"`
			}{},
			expectExpand: true,
		},
		{
			cfg: struct {
				Test string `env:"no-expand"`
			}{},
			expectNoExpand: true,
		},
		{
			cfg: struct {
				Test string `env:"expand,no-expand"`
			}{},
			expectNoExpand: true,
		},
		{
			cfg: struct {
				Test string `env:"unknown=unknown=unknown"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"default"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"prefix"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"separator"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"sep"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"delimiter"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"delim"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"match"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test string `env:"encoding"`
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test error
			}{},
			options:            []any{&errorSetter{}},
			expectCustomSetter: true,
		},
		{
			cfg: struct {
				Test []error
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test *[]int
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test []int
			}{},
		},
		{
			cfg: struct {
				Test []*int
			}{},
		},
		{
			cfg: struct {
				Test *map[int]int
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test map[int]*int
			}{},
		},
		{
			cfg: struct {
				Test map[error]int
			}{},
			expectErr: true,
		},
		{
			cfg: struct {
				Test map[int]error
			}{},
			expectErr: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			fld := reflect.TypeOf(tc.cfg).Field(0)
			options, err := buildOpts(tc.options...)
			require.NoError(t, err)
			fi, err := getFieldInfo(fld, options)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectName, fi.name)
				assert.Equal(t, tc.expectOptional, fi.optional)
				assert.Equal(t, tc.expectPointer, fi.pointer)
				assert.Equal(t, tc.expectOptionalSetter, fi.optionalSetter != nil)
				assert.Equal(t, tc.expectDefault, fi.defaultValue)
				assert.Equal(t, tc.expectPrefix, fi.prefix)
				assert.Equal(t, tc.expectStruct, fi.isStruct)
				assert.Equal(t, tc.expectPrefixedMap, fi.isPrefixedMap)
				assert.Equal(t, tc.expectMatchedMap, fi.isMatchedMap)
				assert.Equal(t, tc.expectMatchRegexp, fi.matchRegex != nil)
				if tc.expectSeparator != "" {
					assert.Equal(t, tc.expectSeparator, fi.separator)
				}
				if tc.expectDelimiter != "" {
					assert.Equal(t, tc.expectDelimiter, fi.delimiter)
				}
				assert.Equal(t, tc.expectDecoder, fi.decoder != nil)
				assert.Equal(t, tc.expectExpand, fi.expand)
				assert.Equal(t, tc.expectNoExpand, fi.noExpand)
				assert.Equal(t, tc.expectCustomSetter, fi.customSetter != nil)
			}
		})
	}
}

type errorSetter struct{}

var _ CustomSetterOption = &errorSetter{}

func (e *errorSetter) IsApplicable(fld reflect.StructField) bool {
	return true
}

func (e *errorSetter) Set(fld reflect.StructField, v reflect.Value, raw string, present bool) error {
	// test only - does nothing
	return nil
}
