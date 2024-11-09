package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParam_Parse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		entry     string
		expected  Param
		expectErr bool
	}{
		{"valid string", "foo.string=hello", Param{key: "foo", kind: "string", value: "hello"}, false},
		{"valid int64", "foo.int64=-1337", Param{key: "foo", kind: "int64", value: int64(-1337)}, false},
		{"valid uint64", "foo.uint64=42", Param{key: "foo", kind: "uint64", value: uint64(42)}, false},
		{"valid bool", "foo.bool=true", Param{key: "foo", kind: "bool", value: true}, false},
		{"valid bytes", "foo.bytes=AAAA", Param{key: "foo", kind: "bytes", value: []byte{0xaa, 0xaa}}, false},
		{"invalid key", "invalidkey=foo", Param{}, true},
		{"invalid kind", "invalid.kind=foo", Param{}, true},
		{"invalid int64", "invalid.int64=foobar", Param{}, true},
		{"invalid uint64", "invalid.uint64=-42", Param{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			param := Param{}
			err := param.Parse(tc.entry)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, param)
			}
		})
	}
}
