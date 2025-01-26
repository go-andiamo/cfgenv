package cfgenv

import (
	"fmt"
	"github.com/go-andiamo/splitter"
	"reflect"
	"regexp"
	"strings"
)

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
	expand         bool
	noExpand       bool
}

var tagSplitter = splitter.MustCreateSplitter(',', splitter.DoubleQuotes, splitter.SingleQuotes).
	AddDefaultOptions(splitter.Trim(" "), splitter.IgnoreEmpties)
var eqSplitter = splitter.MustCreateSplitter('=', splitter.DoubleQuotes, splitter.SingleQuotes).
	AddDefaultOptions(splitter.Trim(" "))

const (
	tokenDefault   = "default"
	tokenDelim     = "delim"
	tokenDelimiter = "delimiter"
	tokenEncoding  = "encoding"
	tokenExpand    = "expand"
	tokenMatch     = "match"
	tokenName      = "name"
	tokenNoExpand  = "no-expand"
	tokenOptional  = "optional"
	tokenPrefix    = "prefix"
	tokenSep       = "sep"
	tokenSeparator = "separator"
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
					result.prefix = unquoted(pts[1])
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
					if dec, ok := options.decoders[unquoted(pts[1])]; ok {
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
				case tokenExpand:
					result.expand = true
					result.noExpand = false
				case tokenNoExpand:
					result.noExpand = true
					result.expand = false
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
