package json

import (
	"strconv"
	"testing"
)

func TestParseStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		isError  bool
	}{
		{`"Hello, World!"`, "\"Hello, World!\"", false},
		{`\uFF11`, "\uFF11", false},
		{`\uFFFF`, "\uFFFF", false},
		{`true`, "true", false},
		{`false`, "false", false},
		{`\uDF00`, "", true},
	}

	for i, tt := range tests {
		s, err := ParseStringLiteral([]byte(tt.input))

		if !tt.isError && err != nil {
			t.Errorf("%d. unexpected error: %s", i, err)
		}

		if tt.isError && err == nil {
			t.Errorf("%d. expected error, but not error", i)
		}

		if s != tt.expected {
			t.Errorf("%d. expected=%s, but actual=%s", i, tt.expected, s)
		}
	}
}

func TestParseBoolLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		isError  bool
	}{
		{`true`, true, false},
		{`false`, false, false},
		{`TRUE`, false, true},
		{`FALSE`, false, true},
		{`foo`, false, true},
		{`"true"`, false, true},
		{`"false"`, false, true},
	}

	for i, tt := range tests {
		b, err := ParseBoolLiteral([]byte(tt.input))

		if !tt.isError && err != nil {
			t.Errorf("%d. unexpected error: %s", i, err)
		}

		if tt.isError && err == nil {
			t.Errorf("%d. expected error, but not error", i)
		}

		if b != tt.expected {
			t.Errorf("%d. expected=%t, but actual=%t", i, tt.expected, b)
		}
	}
}

func TestParseFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"123", 123},
		{"-123", -123},
		{"123.456", 123.456},
		{"-123.456", -123.456},
		{"12345678.1234567890", 12345678.1234567890},
		{"-12345678.09123456789", -12345678.09123456789},
		{"0.123", 0.123},
		{"-0.123", -0.123},
		{"", -1},
		{"abc", -1},
		{"123.45.6", -1},
		{"999999999999999999999", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := ParseFloatLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseFloatLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseFloatWithScientificNotation(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1e6", 1000000},
		{"1E6", 1000000},
		{"1.23e10", 1.23e10},
		{"1.23E10", 1.23e10},
		{"-1.23e10", -1.23e10},
		{"-1.23E10", -1.23e10},
		{"2.45e-8", 2.45e-8},
		{"2.45E-8", 2.45e-8},
		{"-2.45e-8", -2.45e-8},
		{"-2.45E-8", -2.45e-8},
		{"5e0", 5},
		{"-5e0", -5},
		{"5E+0", 5},
		{"5e+1", 50},
		{"1e-1", 0.1},
		{"1E-1", 0.1},
		{"-1e-1", -0.1},
		{"-1E-1", -0.1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFloatLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseFloatLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}

			if err != nil {
				t.Errorf("ParseFloatLiteral(%s): got error %v", tt.input, err)
			}
		})
	}
}

func TestParseFloat_May_Interoperability_Problem(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
	}{
		{"3.141592653589793238462643383279", true},
		{"1E400", false}, // TODO: should error
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseFloatLiteral([]byte(tt.input))
			if tt.shouldErr && err == nil {
				t.Errorf("ParseFloatLiteral(%s): expected error, but not error", tt.input)
			}
		})
	}
}

func TestParseIntLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"0", 0},
		{"1", 1},
		{"-1", -1},
		{"12345", 12345},
		{"-12345", -12345},
		{"9223372036854775807", 9223372036854775807},
		{"-9223372036854775808", -9223372036854775808},
		{"-92233720368547758081", 0},
		{"18446744073709551616", 0},
		{"9223372036854775808", 0},
		{"-9223372036854775809", 0},
		{"", 0},
		{"abc", 0},
		{"12345x", 0},
		{"123e5", 0},
		{"9223372036854775807x", 0},
		{"27670116110564327410", 0},
		{"-27670116110564327410", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := ParseIntLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseIntLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TODO: remove comment after benchmark added.
// func benchmarkParseStringLiteral(b *testing.B, input []byte) {
// 	for i := 0; i < b.N; i++ {
// 		_, err := ParseStringLiteral(input)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// }

// func BenchmarkParseStringLiteral_Simple(b *testing.B) {
// 	benchmarkParseStringLiteral(b, []byte(`"Hello, World!"`))
// }

// func BenchmarkParseStringLiteral_Long(b *testing.B) {
// 	benchmarkParseStringLiteral(b, []byte(`"Lorem ipsum dolor sit amet, consectetur adipiscing elit."`))
// }

// func BenchmarkParseStringLiteral_Escaped(b *testing.B) {
// 	benchmarkParseStringLiteral(b, []byte(`"Hello, \"World!\""`))
// }

// func BenchmarkParseStringLiteral_Unicode(b *testing.B) {
// 	benchmarkParseStringLiteral(b, []byte(`"안녕, 世界!"`))
// }

// func benchmarkParseFloatLiteral(b *testing.B, input []byte) {
// 	for i := 0; i < b.N; i++ {
// 		ParseFloatLiteral(input)
// 	}
// }

// func benchmarkStrconvParseFloat(b *testing.B, input string) {
// 	for i := 0; i < b.N; i++ {
// 		_, err := strconv.ParseFloat(input, 64)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// }

// func BenchmarkParseFloatLiteral_Simple(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("123.456"))
// }

// func BenchmarkStrconvParseFloat_Simple(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "123.456")
// }

// func BenchmarkParseFloatLiteral_Long(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("123456.78901"))
// }

// func BenchmarkStrconvParseFloat_Long(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "123456.78901")
// }

// func BenchmarkParseFloatLiteral_Negative(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("-123.456"))
// }

// func BenchmarkStrconvParseFloat_Negative(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "-123.456")
// }

// func BenchmarkParseFloatLiteral_Negative_Long(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("-123456.78901"))
// }

// func BenchmarkStrconvParseFloat_Negative_Long(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "-123456.78901")
// }

// func BenchmarkParseFloatLiteral_Decimal(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("123456"))
// }

// func BenchmarkStrconvParseFloat_Decimal(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "123456")
// }

// func BenchmarkParseFloatLiteral_Float_No_Decimal(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("123."))
// }

// func BenchmarkStrconvParseFloat_Float_No_Decimal(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "123.")
// }

// func BenchmarkParseFloatLiteral_Complex(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("-123456.798012345"))
// }

// func BenchmarkStrconvParseFloat_Complex(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "-123456.798012345")
// }

// func BenchmarkParseFloatLiteral_Science_Notation(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("1e6"))
// }

// func BenchmarkStrconvParseFloat_Science_Notation(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "1e6")
// }

// func BenchmarkParseFloatLiteral_Science_Notation_Long(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("1.23e10"))
// }

// func BenchmarkStrconvParseFloat_Science_Notation_Long(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "1.23e10")
// }

// func BenchmarkParseFloatLiteral_Science_Notation_Negative(b *testing.B) {
// 	benchmarkParseFloatLiteral(b, []byte("-1.23e10"))
// }

// func BenchmarkStrconvParseFloat_Science_Notation_Negative(b *testing.B) {
// 	benchmarkStrconvParseFloat(b, "-1.23e10")
// }
