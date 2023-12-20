package json

import (
	"errors"
	"testing"
)

func TestBlockEnd(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		open     byte
		close    byte
		expected int
	}{
		{"Empty Object", []byte("{}"), '{', '}', 2},
		{"Nested Object", []byte(`{"key": {"nestedKey": "value"}}`), '{', '}', 31},
		{"Array", []byte(`["item1", "item2"]`), '[', ']', 18},
		{"Complex Object", []byte(`{"key": [1, 2, 3], "anotherKey": "value"}`), '{', '}', 41},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := blockEnd(test.data, test.open, test.close)
			if result != test.expected {
				t.Errorf("Failed %s: expected %d, got %d", test.name, test.expected, result)
			}
		})
	}
}
