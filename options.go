package cfgenv

import (
	"reflect"
	"regexp"
	"strings"
	"time"
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

// CustomSetterOption is an option that can be passed to Load or LoadAs
// and provides support for reading additional struct field types
type CustomSetterOption interface {
	// IsApplicable should return true if the fld type is supported by this custom setter
	IsApplicable(fld reflect.StructField) bool
	// Set sets the field value `v` using the environment var `raw` value
	Set(fld reflect.StructField, v reflect.Value, raw string) error
}

type dateTimeSetterOption struct {
	format string
}

func NewDatetimeSetter(format string) CustomSetterOption {
	if format == "" {
		return &dateTimeSetterOption{
			format: time.RFC3339,
		}
	}
	return &dateTimeSetterOption{
		format: format,
	}
}

func (d *dateTimeSetterOption) IsApplicable(fld reflect.StructField) bool {
	return fld.Type == dtType
}

func (d *dateTimeSetterOption) Set(fld reflect.StructField, v reflect.Value, raw string) error {
	dt, err := time.Parse(d.format, raw)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(dt))
	return nil
}

var dtType = reflect.TypeOf(time.Time{})
