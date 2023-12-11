package json

import (
	"testing"
)

type valueTypeTest struct {
	valueType ValueType
	expected  string
}

func TestValueTypeToString(t *testing.T) {
	tests := []valueTypeTest{
		{NotExist, "not-exist"},
		{String, "string"},
		{Number, "number"},
		{Object, "object"},
		{Array, "array"},
		{Boolean, "boolean"},
		{Null, "null"},
		{Unknown, "unknown"},
	}

	for _, test := range tests {
		result := test.valueType.ToString()
		if result != test.expected {
			t.Errorf("Expected %s, but got %s", test.expected, result)
		}
	}
}
