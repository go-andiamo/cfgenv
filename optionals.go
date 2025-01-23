package cfgenv

import (
	"github.com/go-andiamo/gopt"
	"reflect"
	"strconv"
)

type optionalSetterFn func(v reflect.Value, raw string, present bool) error

var optionalTypeSetters = map[reflect.Type]optionalSetterFn{
	reflect.TypeOf(gopt.Optional[string]{}):  optionalStringSetter,
	reflect.TypeOf(gopt.Optional[bool]{}):    optionalBoolSetter,
	reflect.TypeOf(gopt.Optional[float32]{}): optionalFloat32Setter,
	reflect.TypeOf(gopt.Optional[float64]{}): optionalFloat64Setter,
	reflect.TypeOf(gopt.Optional[int]{}):     optionalIntSetter,
	reflect.TypeOf(gopt.Optional[int8]{}):    optionalInt8Setter,
	reflect.TypeOf(gopt.Optional[int16]{}):   optionalInt16Setter,
	reflect.TypeOf(gopt.Optional[int32]{}):   optionalInt32Setter,
	reflect.TypeOf(gopt.Optional[int64]{}):   optionalInt64Setter,
	reflect.TypeOf(gopt.Optional[uint]{}):    optionalUintSetter,
	reflect.TypeOf(gopt.Optional[uint8]{}):   optionalUint8Setter,
	reflect.TypeOf(gopt.Optional[uint16]{}):  optionalUint16Setter,
	reflect.TypeOf(gopt.Optional[uint32]{}):  optionalUint32Setter,
	reflect.TypeOf(gopt.Optional[uint64]{}):  optionalUint64Setter,
}

func optionalStringSetter(v reflect.Value, raw string, present bool) error {
	if present {
		av := gopt.Empty[string]().WasSetElseSet(raw)
		v.Set(reflect.ValueOf(*av))
	} else {
		av := gopt.Of[string](raw)
		v.Set(reflect.ValueOf(*av))
	}
	return nil
}

func optionalBoolSetter(v reflect.Value, raw string, present bool) error {
	var bv bool
	var err error
	if bv, err = strconv.ParseBool(raw); err == nil {
		if present {
			av := gopt.Empty[bool]().WasSetElseSet(bv)
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[bool](bv)
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalFloat32Setter(v reflect.Value, raw string, present bool) error {
	var iv float64
	var err error
	if iv, err = strconv.ParseFloat(raw, 32); err == nil {
		if present {
			av := gopt.Empty[float32]().WasSetElseSet(float32(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[float32](float32(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalFloat64Setter(v reflect.Value, raw string, present bool) error {
	var iv float64
	var err error
	if iv, err = strconv.ParseFloat(raw, 64); err == nil {
		if present {
			av := gopt.Empty[float64]().WasSetElseSet(iv)
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[float64](iv)
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalIntSetter(v reflect.Value, raw string, present bool) error {
	var iv int
	var err error
	if iv, err = strconv.Atoi(raw); err == nil {
		if present {
			av := gopt.Empty[int]().WasSetElseSet(iv)
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[int](iv)
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalInt8Setter(v reflect.Value, raw string, present bool) error {
	var iv int64
	var err error
	if iv, err = strconv.ParseInt(raw, 10, 8); err == nil {
		if present {
			av := gopt.Empty[int8]().WasSetElseSet(int8(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[int8](int8(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalInt16Setter(v reflect.Value, raw string, present bool) error {
	var iv int64
	var err error
	if iv, err = strconv.ParseInt(raw, 10, 16); err == nil {
		if present {
			av := gopt.Empty[int16]().WasSetElseSet(int16(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[int16](int16(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalInt32Setter(v reflect.Value, raw string, present bool) error {
	var iv int64
	var err error
	if iv, err = strconv.ParseInt(raw, 10, 32); err == nil {
		if present {
			av := gopt.Empty[int32]().WasSetElseSet(int32(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[int32](int32(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalInt64Setter(v reflect.Value, raw string, present bool) error {
	var iv int64
	var err error
	if iv, err = strconv.ParseInt(raw, 10, 64); err == nil {
		if present {
			av := gopt.Empty[int64]().WasSetElseSet(iv)
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[int64](iv)
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalUintSetter(v reflect.Value, raw string, present bool) error {
	var iv uint64
	var err error
	if iv, err = strconv.ParseUint(raw, 10, 0); err == nil {
		if present {
			av := gopt.Empty[uint]().WasSetElseSet(uint(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[uint](uint(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalUint8Setter(v reflect.Value, raw string, present bool) error {
	var iv uint64
	var err error
	if iv, err = strconv.ParseUint(raw, 10, 8); err == nil {
		if present {
			av := gopt.Empty[uint8]().WasSetElseSet(uint8(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[uint8](uint8(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalUint16Setter(v reflect.Value, raw string, present bool) error {
	var iv uint64
	var err error
	if iv, err = strconv.ParseUint(raw, 10, 16); err == nil {
		if present {
			av := gopt.Empty[uint16]().WasSetElseSet(uint16(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[uint16](uint16(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalUint32Setter(v reflect.Value, raw string, present bool) error {
	var iv uint64
	var err error
	if iv, err = strconv.ParseUint(raw, 10, 32); err == nil {
		if present {
			av := gopt.Empty[uint32]().WasSetElseSet(uint32(iv))
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[uint32](uint32(iv))
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}

func optionalUint64Setter(v reflect.Value, raw string, present bool) error {
	var iv uint64
	var err error
	if iv, err = strconv.ParseUint(raw, 10, 64); err == nil {
		if present {
			av := gopt.Empty[uint64]().WasSetElseSet(iv)
			v.Set(reflect.ValueOf(*av))
		} else {
			av := gopt.Of[uint64](iv)
			v.Set(reflect.ValueOf(*av))
		}
	}
	return err
}
