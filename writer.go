package cfgenv

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Write writes the current config
//
// the type of T must be a struct
//
// Use any options (such as PrefixOption, SeparatorOption, NamingOption or multiple CustomSetterOption) to alter
// loading behaviour
func Write(w io.Writer, cfg any, options ...any) error {
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
	return write(w, v, o.prefix.GetPrefix(), true, o)
}

// ExampleOf writes an example of the specified T config
//
// the type of T must be a struct
//
// Use any options (such as PrefixOption, SeparatorOption, NamingOption or multiple CustomSetterOption) to alter
// loading behaviour
func ExampleOf[T any](w io.Writer, options ...any) error {
	var cfg T
	return Example(w, &cfg, options...)
}

// Example writes an example of the config
//
// the supplied cfg arg must be a pointer to a struct
//
// Use any options (such as PrefixOption, SeparatorOption, NamingOption or multiple CustomSetterOption) to alter
// loading behaviour
func Example(w io.Writer, cfg any, options ...any) error {
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
	return write(w, v, o.prefix.GetPrefix(), false, o)
}

func write(w io.Writer, v reflect.Value, prefix string, actual bool, options *opts) error {
	seen := map[string]bool{}
	added := map[string]string{}
	err := writeValue(w, v, prefix, actual, options, seen, added)
	if err != nil {
		return err
	}
	for k, v := range added {
		if !seen[k] {
			seen[k] = true
			if _, err = w.Write([]byte(k + "=" + v + "\n")); err != nil {
				return err
			}
		}
	}
	return err
}

func writeValue(w io.Writer, v reflect.Value, prefix string, actual bool, options *opts, seen map[string]bool, added map[string]string) error {
	t := v.Type()
	for f := 0; f < t.NumField(); f++ {
		if fld := t.Field(f); fld.Anonymous {
			ev := v.Field(f)
			if err := writeValue(w, ev, prefix, actual, options, seen, added); err != nil {
				return err
			}
		} else if fld.IsExported() {
			fi, err := getFieldInfo(fld, options)
			if err != nil {
				return err
			}
			name := options.naming.BuildName(prefix, options.separator.GetSeparator(), fld, fi.name)
			if !seen[name] {
				seen[name] = true
				if fi.isStruct {
					fv := v.Field(f)
					if fi.pointer {
						if fv.IsNil() && actual {
							continue
						} else if fv.IsNil() {
							fv = reflect.New(fv.Type().Elem()).Elem()
						} else {
							fv = fv.Elem()
						}
					}
					pfx := addPrefixes(prefix, fi.prefix, options.separator.GetSeparator())
					if err = write(w, fv, pfx, actual, options); err != nil {
						return err
					}
				} else if !actual {
					if !fi.isPrefixedMap {
						if err = writeExampleValue(w, name, v.Field(f), fi); err != nil {
							return err
						}
					}
				} else if fi.isPrefixedMap {
					m := v.Field(f).Interface().(map[string]string)
					for k, v := range m {
						added[k] = v
					}
				} else if err = writeActualValue(w, name, v.Field(f), fi); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeExampleValue(w io.Writer, name string, fv reflect.Value, fi *fieldInfo) error {
	eg := "<value>"
	if fi.hasDefault {
		eg = fi.defaultValue
	} else if fi.customSetter == nil {
		switch fv.Type().Kind() {
		case reflect.String:
			eg = "<string>"
		case reflect.Bool:
			eg = "true|false"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			eg = "0"
		case reflect.Float32, reflect.Float64:
			eg = "0.0"
		case reflect.Slice:
			eg = fmt.Sprintf("value%svalue%s...", fi.delimiter, fi.delimiter)
		case reflect.Map:
			eg = fmt.Sprintf("key%svalue%skey%svalue%s...", fi.separator, fi.delimiter, fi.separator, fi.delimiter)
		}
	}
	_, err := w.Write([]byte(name + "=" + eg + "\n"))
	return err
}

func writeActualValue(w io.Writer, name string, fv reflect.Value, fi *fieldInfo) error {
	eg := "<value>"
	skip := false
	if fi.customSetter == nil {
		if fi.pointer {
			skip = fv.IsNil()
			fv = fv.Elem()
		}
		if !skip {
			switch fv.Type().Kind() {
			case reflect.String:
				eg = fv.String()
			case reflect.Bool:
				if fv.Bool() {
					eg = "true"
				} else {
					eg = "false"
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				eg = fmt.Sprintf("%d", fv.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				eg = fmt.Sprintf("%d", fv.Uint())
			case reflect.Float32:
				eg = strconv.FormatFloat(fv.Float(), 'f', -1, 32)
			case reflect.Float64:
				eg = strconv.FormatFloat(fv.Float(), 'f', -1, 64)
			case reflect.Slice:
				items := make([]string, 0)
				for i := 0; i < fv.Len(); i++ {
					items = append(items, fmt.Sprintf("%v", fv.Index(i).Interface()))
				}
				eg = strings.Join(items, fi.delimiter)
			case reflect.Map:
				items := make([]string, 0)
				for _, mk := range fv.MapKeys() {
					mv := fv.MapIndex(mk)
					items = append(items, fmt.Sprintf("%v%s%v", mk.Interface(), fi.separator, mv.Interface()))
				}
				eg = strings.Join(items, fi.delimiter)
			}
		}
	}
	var err error
	if !skip {
		_, err = w.Write([]byte(name + "=" + eg + "\n"))
	}
	return err
}
