package std

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// ValidateUTF8 returns an error naming the first string in v, in field order,
// that is not valid UTF-8. It walks what amino encodes: through pointers,
// interfaces, structs, slices and arrays, over the exported fields amino does
// not skip, and through the MarshalAmino representation of a type that has
// one rather than its Go fields. Byte slices and byte arrays hold binary
// data, not text, and are not checked.
//
// Walking the representation matters twice over. It is what the signature
// covers, so no string escapes the check by hiding in fields amino skips (a
// MapList's middle items are reachable only through Next). And a decoder may
// alias the Go value (a one-item MapList has Head == Tail), which a field
// walk would revisit once per alias, exponentially in the nesting; the
// representation has no aliases, so the walk costs what encoding v costs.
//
// A transaction's signature covers its amino JSON rendering, and JSON encoding
// turns every invalid UTF-8 byte into U+FFFD. So the signature does not cover
// such bytes: anyone relaying the transaction can swap them for U+FFFD, or
// U+FFFD for any invalid byte, and the signature still verifies, while the
// handlers receive different bytes. The baseapp refuses these transactions
// once their messages pass ValidateBasic, before the ante handler.
func ValidateUTF8(v any) error {
	if err := validateUTF8(reflect.ValueOf(v)); err != nil {
		return err
	}
	return nil
}

// invalidUTF8Error names where the invalid string sits. The path is built
// while unwinding, so a valid value allocates nothing for it.
type invalidUTF8Error struct {
	path  []string // innermost segment first
	cause error    // set when MarshalAmino failed, so nothing could be checked
}

func (e *invalidUTF8Error) Error() string {
	var b strings.Builder
	for i := len(e.path) - 1; i >= 0; i-- {
		b.WriteString(e.path[i])
	}
	where := strings.TrimPrefix(b.String(), ".")
	switch {
	case e.cause != nil && where == "":
		return "cannot check UTF-8: " + e.cause.Error()
	case e.cause != nil:
		return "cannot check UTF-8 in " + where + ": " + e.cause.Error()
	case where == "":
		return "invalid UTF-8 string"
	}
	return "invalid UTF-8 in " + where
}

// aminoReprTypes caches hasAminoRepr per type.
var aminoReprTypes sync.Map // reflect.Type -> bool

// hasAminoRepr reports whether amino encodes rt through its MarshalAmino
// representation, by amino's own rule: a MarshalAmino method on the value.
func hasAminoRepr(rt reflect.Type) bool {
	if v, ok := aminoReprTypes.Load(rt); ok {
		return v.(bool)
	}
	_, ok := rt.MethodByName("MarshalAmino")
	aminoReprTypes.Store(rt, ok)
	return ok
}

func validateUTF8(rv reflect.Value) *invalidUTF8Error {
	if k := rv.Kind(); k != reflect.Pointer && k != reflect.Interface && k != reflect.Invalid &&
		hasAminoRepr(rv.Type()) {
		out := rv.MethodByName("MarshalAmino").Call(nil)
		if err, _ := out[1].Interface().(error); err != nil {
			return &invalidUTF8Error{cause: err}
		}
		return validateUTF8(out[0])
	}
	switch rv.Kind() {
	case reflect.String:
		if !utf8.ValidString(rv.String()) {
			return &invalidUTF8Error{}
		}
	case reflect.Pointer, reflect.Interface:
		if !rv.IsNil() {
			return validateUTF8(rv.Elem())
		}
	case reflect.Struct:
		rt := rv.Type()
		for i := range rt.NumField() {
			f := rt.Field(i)
			if !f.IsExported() || f.Tag.Get("json") == "-" {
				continue // amino neither encodes nor decodes it
			}
			if err := validateUTF8(rv.Field(i)); err != nil {
				err.path = append(err.path, "."+jsonName(f))
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		et := rv.Type().Elem()
		if et.Kind() == reflect.Uint8 {
			return nil // amino writes a byte list as bytes, repr or not
		}
		if !hasAminoRepr(et) {
			switch et.Kind() {
			case reflect.String, reflect.Pointer, reflect.Interface,
				reflect.Struct, reflect.Slice, reflect.Array:
			default:
				return nil // other scalars hold no text
			}
		}
		for i := range rv.Len() {
			if err := validateUTF8(rv.Index(i)); err != nil {
				err.path = append(err.path, "["+strconv.Itoa(i)+"]")
				return err
			}
		}
	}
	return nil
}

// jsonName is the field's name in the signed payload, as amino derives it.
func jsonName(f reflect.StructField) string {
	if name, _, _ := strings.Cut(f.Tag.Get("json"), ","); name != "" {
		return name
	}
	return f.Name
}
