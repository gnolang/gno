package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKindString(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		desc     string
		in       Kind
		expected string
	}{
		{"InvalidKind", 0, "InvalidKind"},
		{"BoolKind", 1, "BoolKind"},
		{"StringKind", 2, "StringKind"},
		{"IntKind", 3, "IntKind"},
		{"Int8Kind", 4, "Int8Kind"},
		{"Int16Kind", 5, "Int16Kind"},
		{"Int32Kind", 6, "Int32Kind"},
		{"Int64Kind", 7, "Int64Kind"},
		{"UintKind", 8, "UintKind"},
		{"Uint8Kind", 9, "Uint8Kind"},
		{"Uint16Kind", 10, "Uint16Kind"},
		{"Uint32Kind", 11, "Uint32Kind"},
		{"Uint64Kind", 12, "Uint64Kind"},
		{"Float32Kind", 13, "Float32Kind"},
		{"Float64Kind", 14, "Float64Kind"},
		{"BigintKind", 15, "BigintKind"},
		{"BigdecKind", 16, "BigdecKind"},
		{"ArrayKind", 17, "ArrayKind"},
		{"SliceKind", 18, "SliceKind"},
		{"PointerKind", 19, "PointerKind"},
		{"StructKind", 20, "StructKind"},
		{"PackageKind", 21, "PackageKind"},
		{"InterfaceKind", 22, "InterfaceKind"},
		{"ChanKind", 23, "ChanKind"},
		{"FuncKind", 24, "FuncKind"},
		{"MapKind", 25, "MapKind"},
		{"TypeKind", 26, "TypeKind"},
		{"BlockKind", 27, "BlockKind"},
		{"TupleKind", 28, "TupleKind"},
		{"RefTypeKind", 29, "RefTypeKind"},
		{"Kind(30)", 30, "Kind(30)"},
		{"Kind(31)", 31, "Kind(31)"},
		{"Kind(32)", 32, "Kind(32)"},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			actual := tt.in.String()
			if actual != tt.expected {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
