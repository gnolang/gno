package params

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
		{"valid string", "foo.string=hello", Param{Key: "foo", Type: "string", Value: "hello"}, false},
		{"valid int64", "foo.int64=-1337", Param{Key: "foo", Type: "int64", Value: int64(-1337)}, false},
		{"valid uint64", "foo.uint64=42", Param{Key: "foo", Type: "uint64", Value: uint64(42)}, false},
		{"valid bool", "foo.bool=true", Param{Key: "foo", Type: "bool", Value: true}, false},
		{"valid bytes", "foo.bytes=AAAA", Param{Key: "foo", Type: "bytes", Value: []byte{0xaa, 0xaa}}, false},
		{"valid strings", "foo.strings=some,strings", Param{Key: "foo", Type: "strings", Value: []string{"some", "strings"}}, false},
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
