package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParam_Parse(t *testing.T) {
	tests := []struct {
		name      string
		entry     string
		expected  Param
		expectErr bool
	}{
		{"valid string", "foo.string=hello", Param{key: "foo.string", stringVal: "hello"}, false},
		{"valid int64", "foo.int64=-1337", Param{key: "foo.int64", int64Val: -1337}, false},
		{"valid uint64", "foo.uint64=42", Param{key: "foo.uint64", uint64Val: 42}, false},
		{"valid bool", "foo.bool=true", Param{key: "foo.bool", boolVal: true}, false},
		{"valid bytes", "foo.bytes=AAAA", Param{key: "foo.bytes", bytesVal: []byte{0xaa, 0xaa}}, false},
		{"invalid key", "invalidkey=foo", Param{}, true},
		{"invalid int64", "invalid.int64=foobar", Param{}, true},
		{"invalid uint64", "invalid.uint64=-42", Param{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
