package ufmt

import (
	"errors"
	"fmt"
	"testing"
)

type stringer struct{}

func (stringer) String() string {
	return "I'm a stringer"
}

func TestSprintf(t *testing.T) {
	tru := true
	cases := []struct {
		format         string
		values         []any
		expectedOutput string
	}{
		{"hello %s!", []any{"planet"}, "hello planet!"},
		{"hello %v!", []any{"planet"}, "hello planet!"},
		{"hi %%%s!", []any{"worl%d"}, "hi %worl%d!"},
		{"%s %c %d %t", []any{"foo", 'α', 421, true}, "foo α 421 true"},
		{"string [%s]", []any{"foo"}, "string [foo]"},
		{"int [%d]", []any{int(42)}, "int [42]"},
		{"int [%v]", []any{int(42)}, "int [42]"},
		{"int8 [%d]", []any{int8(8)}, "int8 [8]"},
		{"int8 [%v]", []any{int8(8)}, "int8 [8]"},
		{"int16 [%d]", []any{int16(16)}, "int16 [16]"},
		{"int16 [%v]", []any{int16(16)}, "int16 [16]"},
		{"int32 [%d]", []any{int32(32)}, "int32 [32]"},
		{"int32 [%v]", []any{int32(32)}, "int32 [32]"},
		{"int64 [%d]", []any{int64(64)}, "int64 [64]"},
		{"int64 [%v]", []any{int64(64)}, "int64 [64]"},
		{"uint [%d]", []any{uint(42)}, "uint [42]"},
		{"uint [%v]", []any{uint(42)}, "uint [42]"},
		{"uint8 [%d]", []any{uint8(8)}, "uint8 [8]"},
		{"uint8 [%v]", []any{uint8(8)}, "uint8 [8]"},
		{"uint16 [%d]", []any{uint16(16)}, "uint16 [16]"},
		{"uint16 [%v]", []any{uint16(16)}, "uint16 [16]"},
		{"uint32 [%d]", []any{uint32(32)}, "uint32 [32]"},
		{"uint32 [%v]", []any{uint32(32)}, "uint32 [32]"},
		{"uint64 [%d]", []any{uint64(64)}, "uint64 [64]"},
		{"uint64 [%v]", []any{uint64(64)}, "uint64 [64]"},
		{"float64 [%e]", []any{float64(64.1)}, "float64 [6.41e+01]"},
		{"float64 [%E]", []any{float64(64.1)}, "float64 [6.41E+01]"},
		{"float64 [%f]", []any{float64(64.1)}, "float64 [64.100000]"},
		{"float64 [%F]", []any{float64(64.1)}, "float64 [64.100000]"},
		{"float64 [%g]", []any{float64(64.1)}, "float64 [64.1]"},
		{"float64 [%G]", []any{float64(64.1)}, "float64 [64.1]"},
		{"bool [%t]", []any{true}, "bool [true]"},
		{"bool [%v]", []any{true}, "bool [true]"},
		{"bool [%t]", []any{false}, "bool [false]"},
		{"bool [%v]", []any{false}, "bool [false]"},
		{"no args", nil, "no args"},
		{"finish with %", nil, "finish with %"},
		{"stringer [%s]", []any{stringer{}}, "stringer [I'm a stringer]"},
		{"â", nil, "â"},
		{"Hello, World! 😊", nil, "Hello, World! 😊"},
		{"unicode formatting: %s", []any{"😊"}, "unicode formatting: 😊"},
		{"invalid hex [%x]", []any{"invalid"}, "invalid hex [(unhandled)]"},
		{"rune as character [%c]", []any{rune('A')}, "rune as character [A]"},
		{"int as character [%c]", []any{int('B')}, "int as character [B]"},
		{"quoted string [%q]", []any{"hello"}, "quoted string [\"hello\"]"},
		{"quoted string with escape [%q]", []any{"\thello\nworld\\"}, "quoted string with escape [\"\\thello\\nworld\\\\\"]"},
		{"invalid quoted string [%q]", []any{123}, "invalid quoted string [(unhandled)]"},
		{"type of bool [%T]", []any{true}, "type of bool [bool]"},
		{"type of int [%T]", []any{123}, "type of int [int]"},
		{"type of string [%T]", []any{"hello"}, "type of string [string]"},
		{"type of []byte [%T]", []any{[]byte{1, 2, 3}}, "type of []byte [[]byte]"},
		{"type of []rune [%T]", []any{[]rune{'a', 'b', 'c'}}, "type of []rune [[]rune]"},
		{"type of unknown [%T]", []any{struct{}{}}, "type of unknown [unknown]"},
		// mismatch printing
		{"%s", []any{nil}, "%!s(<nil>)"},
		{"%s", []any{421}, "%!s(int=421)"},
		{"%s", []any{"z"}, "z"},
		{"%s", []any{tru}, "%!s(bool=true)"},
		{"%s", []any{'z'}, "%!s(int32=122)"},

		{"%c", []any{nil}, "%!c(<nil>)"},
		{"%c", []any{421}, "ƥ"},
		{"%c", []any{"z"}, "%!c(string=z)"},
		{"%c", []any{tru}, "%!c(bool=true)"},
		{"%c", []any{'z'}, "z"},

		{"%d", []any{nil}, "%!d(<nil>)"},
		{"%d", []any{421}, "421"},
		{"%d", []any{"z"}, "%!d(string=z)"},
		{"%d", []any{tru}, "%!d(bool=true)"},
		{"%d", []any{'z'}, "122"},

		{"%t", []any{nil}, "%!t(<nil>)"},
		{"%t", []any{421}, "%!t(int=421)"},
		{"%t", []any{"z"}, "%!t(string=z)"},
		{"%t", []any{tru}, "true"},
		{"%t", []any{'z'}, "%!t(int32=122)"},
	}

	for _, tc := range cases {
		name := fmt.Sprintf(tc.format, tc.values...)
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
		args     []any
		expected string
	}{
		{
			name:     "simple string",
			format:   "error: %s",
			args:     []any{"something went wrong"},
			expected: "error: something went wrong",
		},
		{
			name:     "integer value",
			format:   "value: %d",
			args:     []any{42},
			expected: "value: 42",
		},
		{
			name:     "boolean value",
			format:   "success: %t",
			args:     []any{true},
			expected: "success: true",
		},
		{
			name:     "multiple values",
			format:   "error %d: %s (success=%t)",
			args:     []any{123, "failure occurred", false},
			expected: "error 123: failure occurred (success=false)",
		},
		{
			name:     "literal percent",
			format:   "literal %%",
			args:     []any{},
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

func TestPrintErrors(t *testing.T) {
	got := Sprintf("error: %s", errors.New("can I be printed?"))
	expectedOutput := "error: can I be printed?"
	if got != expectedOutput {
		t.Errorf("got %q, want %q.", got, expectedOutput)
	}
}

// NOTE: Currently, there is no way to get the output of Println without using os.Stdout,
// so we can only test that it doesn't panic and print arguments well.
func TestPrintln(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "Empty args",
			args:     []any{},
			expected: "",
		},
		{
			name:     "String args",
			args:     []any{"Hello", "World"},
			expected: "Hello World",
		},
		{
			name:     "Integer args",
			args:     []any{1, 2, 3},
			expected: "1 2 3",
		},
		{
			name:     "Mixed args",
			args:     []any{"Hello", 42, true, false, "World"},
			expected: "Hello 42 true false World",
		},
		{
			name:     "Unhandled type",
			args:     []any{"Hello", 3.14, []int{1, 2, 3}},
			expected: "Hello (unhandled) (unhandled)",
		},
	}

	// TODO: replace os.Stdout with a buffer to capture the output and test it.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Println(tt.args...)
		})
	}
}
