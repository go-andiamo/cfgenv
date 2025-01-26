package cfgenv

import (
	"github.com/go-andiamo/gopt"
	"reflect"
	"time"
)

// CustomSetterOption is an option that can be passed to Load or LoadAs
// and provides support for reading additional struct field types
type CustomSetterOption interface {
	// IsApplicable should return true if the fld type is supported by this custom setter
	IsApplicable(fld reflect.StructField) bool
	// Set sets the field value `v` using the environment var `raw` value
	Set(fld reflect.StructField, v reflect.Value, raw string, present bool) error
}

type dateTimeSetterOption struct {
	format string
}

// NewDatetimeSetter creates a CustomSetterOption that can be passed to Load or LoadAs
// and provides support for reading time.Time fields
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

var dtType = reflect.TypeOf(time.Time{})
var optDtType = reflect.TypeOf(gopt.Optional[time.Time]{})

func (d *dateTimeSetterOption) IsApplicable(fld reflect.StructField) bool {
	return fld.Type == dtType || fld.Type == optDtType
}

func (d *dateTimeSetterOption) Set(fld reflect.StructField, v reflect.Value, raw string, present bool) error {
	dt, err := time.Parse(d.format, raw)
	if err != nil {
		return err
	}
	if fld.Type == dtType {
		v.Set(reflect.ValueOf(dt))
	} else if present {
		av := gopt.Empty[time.Time]().WasSetElseSet(dt)
		v.Set(reflect.ValueOf(*av))
	} else {
		av := gopt.Of[time.Time](dt)
		v.Set(reflect.ValueOf(*av))
	}
	return nil
}

type durationSetterOption struct{}

// NewDurationSetter creates a CustomSetterOption that can be passed to Load or LoadAs
// and provides support for reading time.Duration fields
func NewDurationSetter() CustomSetterOption {
	return &durationSetterOption{}
}

var durationType = reflect.TypeOf(time.Duration(0))
var optDurationType = reflect.TypeOf(gopt.Optional[time.Duration]{})

func (d *durationSetterOption) IsApplicable(fld reflect.StructField) bool {
	return fld.Type == durationType || fld.Type == optDurationType
}

func (d *durationSetterOption) Set(fld reflect.StructField, v reflect.Value, raw string, present bool) error {
	dur, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}
	if fld.Type == durationType {
		v.Set(reflect.ValueOf(dur))
	} else if present {
		av := gopt.Empty[time.Duration]().WasSetElseSet(dur)
		v.Set(reflect.ValueOf(*av))
	} else {
		av := gopt.Of[time.Duration](dur)
		v.Set(reflect.ValueOf(*av))
	}
	return nil
}
