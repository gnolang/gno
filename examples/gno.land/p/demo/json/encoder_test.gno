package json

import (
	"errors"
	"strings"
	"testing"
)

func equal(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

func mapsAreEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for ka, va := range a {
		vb, ok := b[ka]
		if !ok {
			return false
		}

		switch vaTyped := va.(type) {
		case map[string]interface{}:
			if vbTyped, ok := vb.(map[string]interface{}); ok {
				if !mapsAreEqual(vaTyped, vbTyped) {
					return false
				}
			} else {
				return false
			}
		default:
			if va != vb {
				return false
			}
		}
	}

	return true
}

func TestFlatten(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name: "Basic Test",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John Doe",
					"age":  30,
					"address": map[string]interface{}{
						"street": "123 Main St",
						"city":   "Anytown",
					},
					"favoriteNumbers": []interface{}{7, 13, 42},
				},
			},
			expected: map[string]string{
				"user.name":               "string:\"John Doe\"",
				"user.age":                "number:30",
				"user.address.street":     "string:\"123 Main St\"",
				"user.address.city":       "string:\"Anytown\"",
				"user.favoriteNumbers[0]": "number:7",
				"user.favoriteNumbers[1]": "number:13",
				"user.favoriteNumbers[2]": "number:42",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Flatten(tc.input)
			if !mapsAreEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
