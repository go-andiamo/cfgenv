package main

import (
	"reflect"
)

type Custom []byte

var customType = reflect.TypeOf(Custom{})

type CustomSetter struct {
}

func (c *CustomSetter) IsApplicable(fld reflect.StructField) bool {
	return fld.Type == customType
}

func (c *CustomSetter) Set(fld reflect.StructField, v reflect.Value, raw string, present bool) error {
	v.Set(reflect.ValueOf(Custom(raw)))
	return nil
}
