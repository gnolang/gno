package json

import (
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
	tests := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name:  "Simple map",
			input: map[string]interface{}{"a": 1, "b": 2},
			want:  map[string]interface{}{"a": 1, "b": 2},
		},
		{
			name:  "Nested map",
			input: map[string]interface{}{"a": map[string]interface{}{"b": 2, "c": 3}},
			want:  map[string]interface{}{"a.b": 2, "a.c": 3},
		},
		{
			name:  "Empty map",
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Flatten(tt.input)
			if !equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnflatten(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name: "Simple nested map",
			input: map[string]interface{}{
				"a.b": 2,
				"a.c": 3,
			},
			want: map[string]interface{}{
				"a": map[string]interface{}{
					"b": 2,
					"c": 3,
				},
			},
		},
		{
			name: "Multi-level nested map",
			input: map[string]interface{}{
				"a.b.c": 1,
				"a.b.d": 2,
			},
			want: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": 1,
						"d": 2,
					},
				},
			},
		},
		{
			name:  "Empty map",
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unflatten(tt.input)
			if !mapsAreEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapToJSONString(t *testing.T) {
    cases := []struct {
        name     string
        input    map[string]interface{}
        expected string
    }{
        {
            name: "Simple Map",
            input: map[string]interface{}{
                "a": 1,
                "b": 2,
            },
            expected: "{\"a\": 1, \"b\": 2}",
        },
        {
            name: "Nested Map",
            input: map[string]interface{}{
                "a": map[string]interface{}{
                    "b": 2,
                    "c": 3,
                },
                "d": 4,
            },
            expected: "{\"a.b\": 2, \"a.c\": 3, \"d\": 4}",
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            flattened := Flatten(tc.input)
            result := MapToJSONString(flattened)
            if result != tc.expected {
                t.Errorf("Test %s failed. Expected %s, got %s", tc.name, tc.expected, result)
            }
        })
    }
}
