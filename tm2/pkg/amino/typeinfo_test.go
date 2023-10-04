package amino

import (
	"reflect"
	"testing"
)

func TestTypeInfoString(t *testing.T) {
	type T struct {
		T *T
	}
	typeInfo := gcdc.newTypeInfoUnregisteredWLocked(reflect.TypeOf(T{}))
	_ = typeInfo.String()
}
