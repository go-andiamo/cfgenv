package cfgenv

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestDateTimeSetterOption(t *testing.T) {
	type Config struct {
		Test time.Time
	}
	_ = os.Setenv("TEST", "2024-01-01T18:00:00Z")
	cfg := &Config{}
	err := Load(cfg, NewDatetimeSetter(""))
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-01T18:00:00Z", cfg.Test.Format(time.RFC3339))

	_ = os.Setenv("TEST", "2024-01-01")
	err = Load(cfg, NewDatetimeSetter(""))
	assert.Error(t, err)
	err = Load(cfg, NewDatetimeSetter("2006-01-01"))
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-01T00:00:00Z", cfg.Test.Format(time.RFC3339))
}

func TestDefaultNamingOption(t *testing.T) {
	type Config struct {
		TestMe string
	}
	ct := reflect.TypeOf(Config{})
	fld := ct.Field(0)
	testCases := []struct {
		prefix       string
		separator    string
		overrideName string
		expect       string
	}{
		{
			expect: "TEST_ME",
		},
		{
			overrideName: "FOO",
			expect:       "FOO",
		},
		{
			prefix: "MY",
			expect: "MYTEST_ME",
		},
		{
			prefix:    "MY",
			separator: "_",
			expect:    "MY_TEST_ME",
		},
		{
			separator: "_",
			expect:    "TEST_ME",
		},
		{
			prefix:       "MY",
			separator:    "_",
			overrideName: "FOO",
			expect:       "MY_FOO",
		},
		{
			prefix:       "",
			separator:    "_",
			overrideName: "FOO",
			expect:       "FOO",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := defaultNamingOption.BuildName(tc.prefix, tc.separator, fld, tc.overrideName)
			assert.Equal(t, tc.expect, n)
		})
	}
}

func TestExpand(t *testing.T) {
	ex := Expand()
	v := ex.Expand("${FOO}-${BAR}")
	assert.Equal(t, "-", v)
	_ = os.Setenv("FOO", "a")
	_ = os.Setenv("BAR", "b")
	v = ex.Expand("${FOO}-${BAR}")
	assert.Equal(t, "a-b", v)

	_ = os.Setenv("FOO", "${BAZ}")
	_ = os.Setenv("BAZ", "baz!")
	v = ex.Expand("${FOO}-${BAR}")
	assert.Equal(t, "baz!-b", v)

	ex = Expand(map[string]string{"FOO": "foo!"}, map[string]string{"BAR": "bar!"})
	v = ex.Expand("${FOO}-${BAR}")
	assert.Equal(t, "foo!-bar!", v)
}
