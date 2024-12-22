package toa

import (
	"reflect"
)

var supportedTypes = map[reflect.Type]bool{
	reflect.TypeOf(([]byte)("")):  true,
	reflect.TypeOf((string)("")):  true,
	reflect.TypeOf((bool)(false)): true,
	reflect.TypeOf((byte)(0)):     true,
	reflect.TypeOf((rune)(0)):     true,
	reflect.TypeOf((float32)(0)):  true,
	reflect.TypeOf((float64)(0)):  true,
	reflect.TypeOf((int)(0)):      true,
	reflect.TypeOf((int8)(0)):     true,
	reflect.TypeOf((int16)(0)):    true,
	reflect.TypeOf((int32)(0)):    true,
	reflect.TypeOf((int64)(0)):    true,
	//reflect.TypeOf((uint)(0)):     true,
	reflect.TypeOf((uint8)(0)):  true,
	reflect.TypeOf((uint16)(0)): true,
	reflect.TypeOf((uint32)(0)): true,
	reflect.TypeOf((uint64)(0)): true,
}

func X_toa(args []interface{}) []string {
	var types []string
	for i := range args {
		t := reflect.TypeOf(args[i])
		if !supportedTypes[t] {
			panic("testing: unsupported type to Add")
		}
		types = append(types, t.String())
	}
	return types

}
