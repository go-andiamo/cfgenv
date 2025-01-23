package cfgenv

import (
	"encoding/base64"
	"github.com/go-andiamo/gopt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOptionalString(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[string]
		Test2 gopt.Optional[string] `env:"default=fooey"`
	}
	er := MapEnvReader{
		"TEST": "${FOO}-${BAR}",
		"FOO":  "foo",
		"BAR":  "bar",
	}
	cfg, err := LoadAs[testCfg](Expand(), er)
	require.NoError(t, err)
	require.Equal(t, "foo-bar", cfg.Test.Default(""))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, "fooey", cfg.Test2.Default(""))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{}
	cfg, err = LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, "", cfg.Test.Default(""))
	require.False(t, cfg.Test.IsPresent())
	require.False(t, cfg.Test.WasSet())
}

func TestOptionalString_WithEncoding(t *testing.T) {
	type testCfg struct {
		Test gopt.Optional[string] `env:"encoding=base64,default=bar"`
	}
	er := MapEnvReader{
		"TEST": base64.StdEncoding.EncodeToString([]byte("foo")),
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, "foo", cfg.Test.Default(""))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())

	er = MapEnvReader{}
	cfg, err = LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, "bar", cfg.Test.Default(""))
	require.True(t, cfg.Test.IsPresent())
	require.False(t, cfg.Test.WasSet())

	er = MapEnvReader{
		"TEST": "not properly encoded",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalBool(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[bool]
		Test2 gopt.Optional[bool] `env:"default=true"`
	}
	er := MapEnvReader{
		"TEST": "true",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, true, cfg.Test.Default(false))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, true, cfg.Test2.Default(false))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a bool",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalFloat32(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[float32]
		Test2 gopt.Optional[float32] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, float32(17.0), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, float32(16.0), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalFloat64(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[float64]
		Test2 gopt.Optional[float64] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, 17.0, cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, 16.0, cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalInt(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[int]
		Test2 gopt.Optional[int] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, 17, cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, 16, cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalInt8(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[int8]
		Test2 gopt.Optional[int8] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, int8(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, int8(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalInt16(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[int16]
		Test2 gopt.Optional[int16] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, int16(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, int16(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalInt32(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[int32]
		Test2 gopt.Optional[int32] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, int32(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, int32(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalInt64(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[int64]
		Test2 gopt.Optional[int64] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, int64(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, int64(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalUint(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[uint]
		Test2 gopt.Optional[uint] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, uint(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, uint(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalUint8(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[uint8]
		Test2 gopt.Optional[uint8] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, uint8(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, uint8(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalUint16(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[uint16]
		Test2 gopt.Optional[uint16] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, uint16(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, uint16(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalUint32(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[uint32]
		Test2 gopt.Optional[uint32] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, uint32(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, uint32(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptionalUint64(t *testing.T) {
	type testCfg struct {
		Test  gopt.Optional[uint64]
		Test2 gopt.Optional[uint64] `env:"default=16"`
	}
	er := MapEnvReader{
		"TEST": "17",
	}
	cfg, err := LoadAs[testCfg](er)
	require.NoError(t, err)
	require.Equal(t, uint64(17), cfg.Test.Default(0))
	require.True(t, cfg.Test.IsPresent())
	require.True(t, cfg.Test.WasSet())
	require.Equal(t, uint64(16), cfg.Test2.Default(0))
	require.True(t, cfg.Test2.IsPresent())
	require.False(t, cfg.Test2.WasSet())

	er = MapEnvReader{
		"TEST": "not a number",
	}
	_, err = LoadAs[testCfg](er)
	require.Error(t, err)
}

func TestOptional_BadDefault(t *testing.T) {
	type testCfg struct {
		Test gopt.Optional[int] `env:"default=not-a-number"`
	}
	er := MapEnvReader{}
	_, err := LoadAs[testCfg](er)
	require.Error(t, err)
}
