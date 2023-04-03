package json

import (
	"reflect"
	"strings"

	gno "github.com/gnolang/gno/pkgs/gnolang"
)

type StructField interface {
	IsZero() bool
	Name() string
	Value() Value
}

type gnoStructField struct {
	fieldType gno.FieldType
	value     Value
}

type tagOptions string

func (sf gnoStructField) IsZero() bool {
	return sf.value.IsZero()
}

func (sf gnoStructField) Name() string {
	stag := reflect.StructTag(string(sf.fieldType.Tag))
	tag := stag.Get("json")
	name, _ := parseTag(tag)
	// TODO: handle omitempty
	/*
		if !isValidTag(name) {
			name = ""
		}
	*/

	if name == "" {
		name = string(sf.fieldType.Name)
	}

	return name
}

func (sf gnoStructField) Value() Value {
	return sf.value
}

type nativeStructField struct {
	field reflect.StructField
	value Value
}

func (sf nativeStructField) IsZero() bool {
	return sf.value.IsZero()
}

func (sf nativeStructField) Name() string {
	tag := sf.field.Tag.Get("json")
	name, _ := parseTag(tag)
	//TODO: handle omitempty
	/*
		if !isValidTag(name) {
			name = ""
		}
	*/

	if name == "" {
		name = sf.field.Name
	}

	return name
}

func (sf nativeStructField) Value() Value {
	return sf.value
}

func parseTag(tag string) (string, tagOptions) {
	tag, opt, _ := strings.Cut(tag, ",")
	return tag, tagOptions(opt)
}
