package cfgenv

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LoadAs loads the specified T config struct type from environment vars
//
// the type of T must be a struct
//
// Use any options (such as PrefixOption, SeparatorOption, NamingOption, EnvReader, Decoder or multiple CustomSetterOption) to alter
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
// Use any options (such as PrefixOption, SeparatorOption, NamingOption, EnvReader, Decoder or multiple CustomSetterOption) to alter
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
		decoders: map[string]Decoder{
			encodingBase64:       NewBase64Decoder(),
			encodingBase64Url:    NewBase64UrlDecoder(),
			encodingRawBase64:    NewRawBase64Decoder(),
			encodingRawBase64Url: NewRawBase64UrlDecoder(),
		},
		reader: defaultReader,
	}
	pfx := false
	sep := false
	name := false
	expand := false
	reader := false
	for _, o := range options {
		if o != nil {
			switch ot := o.(type) {
			case PrefixOption:
				if pfx {
					return nil, errors.New("multiple prefix options")
				}
				result.prefix = ot
				pfx = true
			case SeparatorOption:
				if sep {
					return nil, errors.New("multiple separator options")
				}
				result.separator = ot
				sep = true
			case NamingOption:
				if name {
					return nil, errors.New("multiple naming options")
				}
				result.naming = ot
				name = true
			case ExpandOption:
				if expand {
					return nil, errors.New("multiple expander options")
				}
				result.expander = ot
				expand = true
			case EnvReader:
				if reader {
					return nil, errors.New("multiple reader options")
				}
				result.reader = ot
				reader = true
			case CustomSetterOption:
				result.customs = append(result.customs, ot)
			case Decoder:
				result.decoders[ot.Encoding()] = ot
			default:
				return nil, errors.New("invalid option")
			}
		}
	}
	return result, nil
}

type opts struct {
	prefix    PrefixOption
	separator SeparatorOption
	naming    NamingOption
	expander  ExpandOption
	customs   []CustomSetterOption
	decoders  map[string]Decoder
	reader    EnvReader
}

func (o *opts) expand(s string, fi *fieldInfo) string {
	if !fi.noExpand && o.expander != nil {
		return o.expander.Expand(s, o.reader)
	} else if fi.expand {
		return defaultExpander.Expand(s, o.reader)
	}
	return s
}

func loadStruct(v reflect.Value, prefix string, options *opts) error {
	t := v.Type()
	for f := 0; f < t.NumField(); f++ {
		if fld := t.Field(f); fld.Anonymous {
			ev := v.Field(f)
			if err := loadStruct(ev, prefix, options); err != nil {
				return err
			}
		} else if fld.IsExported() {
			fi, err := getFieldInfo(fld, options)
			if err != nil {
				return err
			}
			name := options.naming.BuildName(prefix, options.separator.GetSeparator(), fld, fi.name)
			switch {
			case fi.optionalSetter != nil:
				if raw, ok := options.reader.LookupEnv(name); ok {
					raw = options.expand(raw, fi)
					if fi.decoder != nil {
						if raw, err = fi.decoder.Decode(raw); err != nil {
							return fmt.Errorf("unable to decode env var '%s' (encoding: '%s'): %s", name, fi.decoder.Encoding(), err.Error())
						}
					}
					if err = fi.optionalSetter(v.Field(f), raw, true); err != nil {
						return err
					}
				} else if fi.hasDefault {
					if err = fi.optionalSetter(v.Field(f), fi.defaultValue, false); err != nil {
						return err
					}
				}
			case fi.customSetter != nil:
				raw, ok := options.reader.LookupEnv(name)
				if !ok && !fi.optional {
					return fmt.Errorf("missing env var '%s'", name)
				} else if !ok && fi.hasDefault {
					raw = fi.defaultValue
				}
				if ok {
					raw = options.expand(raw, fi)
					if fi.decoder != nil {
						if raw, err = fi.decoder.Decode(raw); err != nil {
							return fmt.Errorf("unable to decode env var '%s' (encoding: '%s'): %s", name, fi.decoder.Encoding(), err.Error())
						}
					}
				}
				if err = fi.customSetter.Set(fld, v.Field(f), raw, ok); err != nil {
					return err
				}
			case fi.isMatchedMap && fi.isPrefixedMap:
				pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
				setPrefixMatchMap(v.Field(f), fi.matchRegex, pfx, fi, options)
			case fi.isMatchedMap:
				setMatchMap(v.Field(f), fi.matchRegex, fi, options)
			case fi.isPrefixedMap:
				pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
				setPrefixMap(v.Field(f), pfx, fi, options)
			case fi.isStruct:
				fv := v.Field(f)
				if fi.pointer {
					fvp := reflect.New(fv.Type().Elem())
					fv.Set(fvp)
					fv = fvp.Elem()
				}
				pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
				if err = loadStruct(fv, pfx, options); err != nil {
					return err
				}
			default:
				raw, ok := options.reader.LookupEnv(name)
				if !ok && !fi.optional {
					return fmt.Errorf("missing env var '%s'", name)
				} else if !ok && fi.hasDefault {
					raw = fi.defaultValue
				} else if !ok && fi.pointer {
					continue
				}
				if ok {
					raw = options.expand(raw, fi)
					if fi.decoder != nil {
						if raw, err = fi.decoder.Decode(raw); err != nil {
							return fmt.Errorf("unable to decode env var '%s' (encoding: '%s'): %s", name, fi.decoder.Encoding(), err.Error())
						}
					}
				}
				if err = setValue(name, raw, fld, fi, v.Field(f)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setValue(name string, raw string, fld reflect.StructField, fi *fieldInfo, fv reflect.Value) (err error) {
	k := fv.Type().Kind()
	if fi.pointer {
		k = fv.Type().Elem().Kind()
	}
	switch k {
	case reflect.String:
		setStringValue(raw, fv, fi.pointer)
	case reflect.Bool:
		err = setBoolValue(name, raw, fv, fi.pointer)
	case reflect.Int:
		err = setIntValue[int](name, raw, fv, fi.pointer)
	case reflect.Int8:
		err = setIntValue[int8](name, raw, fv, fi.pointer)
	case reflect.Int16:
		err = setIntValue[int16](name, raw, fv, fi.pointer)
	case reflect.Int32:
		err = setIntValue[int32](name, raw, fv, fi.pointer)
	case reflect.Int64:
		if fv.Type() == durationType {
			err = setIntValue[time.Duration](name, raw, fv, fi.pointer)
		} else {
			err = setIntValue[int64](name, raw, fv, fi.pointer)
		}
	case reflect.Uint:
		err = setUintValue[uint](name, raw, fv, fi.pointer)
	case reflect.Uint8:
		err = setUintValue[uint8](name, raw, fv, fi.pointer)
	case reflect.Uint16:
		err = setUintValue[uint16](name, raw, fv, fi.pointer)
	case reflect.Uint32:
		err = setUintValue[uint32](name, raw, fv, fi.pointer)
	case reflect.Uint64:
		err = setUintValue[uint64](name, raw, fv, fi.pointer)
	case reflect.Float32:
		err = setFloatValue[float32](name, raw, fv, fi.pointer)
	case reflect.Float64:
		err = setFloatValue[float64](name, raw, fv, fi.pointer)
	case reflect.Slice:
		err = setSlice(name, raw, fld, fi, fv)
	case reflect.Map:
		err = setMap(name, raw, fld, fi, fv)
	}
	return
}

func setSlice(name string, raw string, fld reflect.StructField, fi *fieldInfo, fv reflect.Value) error {
	if raw != "" {
		if fld.Type.Elem().Kind() == reflect.Uint8 {
			sl := []byte(raw)
			fv.Set(reflect.ValueOf(sl))
			return nil
		}
		vs := strings.Split(raw, fi.delimiter)
		sl := reflect.MakeSlice(fv.Type(), len(vs), len(vs))
		for i, v := range vs {
			if err := setValue(name, v, fld, fi, sl.Index(i)); err != nil {
				return err
			}
		}
		fv.Set(sl)
	}
	return nil
}

func setMap(name string, raw string, fld reflect.StructField, fi *fieldInfo, fv reflect.Value) error {
	if raw != "" {
		vs := strings.Split(raw, fi.delimiter)
		m := reflect.MakeMap(fv.Type())
		kt := fv.Type().Key()
		vt := fv.Type().Elem()
		for _, v := range vs {
			kvp := strings.Split(v, fi.separator)
			if len(kvp) != 2 {
				return fmt.Errorf("env var '%s' contains invalid key/value pair - %s", name, v)
			}
			kv := reflect.New(kt).Elem()
			if err := setValue(name, kvp[0], fld, fi, kv); err != nil {
				return err
			}
			vv := reflect.New(vt).Elem()
			if err := setValue(name, kvp[1], fld, fi, vv); err != nil {
				return err
			}
			m.SetMapIndex(kv, vv)
		}
		fv.Set(m)
	}
	return nil
}

func setPrefixMap(fv reflect.Value, prefix string, fi *fieldInfo, options *opts) {
	m := map[string]string{}
	for _, e := range options.reader.Environ() {
		if strings.HasPrefix(e, prefix) {
			ev := strings.SplitN(e, "=", 2)
			m[ev[0][len(prefix):]] = options.expand(ev[1], fi)
		}
	}
	fv.Set(reflect.ValueOf(m))
}

func setPrefixMatchMap(fv reflect.Value, rx *regexp.Regexp, prefix string, fi *fieldInfo, options *opts) {
	m := map[string]string{}
	for _, e := range options.reader.Environ() {
		if strings.HasPrefix(e, prefix) {
			ev := strings.SplitN(e, "=", 2)
			if name := ev[0][len(prefix):]; rx.MatchString(name) {
				m[name] = options.expand(ev[1], fi)
			}
		}
	}
	fv.Set(reflect.ValueOf(m))
}

func setMatchMap(fv reflect.Value, rx *regexp.Regexp, fi *fieldInfo, options *opts) {
	m := map[string]string{}
	for _, e := range options.reader.Environ() {
		ev := strings.SplitN(e, "=", 2)
		if rx.MatchString(ev[0]) {
			m[ev[0]] = options.expand(ev[1], fi)
		}
	}
	fv.Set(reflect.ValueOf(m))
}

func setStringValue(raw string, fv reflect.Value, isPtr bool) {
	if isPtr {
		fv.Set(reflect.ValueOf(&raw))
	} else {
		fv.Set(reflect.ValueOf(raw))
	}
}

func setBoolValue(name string, raw string, fv reflect.Value, isPtr bool) error {
	if b, bErr := strconv.ParseBool(raw); bErr == nil {
		if isPtr {
			fv.Set(reflect.ValueOf(&b))
		} else {
			fv.Set(reflect.ValueOf(b))
		}
		return nil
	} else {
		return fmt.Errorf("env var '%s' is not a bool", name)
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
		return fmt.Errorf("env var '%s' is not a uint", name)
	}
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
