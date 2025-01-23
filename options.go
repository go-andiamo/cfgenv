package cfgenv

import (
	"os"
	"reflect"
	"regexp"
	"strings"
)

// PrefixOption is an option that can be passed to Load or LoadAs
// and provides a prefix for all env var names
type PrefixOption interface {
	// GetPrefix returns the prefix for all env var names
	GetPrefix() string
}

type prefixOpt struct {
	value string
}

func (p *prefixOpt) GetPrefix() string {
	return p.value
}

// NewPrefix creates a new PrefixOption with the specified prefix
func NewPrefix(prefix string) PrefixOption {
	return &prefixOpt{value: prefix}
}

// SeparatorOption is an option that can be passed to Load or LoadAs
// and provides the separator to be used between prefixes used in env var names
type SeparatorOption interface {
	// GetSeparator returns the separator to be used between prefixes used in env var names
	GetSeparator() string
}

type separatorOpt struct {
	value string
}

func (s *separatorOpt) GetSeparator() string {
	return s.value
}

// NewSeparator creates a new NewSeparator with the specified separator
func NewSeparator(separator string) SeparatorOption {
	return &separatorOpt{
		value: separator,
	}
}

// NamingOption is an option that can be passed to Load or LoadAs
// and provides a means of overriding how env var names are deduced
type NamingOption interface {
	BuildName(prefix string, separator string, fld reflect.StructField, overrideName string) string
}

var defaultNamingOption NamingOption = &namingOption{}

type namingOption struct{}

func (n *namingOption) BuildName(prefix string, separator string, fld reflect.StructField, overrideName string) string {
	name := overrideName
	if name == "" {
		name = toSnakeCase(fld.Name)
	}
	if prefix != "" {
		name = prefix + separator + name
	}
	return name
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToUpper(snake)
}

// Expand creates a default ExpandOption (for use in Load / LoadAs)
//
// Any supplied lookup maps are checked first - if a given env var name, e.g. "${FOO}", is
// not found in lookups then the value is taken from env var
func Expand(lookups ...map[string]string) ExpandOption {
	return &expandOpt{
		lookups: lookups,
	}
}

// ExpandOption is an option that can be passed to Load or LoadAs
// and provides support for expanding environment var values like...
//
//	FOO=${BAR}
type ExpandOption interface {
	// Expand expands the env var value s
	Expand(s string, er EnvReader) string
}

type expandOpt struct {
	lookups []map[string]string
}

func (e *expandOpt) Expand(s string, er EnvReader) string {
	if er == nil {
		er = defaultReader
	}
	return os.Expand(s, func(s string) string {
		return e.expand(s, er)
	})
}

func (e *expandOpt) expand(s string, er EnvReader) string {
	for _, m := range e.lookups {
		if v, ok := m[s]; ok {
			return e.Expand(v, er)
		}
	}
	if v, ok := er.LookupEnv(s); ok {
		return e.Expand(v, er)
	}
	return ""
}
