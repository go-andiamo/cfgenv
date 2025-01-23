package cfgenv

import (
	"encoding/base64"
	"github.com/go-andiamo/gopt"
	"github.com/stretchr/testify/assert"
	"os"
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

func TestDateTimeSetterOption_Encoded(t *testing.T) {
	type Config struct {
		Test time.Time `env:"encoding=base64"`
	}
	_ = os.Setenv("TEST", base64.StdEncoding.EncodeToString([]byte("2024-01-01T18:00:00Z")))
	cfg := &Config{}
	err := Load(cfg, NewDatetimeSetter(""))
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-01T18:00:00Z", cfg.Test.Format(time.RFC3339))

	_ = os.Setenv("TEST", "not properly encoded")
	cfg = &Config{}
	err = Load(cfg, NewDatetimeSetter(""))
	assert.Error(t, err)
}

func TestDateTimeSetterOption_WithExpand(t *testing.T) {
	type Config struct {
		Test time.Time
	}
	menv := MapEnvReader{
		"TEST":  "${YEAR}-${MONTH}-${DAY}",
		"YEAR":  "2024",
		"MONTH": "01",
		"DAY":   "10",
	}
	cfg := &Config{}
	err := Load(cfg, NewDatetimeSetter("2006-01-02"), Expand(), menv)
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-10", cfg.Test.Format("2006-01-02"))
}

func TestDateTimeSetterOption_Optional(t *testing.T) {
	type Config struct {
		Test gopt.Optional[time.Time] `env:"optional,default=2024-01-01"`
	}
	menv := MapEnvReader{
		"TEST": "2024-01-10",
	}
	cfg := &Config{}
	err := Load(cfg, NewDatetimeSetter("2006-01-02"), menv)
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-10", cfg.Test.Default(time.Now()).Format("2006-01-02"))
	assert.True(t, cfg.Test.WasSet())
	assert.True(t, cfg.Test.IsPresent())

	menv = MapEnvReader{}
	cfg = &Config{}
	err = Load(cfg, NewDatetimeSetter("2006-01-02"), menv)
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-01", cfg.Test.Default(time.Now()).Format("2006-01-02"))
	assert.False(t, cfg.Test.WasSet())
	assert.True(t, cfg.Test.IsPresent())
}

func TestDurationSetterOption(t *testing.T) {
	type Config struct {
		Test time.Duration
	}
	_ = os.Setenv("TEST", "0")
	cfg := &Config{}
	err := Load(cfg, NewDurationSetter())
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), cfg.Test)

	_ = os.Setenv("TEST", "16m")
	err = Load(cfg, NewDurationSetter())
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(16*time.Minute), cfg.Test)

	_ = os.Setenv("TEST", "not a duration")
	err = Load(cfg, NewDurationSetter())
	assert.Error(t, err)
}

func TestDurationSetterOption_Optional(t *testing.T) {
	type Config struct {
		Test gopt.Optional[time.Duration] `env:"optional,default=16m"`
	}
	menv := MapEnvReader{
		"TEST": "2h",
	}
	cfg := &Config{}
	err := Load(cfg, NewDurationSetter(), menv)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(2*time.Hour), cfg.Test.Default(0))
	assert.True(t, cfg.Test.WasSet())
	assert.True(t, cfg.Test.IsPresent())

	menv = MapEnvReader{}
	cfg = &Config{}
	err = Load(cfg, NewDurationSetter(), menv)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(16*time.Minute), cfg.Test.Default(0))
	assert.False(t, cfg.Test.WasSet())
	assert.True(t, cfg.Test.IsPresent())
}
