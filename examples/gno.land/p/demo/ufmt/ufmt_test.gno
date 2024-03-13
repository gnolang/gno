package ufmt

import (
	"bytes"
	"testing"
)

type stringer struct{}

func (stringer) String() string {
	return "I'm a stringer"
}

func TestSprintf(t *testing.T) {
	cases := []struct {
		format         string
		values         []interface{}
		expectedOutput string
	}{
		{"hello %s!", []interface{}{"planet"}, "hello planet!"},
		{"hi %%%s!", []interface{}{"worl%d"}, "hi %worl%d!"},
		{"string [%s]", []interface{}{"foo"}, "string [foo]"},
		{"int [%d]", []interface{}{int(42)}, "int [42]"},
		{"uint [%d]", []interface{}{uint(42)}, "uint [42]"},
		{"int64 [%d]", []interface{}{int64(42)}, "int64 [42]"},
		{"uint64 [%d]", []interface{}{uint64(42)}, "uint64 [42]"},
		{"bool [%t]", []interface{}{true}, "bool [true]"},
		{"bool [%t]", []interface{}{false}, "bool [false]"},
		{"invalid bool [%t]", []interface{}{"invalid"}, "invalid bool [(unhandled)]"},
		{"invalid integer [%d]", []interface{}{"invalid"}, "invalid integer [(unhandled)]"},
		{"invalid string [%s]", []interface{}{1}, "invalid string [(unhandled)]"},
		{"no args", nil, "no args"},
		{"finish with %", nil, "finish with %"},
		{"stringer [%s]", []interface{}{stringer{}}, "stringer [I'm a stringer]"},
	}

	for _, tc := range cases {
		name := Sprintf(tc.format, tc.values...)
		t.Run(name, func(t *testing.T) {
			got := Sprintf(tc.format, tc.values...)
			if got != tc.expectedOutput {
				t.Errorf("got %q, want %q.", got, tc.expectedOutput)
			}
		})
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple string",
			format:   "error: %s",
			args:     []interface{}{"something went wrong"},
			expected: "error: something went wrong",
		},
		{
			name:     "integer value",
			format:   "value: %d",
			args:     []interface{}{42},
			expected: "value: 42",
		},
		{
			name:     "boolean value",
			format:   "success: %t",
			args:     []interface{}{true},
			expected: "success: true",
		},
		{
			name:     "multiple values",
			format:   "error %d: %s (success=%t)",
			args:     []interface{}{123, "failure occurred", false},
			expected: "error 123: failure occurred (success=false)",
		},
		{
			name:     "literal percent",
			format:   "literal %%",
			args:     []interface{}{},
			expected: "literal %",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Errorf(tt.format, tt.args...)
			if err.Error() != tt.expected {
				t.Errorf("Errorf(%q, %v) = %q, expected %q", tt.format, tt.args, err.Error(), tt.expected)
			}
		})
	}
}
