package ufmt

import (
	"bytes"
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
		// normal usage
		{"hello %s!", []any{"planet"}, "hello planet!"},
		{"hello %v!", []any{"planet"}, "hello planet!"},
		{"hi %%%s!", []any{"worl%d"}, "hi %worl%d!"},
		{"%s %c %d %t", []any{"foo", 'α', 421, true}, "foo α 421 true"},
		{"string [%s]", []any{"foo"}, "string [foo]"},

		// ints
		{"int [%d]", []any{int(42)}, "int [42]"},
		{"int [%03d]", []any{int(42)}, "int [042]"},
		{"int [%010d]", []any{int(42)}, "int [0000000042]"},
		{"int [% 10d]", []any{int(42)}, "int [        42]"},
		{"hex [%x]", []any{int(42)}, "hex [2a]"},
		{"hex [%X]", []any{int(42)}, "hex [2A]"},
		{"hex [%02x]", []any{int(42)}, "hex [2a]"},
		{"hex [%03x]", []any{int(42)}, "hex [02a]"},
		{"hex [%#03x]", []any{int(42)}, "hex [0x02a]"},
		{"hex [%#03X]", []any{int(42)}, "hex [0X02A]"},
		{"hex [%#X]", []any{[]uint8{1, 2, 42}}, "hex [0X01022A]"},
		{"hex [%x]", []any{[]uint8("hello")}, "hex [68656c6c6f]"},
		{"hex [%x]", []any{[]uint8{0, 0}}, "hex [0000]"},
		{"hex [%x]", []any{uint8(0)}, "hex [0]"},
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

		// floats
		{"float64 [%e]", []any{float64(64.1)}, "float64 [6.410000e+01]"},
		{"float64 [%E]", []any{float64(64.1)}, "float64 [6.410000E+01]"},
		{"float64 [%f]", []any{float64(64.1)}, "float64 [64.100000]"},
		{"float64 [%.2f]", []any{float64(64.1)}, "float64 [64.10]"},
		{"float32 [%.1f]", []any{float32(64.1)}, "float32 [64.1]"},
		{"float64 [%.f]", []any{float64(64.1)}, "float64 [64]"},
		{"float64 [%.0f]", []any{float64(64.1)}, "float64 [64]"},
		{"float64 [%.1f]", []any{float64(64.1)}, "float64 [64.1]"},
		{"float64 [%F]", []any{float64(64.1)}, "float64 [64.100000]"},
		{"float64 [%g]", []any{float64(64.1)}, "float64 [64.1]"},
		{"float64 [%G]", []any{float64(64.1)}, "float64 [64.1]"},
		{"float64 [%.000001f]", []any{float64(64.1)}, "float64 [64.1]"},

		// bool
		{"bool [%t]", []any{true}, "bool [true]"},
		{"bool [%v]", []any{true}, "bool [true]"},
		{"bool [%t]", []any{false}, "bool [false]"},
		{"bool [%v]", []any{false}, "bool [false]"},

		// errros/special cases
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

		{"%.2f", []any{3.14159}, "3.14"},
		{"%.4f", []any{3.14159}, "3.1416"},
		{"%.0f", []any{3.14159}, "3"},
		{"%.1f", []any{3.0}, "3.0"},
		{"%.3F", []any{3.14159}, "3.142"},
		{"%.2e", []any{314.159}, "3.14e+02"},
		{"%.3E", []any{314.159}, "3.142E+02"},
		{"%.3g", []any{3.14159}, "3.14"},
		{"%.5G", []any{3.14159}, "3.1416"},
		{"%.0f", []any{3.6}, "4"},
		{"%.0f", []any{3.4}, "3"},
		{"%.1f", []any{0.0}, "0.0"},
		{"%.2f", []any{1e6}, "1000000.00"},
		{"%.2f", []any{1e-6}, "0.00"},

		{"%5s", []any{"Hello World"}, "Hello World"},
		{"%3s", []any{"Hi"}, " Hi"},
		{"%2s", []any{"Hello"}, "Hello"},
		{"%1s", []any{"A"}, "A"},
		{"%0s", []any{"Test"}, "Test"},
		{"%5s!", []any{"Hello World"}, "Hello World!"},
		{"_%5s_", []any{"abc"}, "_  abc_"},
		{"%2s%4s", []any{"ab", "cde"}, "ab cde"},
		{"%5s", []any{""}, "     "},
		{"%3s", []any{nil}, "%!s(<nil>)"},
		{"%2s", []any{123}, "%!s(int=123)"},

		// entering dangerous printf zone

		// advanced int (from go/fmt)
		{"%.d", []any{0}, ""},
		{"%.0d", []any{0}, ""},
		{"% d", []any{12345}, " 12345"},
		{"%+d", []any{12345}, "+12345"},
		{"%+d", []any{-12345}, "-12345"},
		{"%#X", []any{0}, "0X0"},
		{"%x", []any{0x12abcdef}, "12abcdef"},
		{"%X", []any{0x12abcdef}, "12ABCDEF"},
		{"%x", []any{^uint32(0)}, "ffffffff"},
		{"%X", []any{^uint64(0)}, "FFFFFFFFFFFFFFFF"},
		// {"%010d", []any{12345}, "0000012345"},
		// {"%010d", []any{-12345}, "-000012345"}, // weird behaviors

		// advanced float (from go/fmt)
		{"========= ADVANCED FLOAT", []any{}, "========= ADVANCED FLOAT"},
		{"%+.3e", []any{0.0}, "+0.000e+00"},
		{"%+.3e", []any{1.0}, "+1.000e+00"},
		{"%+.3x", []any{0.0}, "+0x0.000p+00"}, // [x]
		{"%+.3x", []any{1.0}, "+0x1.000p+00"}, // [x]
		{"%+.3f", []any{-1.0}, "-1.000"}, // [f]
		{"%+.3F", []any{-1.0}, "-1.000"}, // [F]
		{"%+.3F", []any{float32(-1.0)}, "-1.000"}, // [F]
		{"% .3E", []any{-1.0}, "-1.000E+00"}, // [E]
		{"% .3e", []any{1.0}, " 1.000e+00"}, // [e]
		{"% .3X", []any{-1.0}, "-0X1.000P+00"}, // [X]
		{"% .3x", []any{1.0}, " 0x1.000p+00"}, // [x]
		{"%+.3g", []any{0.0}, "+0"}, // [g]
		{"%+.3g", []any{1.0}, "+1"}, // [g]
		{"%+.3g", []any{-1.0}, "-1"}, // [g]
		{"% .3g", []any{-1.0}, "-1"}, // [g]
		{"% .3g", []any{1.0}, " 1"}, // [g]

		// Test sharp flag used with floats.
		{"%#g", []any{1e-323}, "1.00000e-323"}, // [g]
		{"%#g", []any{-1.0}, "-1.00000"}, // [g]
		{"%#g", []any{1.1}, "1.10000"}, // [g]
		{"%#g", []any{123456.0}, "123456."}, // [g]
		{"%#g", []any{1234567.0}, "1.234567e+06"}, // [g]
		{"%#g", []any{1230000.0}, "1.23000e+06"}, // [g]
		{"%#g", []any{1000000.0}, "1.00000e+06"}, // [g]
		{"%#.0f", []any{1.0}, "1."}, // [f]
		{"%#.0e", []any{1.0}, "1.e+00"}, // [e]
		{"%#.0x", []any{1.0}, "0x1.p+00"}, // [x]
		{"%#.0g", []any{1.0}, "1."}, // [g]
		{"%#.0g", []any{1100000.0}, "1.e+06"}, // [g]
		{"%#.4f", []any{1.0}, "1.0000"}, // [f]
		{"%#.4e", []any{1.0}, "1.0000e+00"}, // [e]
		{"%#.4x", []any{1.0}, "0x1.0000p+00"}, // [x]
		{"%#.4g", []any{1.0}, "1.000"}, // [g]
		{"%#.4g", []any{100000.0}, "1.000e+05"}, // [g]
		{"%#.4g", []any{1.234}, "1.234"}, // [g]
		{"%#.4g", []any{0.1234}, "0.1234"}, // [g]
		{"%#.4g", []any{1.23}, "1.230"}, // [g]
		{"%#.4g", []any{0.123}, "0.1230"}, // [g]
		{"%#.4g", []any{1.2}, "1.200"}, // [g]
		{"%#.4g", []any{0.12}, "0.1200"}, // [g]
		{"%#.4g", []any{10.2}, "10.20"}, // [g]
		{"%#.4g", []any{0.0}, "0.000"}, // [g]
		{"%#.4g", []any{0.012}, "0.01200"}, // [g]
		{"%#.0f", []any{123.0}, "123."}, // [f]
		{"%#.0e", []any{123.0}, "1.e+02"}, // [e]
		{"%#.0x", []any{123.0}, "0x1.p+07"}, // [x]
		{"%#.0g", []any{123.0}, "1.e+02"}, // [g]
		{"%#.4f", []any{123.0}, "123.0000"}, // [f]
		{"%#.4e", []any{123.0}, "1.2300e+02"}, // [e]
		{"%#.4x", []any{123.0}, "0x1.ec00p+06"}, // [x]
		{"%#.4g", []any{123.0}, "123.0"}, // [g]
		{"%#.4g", []any{123000.0}, "1.230e+05"}, // [g]
	}

	nFails := 0
	for _, tc := range cases {
		name := fmt.Sprintf(tc.format, tc.values...)
		t.Run(name, func(t *testing.T) {
			got := Sprintf(tc.format, tc.values...)
			if got != tc.expectedOutput {
				nFails++
				t.Errorf("got %q, want %q.", got, tc.expectedOutput)
			}
		})
	}
	if nFails != 0 {
		t.Logf("=> (%d tests failed)", nFails)
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

func TestSprint(t *testing.T) {
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
			expected: "Hello 3.140000 (unhandled)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Sprint(tc.args...)
			if got != tc.expected {
				t.Errorf("got %q, want %q.", got, tc.expected)
			}
		})
	}
}

func TestFprintf(t *testing.T) {
	var buf bytes.Buffer
	n, err := Fprintf(&buf, "Count: %d, Message: %s", 42, "hello")
	if err != nil {
		t.Fatalf("Fprintf failed: %v", err)
	}

	const expected = "Count: 42, Message: hello"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes written, got %d", len(expected), n)
	}
}

// TODO: replace os.Stdout with a buffer to capture the output and test it.
func TestPrintf(t *testing.T) {
	n, err := Printf("The answer is %d", 42)
	if err != nil {
		t.Fatalf("Printf failed: %v", err)
	}

	const expected = "The answer is 42"
	if n != len(expected) {
		t.Errorf("Expected 14 bytes written, got %d", n)
	}
}

func TestAppendf(t *testing.T) {
	b := []byte("Header: ")
	result := Appendf(b, "Value %d", 7)
	const expected = "Header: Value 7"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFprint(t *testing.T) {
	var buf bytes.Buffer
	n, err := Fprint(&buf, "Hello", 42, true)
	if err != nil {
		t.Fatalf("Fprint failed: %v", err)
	}

	const expected = "Hello 42 true"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes written, got %d", len(expected), n)
	}
}

// TODO: replace os.Stdout with a buffer to capture the output and test it.
func TestPrint(t *testing.T) {
	n, err := Print("Mixed", 3.14, false)
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	const expected = "Mixed 3.140000 false"
	if n != len(expected) {
		t.Errorf("Expected 12 bytes written, got %d", n)
	}
}

func TestAppend(t *testing.T) {
	b := []byte{0x01, 0x02}
	result := Append(b, "Test", 99)

	const expected = "\x01\x02Test 99"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFprintln(t *testing.T) {
	var buf bytes.Buffer
	n, err := Fprintln(&buf, "Line", 1)
	if err != nil {
		t.Fatalf("Fprintln failed: %v", err)
	}

	const expected = "Line 1\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes written, got %d", len(expected), n)
	}
}

// TODO: replace os.Stdout with a buffer to capture the output and test it.
func TestPrintln(t *testing.T) {
	n, err := Println("Output", "test")
	if err != nil {
		t.Fatalf("Println failed: %v", err)
	}

	const expected = "Output test\n"
	if n != len(expected) {
		t.Errorf("Expected 12 bytes written, got %d", n)
	}
}

func TestSprintln(t *testing.T) {
	result := Sprintln("Item", 42)

	const expected = "Item 42\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestAppendln(t *testing.T) {
	b := []byte("Start:")
	result := Appendln(b, "End")

	const expected = "Start:End\n"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
