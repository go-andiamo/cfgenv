package cfgenv

import (
	"errors"
	"fmt"
	"github.com/go-andiamo/splitter"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// LoadAs loads the specified T config struct type from environment vars
//
// the type of T must be a struct
//
// Use any options (such as PrefixOption, SeparatorOption, NamingOption or multiple CustomSetterOption) to alter
// loading behaviour
func LoadAs[T any](options ...any) (*T, error) {
	var cfg T
	if err := Load(&cfg, options...); err == nil {
		return &cfg, nil
	} else {
		return nil, err
	}
}

// Load loads a config struct from environment vars
//
// the supplied cfg arg must be a pointer to a struct
//
// Use any options (such as PrefixOption, SeparatorOption, NamingOption or multiple CustomSetterOption) to alter
// loading behaviour
func Load(cfg any, options ...any) error {
	o, err := buildOpts(options...)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return errors.New("cfg not a pointer")
	} else {
		v = v.Elem()
		if v.Kind() != reflect.Struct {
			return errors.New("cfg not a struct")
		}
	}
	return loadStruct(v, o.prefix.GetPrefix(), o)
}

func buildOpts(options ...any) (*opts, error) {
	result := &opts{
		prefix:    NewPrefix(""),
		separator: NewSeparator("_"),
		naming:    defaultNamingOption,
	}
	countPfx := 0
	countSep := 0
	countNm := 0
	for _, o := range options {
		if o != nil {
			used := false
			if pfx, ok := o.(PrefixOption); ok {
				used = true
				result.prefix = pfx
				countPfx++
			}
			if sep, ok := o.(SeparatorOption); ok {
				used = true
				result.separator = sep
				countSep++
			}
			if nm, ok := o.(NamingOption); ok {
				used = true
				result.naming = nm
				countNm++
			}
			if cs, ok := o.(CustomSetterOption); ok {
				used = true
				result.customs = append(result.customs, cs)
			}
			if !used {
				return nil, errors.New("invalid option")
			}
		}
	}
	if countPfx > 1 {
		return nil, errors.New("multiple prefix options")
	}
	if countSep > 1 {
		return nil, errors.New("multiple separator options")
	}
	if countNm > 1 {
		return nil, errors.New("multiple naming options")
	}
	return result, nil
}

type opts struct {
	prefix    PrefixOption
	separator SeparatorOption
	naming    NamingOption
	customs   []CustomSetterOption
}

func loadStruct(v reflect.Value, prefix string, options *opts) error {
	t := v.Type()
	for f := 0; f < t.NumField(); f++ {
		if fld := t.Field(f); fld.IsExported() {
			fi, err := getFieldInfo(fld, options)
			if err != nil {
				return err
			}
			name := options.naming.BuildName(prefix, options.separator.GetSeparator(), fld, fi.name)
			if fi.customSetter != nil {
				raw, ok := os.LookupEnv(name)
				if !ok && !fi.optional {
					return fmt.Errorf("missing env var '%s'", name)
				} else if !ok && fi.hasDefault {
					raw = fi.defaultValue
				}
				if err = fi.customSetter.Set(fld, v.Field(f), raw); err != nil {
					return err
				}
			} else if fi.prefixedMap {
				pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
				setPrefixMap(v.Field(f), pfx)
			} else if fld.Type.Kind() == reflect.Struct {
				fv := v.Field(f)
				pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
				if err = loadStruct(fv, pfx, options); err != nil {
					return err
				}
			} else {
				raw, ok := os.LookupEnv(name)
				if !ok && !fi.optional {
					return fmt.Errorf("missing env var '%s'", name)
				} else if !ok && fi.hasDefault {
					raw = fi.defaultValue
				} else if !ok && fi.pointer {
					continue
				}
				if err = setValue(name, raw, fld, v.Field(f)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var durationType = reflect.TypeOf(time.Duration(0))

func setValue(name string, raw string, fld reflect.StructField, fv reflect.Value) (err error) {
	k := fv.Type().Kind()
	isPtr := false
	if k == reflect.Pointer {
		isPtr = true
		k = fv.Type().Elem().Kind()
	}
	switch k {
	case reflect.String:
		if isPtr {
			fv.Set(reflect.ValueOf(&raw))
		} else {
			fv.Set(reflect.ValueOf(raw))
		}
	case reflect.Bool:
		if b, bErr := strconv.ParseBool(raw); bErr == nil {
			if isPtr {
				fv.Set(reflect.ValueOf(&b))
			} else {
				fv.Set(reflect.ValueOf(b))
			}
		} else {
			return fmt.Errorf("env var '%s' is not a bool", name)
		}
	case reflect.Int:
		err = setIntValue[int](name, raw, fv, isPtr)
	case reflect.Int8:
		err = setIntValue[int8](name, raw, fv, isPtr)
	case reflect.Int16:
		err = setIntValue[int16](name, raw, fv, isPtr)
	case reflect.Int32:
		err = setIntValue[int32](name, raw, fv, isPtr)
	case reflect.Int64:
		if fv.Type() == durationType {
			err = setIntValue[time.Duration](name, raw, fv, isPtr)
		} else {
			err = setIntValue[int64](name, raw, fv, isPtr)
		}
	case reflect.Uint:
		err = setUintValue[uint](name, raw, fv, isPtr)
	case reflect.Uint8:
		err = setUintValue[uint8](name, raw, fv, isPtr)
	case reflect.Uint16:
		err = setUintValue[uint16](name, raw, fv, isPtr)
	case reflect.Uint32:
		err = setUintValue[uint32](name, raw, fv, isPtr)
	case reflect.Uint64:
		err = setUintValue[uint64](name, raw, fv, isPtr)
	case reflect.Float32:
		err = setFloatValue[float32](name, raw, fv, isPtr)
	case reflect.Float64:
		err = setFloatValue[float64](name, raw, fv, isPtr)
	case reflect.Slice:
		err = setSlice(name, raw, fld, fv)
	case reflect.Map:
		err = setMap(name, raw, fld, fv)
	}
	return
}

func setSlice(name string, raw string, fld reflect.StructField, fv reflect.Value) error {
	if raw != "" {
		vs := strings.Split(raw, ",")
		sl := reflect.MakeSlice(fv.Type(), len(vs), len(vs))
		for i, v := range vs {
			if err := setValue(name, v, fld, sl.Index(i)); err != nil {
				return err
			}
		}
		fv.Set(sl)
	}
	return nil
}

func setMap(name string, raw string, fld reflect.StructField, fv reflect.Value) error {
	if raw != "" {
		vs := strings.Split(raw, ",")
		m := reflect.MakeMap(fv.Type())
		kt := fv.Type().Key()
		vt := fv.Type().Elem()
		for _, v := range vs {
			kvp := strings.Split(v, ":")
			if len(kvp) != 2 {
				return fmt.Errorf("env var '%s' contains invalid key/value pair - %s", name, v)
			}
			kv := reflect.New(kt).Elem()
			if err := setValue(name, kvp[0], fld, kv); err != nil {
				return err
			}
			vv := reflect.New(vt).Elem()
			if err := setValue(name, kvp[1], fld, vv); err != nil {
				return err
			}
			m.SetMapIndex(kv, vv)
		}
		fv.Set(m)
	}
	return nil
}

func setPrefixMap(fv reflect.Value, prefix string) {
	m := map[string]string{}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, prefix) {
			ev := strings.SplitN(e, "=", 2)
			m[ev[0][len(prefix):]] = ev[1]
		}
	}
	fv.Set(reflect.ValueOf(m))
}

func setFloatValue[T float32 | float64](name string, raw string, fv reflect.Value, isPtr bool) error {
	if f, err := strconv.ParseFloat(raw, getBitSize(fv, isPtr)); err == nil {
		if isPtr {
			pv := T(f)
			fv.Set(reflect.ValueOf(&pv))
		} else {
			fv.Set(reflect.ValueOf(T(f)))
		}
		return nil
	} else {
		return fmt.Errorf("env var '%s' is not a float", name)
	}
}

func setIntValue[T int | int8 | int16 | int32 | int64 | time.Duration](name string, raw string, fv reflect.Value, isPtr bool) error {
	if i, err := strconv.ParseInt(raw, 0, getBitSize(fv, isPtr)); err == nil {
		if isPtr {
			pv := T(i)
			fv.Set(reflect.ValueOf(&pv))
		} else {
			fv.Set(reflect.ValueOf(T(i)))
		}
		return nil
	} else {
		return fmt.Errorf("env var '%s' is not an int", name)
	}
}

func setUintValue[T uint | uint8 | uint16 | uint32 | uint64](name string, raw string, fv reflect.Value, isPtr bool) error {
	if i, err := strconv.ParseUint(raw, 0, getBitSize(fv, isPtr)); err == nil {
		if isPtr {
			pv := T(i)
			fv.Set(reflect.ValueOf(&pv))
		} else {
			fv.Set(reflect.ValueOf(T(i)))
		}
		return nil
	} else {
		return fmt.Errorf("env var '%s' is not an uint", name)
	}
}

func getBitSize(fv reflect.Value, isPtr bool) int {
	if isPtr {
		return fv.Type().Elem().Bits()
	}
	return fv.Type().Bits()
}

func addPrefixes(currPfx, addPfx string, separator string) string {
	if currPfx != "" && addPfx != "" {
		return currPfx + separator + addPfx
	} else if addPfx != "" {
		return addPfx
	}
	return currPfx
}

var tagSplitter = splitter.MustCreateSplitter(',', splitter.DoubleQuotes, splitter.SingleQuotes).
	AddDefaultOptions(splitter.Trim(" "), splitter.IgnoreEmpties)

func getFieldInfo(fld reflect.StructField, options *opts) (*fieldInfo, error) {
	isPtr, custom, err := checkFieldType(fld, options)
	if err != nil {
		return nil, err
	}
	result := &fieldInfo{
		pointer:      isPtr,
		optional:     isPtr,
		customSetter: custom,
	}
	if tag, ok := fld.Tag.Lookup("env"); ok {
		parts, err := tagSplitter.Split(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid tag '%s' on field '%s'", tag, fld.Name)
		}
		for _, s := range parts {
			if strings.Contains(s, "=") {
				if pts := strings.Split(s, "="); len(pts) == 2 {
					switch pts[0] {
					case "default":
						result.hasDefault = true
						result.defaultValue = unquoted(pts[1])
						continue
					case "prefix":
						result.prefix = pts[1]
						if fld.Type.Kind() == reflect.Map {
							result.prefixedMap = fld.Type.Elem().Kind() == reflect.String && fld.Type.Key().Kind() == reflect.String
						}
						if !result.prefixedMap && fld.Type.Kind() != reflect.Struct {
							return nil, fmt.Errorf("cannot use env tag 'prefix' on field '%s' (only for structs or map[string]string)", fld.Name)
						}
						continue
					}
				}
				return nil, fmt.Errorf("invalid tag '%s' on field '%s'", s, fld.Name)
			} else {
				switch s {
				case "optional":
					result.optional = true
				default:
					result.name = unquoted(s)
				}
			}
		}
	}
	return result, nil
}

func unquoted(s string) string {
	if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) {
		return s[1 : len(s)-1]
	}
	return s
}

func checkFieldType(fld reflect.StructField, options *opts) (isPtr bool, custom CustomSetterOption, err error) {
	k := fld.Type.Kind()
	if isPtr = k == reflect.Pointer; isPtr {
		k = fld.Type.Elem().Kind()
	}
	for _, c := range options.customs {
		if ok := c.IsApplicable(fld); ok {
			custom = c
			break
		}
	}
	if isNativeType(k) {
		return
	}
	switch k {
	case reflect.Slice:
		if isPtr {
			err = fmt.Errorf("field '%s' has unsupported type - %s", fld.Name, fld.Type.String())
		} else {
			// check slice item type...
			it := fld.Type.Elem()
			if it.Kind() == reflect.Pointer {
				it = it.Elem()
			}
			if !isNativeType(it.Kind()) {
				err = fmt.Errorf("field '%s' has unsupported slice item type", fld.Name)
			}
		}
	case reflect.Map:
		if isPtr {
			err = fmt.Errorf("field '%s' has unsupported type - %s", fld.Name, fld.Type.String())
		} else {
			// check map item type...
			it := fld.Type.Elem()
			if it.Kind() == reflect.Pointer {
				it = it.Elem()
			}
			if !isNativeType(it.Kind()) {
				err = fmt.Errorf("field '%s' has unsupported map item type", fld.Name)
			} else {
				// check map key type...
				it = fld.Type.Key()
				if !isNativeType(it.Kind()) {
					err = fmt.Errorf("field '%s' has unsupported map key type", fld.Name)
				}
			}
		}
	case reflect.Struct:
		if isPtr {
			err = fmt.Errorf("field '%s' has unsupported embedded struct ptr", fld.Name)
		}
	default:
		if custom == nil {
			err = fmt.Errorf("field '%s' has unsupported type - %s", fld.Name, fld.Type.String())
		}
	}
	return
}

func isNativeType(k reflect.Kind) bool {
	switch k {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

type fieldInfo struct {
	name         string
	optional     bool
	pointer      bool
	hasDefault   bool
	defaultValue string
	prefix       string
	prefixedMap  bool
	customSetter CustomSetterOption
}
