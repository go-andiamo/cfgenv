package cfgenv

import (
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
