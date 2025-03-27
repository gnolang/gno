package vm

import (
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestJSONSnakeCase(t *testing.T) {
	t.Parallel()
	for _, typ := range Package.Types {
		assertJSONSnakeCase(t, typ.Type)
	}
}

func assertJSONSnakeCase(t *testing.T, typ reflect.Type) {
	t.Helper()

	switch typ.Kind() {
	case reflect.Array, reflect.Slice, reflect.Pointer:
		assertJSONSnakeCase(t, typ.Elem())
	case reflect.Map:
		assertJSONSnakeCase(t, typ.Key())
		assertJSONSnakeCase(t, typ.Elem())
	case reflect.Struct:
		for i := range typ.NumField() {
			fld := typ.Field(i)
			if !fld.IsExported() {
				continue
			}
			jt := fld.Tag.Get("json")
			if jt == "" {
				if fld.Anonymous {
					assertJSONSnakeCase(t, fld.Type)
					continue
				}
				t.Errorf("field %s.%s does not have a json tag but is exported", typ.Name(), fld.Name)
				continue
			}
			has := strings.ContainsFunc(jt, unicode.IsUpper)
			assert.False(t, has,
				"field %s.%s contains uppercase symbols in json tag", typ.Name(), fld.Name)
			assertJSONSnakeCase(t, fld.Type)
		}
	}
}
