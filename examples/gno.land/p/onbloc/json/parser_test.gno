package json

import "testing"

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
