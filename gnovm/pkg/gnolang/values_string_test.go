package gnolang_test

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestStringValue_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    gnolang.StringValue
		expected string
	}{
		{
			name:     "empty",
			expected: `""`,
		},
		{
			name:     "normal string",
			value:    gnolang.StringValue("hello"),
			expected: `"hello"`,
		},
		{
			name:     "string with quotes",
			value:    gnolang.StringValue(`"hello"`),
			expected: `"\"hello\""`,
		},
		{
			name:     "string with nested quotes",
			value:    gnolang.StringValue(`"hello and \"goodbye\""`),
			expected: `"\"hello and \\\"goodbye\\\"\""`,
		},
		{
			name:     "long string",
			value:    gnolang.StringValue(strings.Repeat("a", 2000)),
			expected: `"` + strings.Repeat("a", 1023) + `..."`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.value.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTypedValue_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    gnolang.TypedValue
		expected string
	}{
		{
			name:     "undefined",
			expected: "(undefined)",
		},
		{
			name:     "nil bool",
			value:    gnolang.TypedValue{T: gnolang.BoolType},
			expected: "(false bool)",
		},
		{
			name: "empty string",
			value: gnolang.TypedValue{
				T: gnolang.StringType,
				V: gnolang.StringValue(""),
			},
			expected: `("" string)`,
		},
		{
			name: "quotes string",
			value: gnolang.TypedValue{
				T: gnolang.StringType,
				V: gnolang.StringValue(`""`),
			},
			expected: `("\"\"" string)`,
		},
		{
			name: "nested quotes string",
			value: gnolang.TypedValue{
				T: gnolang.StringType,
				V: gnolang.StringValue(`"\"\""`),
			},
			expected: `("\"\\\"\\\"\"" string)`,
		},
		{
			name: "string:string map",
			value: gnolang.TypedValue{
				T: &gnolang.MapType{
					Key:   gnolang.StringType,
					Value: gnolang.StringType,
				},
				V: &gnolang.MapValue{
					List: &gnolang.MapList{
						Head: &gnolang.MapListItem{
							Key: gnolang.TypedValue{
								T: gnolang.StringType,
								V: gnolang.StringValue("key"),
							},
							Value: gnolang.TypedValue{
								T: gnolang.StringType,
								V: gnolang.StringValue("value"),
							},
							Next: &gnolang.MapListItem{
								Key: gnolang.TypedValue{
									T: gnolang.StringType,
									V: gnolang.StringValue(`"key"`),
								},
								Value: gnolang.TypedValue{
									T: gnolang.StringType,
									V: gnolang.StringValue(`"value"`),
								},
							},
						},
					},
				},
			},
			expected: `(map{("key" string):("value" string),("\"key\"" string):("\"value\"" string)} map[string]string)`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.value.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
