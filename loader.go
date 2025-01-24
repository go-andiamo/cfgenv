package cfgenv

import (
	"errors"
	"fmt"
	"github.com/go-andiamo/splitter"
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
					return nil, errors.New("multiple expand options")
				}
				result.expand = ot
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
	expand    ExpandOption
	customs   []CustomSetterOption
	decoders  map[string]Decoder
	reader    EnvReader
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
					if options.expand != nil {
						raw = options.expand.Expand(raw, options.reader)
					}
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
				if options.expand != nil {
					raw = options.expand.Expand(raw, options.reader)
				}
				if ok && fi.decoder != nil {
					if raw, err = fi.decoder.Decode(raw); err != nil {
						return fmt.Errorf("unable to decode env var '%s' (encoding: '%s'): %s", name, fi.decoder.Encoding(), err.Error())
					}
				}
				if err = fi.customSetter.Set(fld, v.Field(f), raw, ok); err != nil {
					return err
				}
			case fi.isMatchedMap && fi.isPrefixedMap:
				pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
				setPrefixMatchMap(v.Field(f), fi.matchRegex, pfx, options)
			case fi.isMatchedMap:
				setMatchMap(v.Field(f), fi.matchRegex, options)
			case fi.isPrefixedMap:
				pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
				setPrefixMap(v.Field(f), pfx, options)
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
				if options.expand != nil {
					raw = options.expand.Expand(raw, options.reader)
				}
				if ok && fi.decoder != nil {
					if raw, err = fi.decoder.Decode(raw); err != nil {
						return fmt.Errorf("unable to decode env var '%s' (encoding: '%s'): %s", name, fi.decoder.Encoding(), err.Error())
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

var durationType = reflect.TypeOf(time.Duration(0))

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

func setPrefixMap(fv reflect.Value, prefix string, options *opts) {
	m := map[string]string{}
	for _, e := range options.reader.Environ() {
		if strings.HasPrefix(e, prefix) {
			ev := strings.SplitN(e, "=", 2)
			if options.expand != nil {
				m[ev[0][len(prefix):]] = options.expand.Expand(ev[1], options.reader)
			} else {
				m[ev[0][len(prefix):]] = ev[1]
			}
		}
	}
	fv.Set(reflect.ValueOf(m))
}

func setPrefixMatchMap(fv reflect.Value, rx *regexp.Regexp, prefix string, options *opts) {
	m := map[string]string{}
	for _, e := range options.reader.Environ() {
		if strings.HasPrefix(e, prefix) {
			ev := strings.SplitN(e, "=", 2)
			if name := ev[0][len(prefix):]; rx.MatchString(name) {
				if options.expand != nil {
					m[name] = options.expand.Expand(ev[1], options.reader)
				} else {
					m[name] = ev[1]
				}
			}
		}
	}
	fv.Set(reflect.ValueOf(m))
}

func setMatchMap(fv reflect.Value, rx *regexp.Regexp, options *opts) {
	m := map[string]string{}
	for _, e := range options.reader.Environ() {
		ev := strings.SplitN(e, "=", 2)
		if rx.MatchString(ev[0]) {
			if options.expand != nil {
				m[ev[0]] = options.expand.Expand(ev[1], options.reader)
			} else {
				m[ev[0]] = ev[1]
			}
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

var tagSplitter = splitter.MustCreateSplitter(',', splitter.DoubleQuotes, splitter.SingleQuotes).
	AddDefaultOptions(splitter.Trim(" "), splitter.IgnoreEmpties)
var eqSplitter = splitter.MustCreateSplitter('=', splitter.DoubleQuotes, splitter.SingleQuotes).
	AddDefaultOptions(splitter.Trim(" "))

const (
	tokenDefault   = "default"
	tokenPrefix    = "prefix"
	tokenMatch     = "match"
	tokenSeparator = "separator"
	tokenSep       = "sep"
	tokenDelimiter = "delimiter"
	tokenDelim     = "delim"
	tokenOptional  = "optional"
	tokenEncoding  = "encoding"
	tokenName      = "name"
)

func getFieldInfo(fld reflect.StructField, options *opts) (*fieldInfo, error) {
	result, err := checkFieldType(fld, options)
	if err != nil {
		return nil, err
	}
	if tag, ok := fld.Tag.Lookup("env"); ok {
		parts, err := tagSplitter.Split(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid tag '%s' on field '%s'", tag, fld.Name)
		}
		for _, s := range parts {
			if pts, _ := eqSplitter.Split(s); len(pts) == 2 {
				switch pts[0] {
				case tokenName:
					result.name = unquoted(pts[1])
					continue
				case tokenDefault:
					result.hasDefault = true
					result.defaultValue = unquoted(pts[1])
					continue
				case tokenPrefix:
					result.prefix = pts[1]
					if fld.Type.Kind() == reflect.Map {
						result.isPrefixedMap = fld.Type.Elem().Kind() == reflect.String && fld.Type.Key().Kind() == reflect.String
					}
					if !result.isPrefixedMap && !result.isStruct {
						return nil, fmt.Errorf("cannot use env tag 'prefix' on field '%s' (only for structs or map[string]string)", fld.Name)
					}
					continue
				case tokenMatch:
					if fld.Type.Kind() == reflect.Map {
						result.isMatchedMap = fld.Type.Elem().Kind() == reflect.String && fld.Type.Key().Kind() == reflect.String
					}
					if !result.isMatchedMap {
						return nil, fmt.Errorf("cannot use env tag 'match' on field '%s' (only for map[string]string)", fld.Name)
					}
					rxs := unquoted(pts[1])
					if result.matchRegex, err = regexp.Compile(rxs); err != nil {
						return nil, fmt.Errorf("env tag 'match' on field '%s' - invalid regexp: %s", fld.Name, err.Error())
					}
					continue
				case tokenSeparator, tokenSep:
					result.separator = unquoted(pts[1])
					continue
				case tokenDelimiter, tokenDelim:
					result.delimiter = unquoted(pts[1])
					continue
				case tokenEncoding:
					if dec, ok := options.decoders[pts[1]]; ok {
						result.decoder = dec
						continue
					} else {
						return nil, fmt.Errorf("unknown encoding '%s' on field '%s'", pts[1], fld.Name)
					}
				}
				return nil, fmt.Errorf("invalid tag '%s' on field '%s'", s, fld.Name)
			} else if len(pts) == 1 {
				switch s {
				case tokenOptional:
					result.optional = true
				case tokenDefault, tokenPrefix, tokenSeparator, tokenSep, tokenDelimiter, tokenDelim, tokenMatch, tokenEncoding:
					return nil, fmt.Errorf("cannot use env tag '%s' without value on field '%s' (use quotes if necessary)", s, fld.Name)
				default:
					result.name = unquoted(s)
				}
			} else {
				return nil, fmt.Errorf("invalid tag '%s' on field '%s'", s, fld.Name)
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

func checkFieldType(fld reflect.StructField, options *opts) (*fieldInfo, error) {
	isPtr := false
	k := fld.Type.Kind()
	if isPtr = k == reflect.Pointer; isPtr {
		k = fld.Type.Elem().Kind()
	}
	result := &fieldInfo{
		pointer:   isPtr,
		optional:  isPtr,
		separator: ":",
		delimiter: ",",
	}
	for _, c := range options.customs {
		if ok := c.IsApplicable(fld); ok {
			result.customSetter = c
			return result, nil
		}
	}
	if setFn, ok := optionalTypeSetters[fld.Type]; ok {
		result.optional = true
		result.optionalSetter = setFn
		return result, nil
	}
	if isNativeType(k) {
		return result, nil
	}
	switch k {
	case reflect.Slice:
		if isPtr {
			return nil, fmt.Errorf("field '%s' has unsupported type - %s", fld.Name, fld.Type.String())
		} else {
			// check slice item type...
			it := fld.Type.Elem()
			if it.Kind() == reflect.Pointer {
				it = it.Elem()
			}
			if !isNativeType(it.Kind()) {
				return nil, fmt.Errorf("field '%s' has unsupported slice item type", fld.Name)
			}
		}
	case reflect.Map:
		if isPtr {
			return nil, fmt.Errorf("field '%s' has unsupported type - %s", fld.Name, fld.Type.String())
		} else {
			// check map item type...
			it := fld.Type.Elem()
			if it.Kind() == reflect.Pointer {
				it = it.Elem()
			}
			if !isNativeType(it.Kind()) {
				return nil, fmt.Errorf("field '%s' has unsupported map item type", fld.Name)
			} else {
				// check map key type...
				it = fld.Type.Key()
				if !isNativeType(it.Kind()) {
					return nil, fmt.Errorf("field '%s' has unsupported map key type", fld.Name)
				}
			}
		}
	case reflect.Struct:
		result.isStruct = true
	default:
		return nil, fmt.Errorf("field '%s' has unsupported type - %s", fld.Name, fld.Type.String())
	}
	return result, nil
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
	name           string
	optional       bool
	pointer        bool
	hasDefault     bool
	defaultValue   string
	prefix         string
	isStruct       bool
	isPrefixedMap  bool
	isMatchedMap   bool
	matchRegex     *regexp.Regexp
	customSetter   CustomSetterOption
	optionalSetter optionalSetterFn
	decoder        Decoder
	separator      string
	delimiter      string
}
