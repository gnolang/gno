package amino

import (
	"bytes"
	"reflect"
)

//----------------------------------------
// DeepEqual

// DeepEqual returns true if the types are the same and the
// binary amino encoding would be the same.
// TODO: optimize, and support genproto.
func DeepEqual(a, b interface{}) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	return bytes.Equal(
		MustMarshal(a),
		MustMarshal(b),
	)
}
