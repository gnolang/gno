package rules

import "testing"

func TestIsExported(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"uppercase", "Foo", true},
		{"lowercase", "foo", false},
		{"uppercase with numbers", "Foo123", true},
		{"lowercase with numbers", "foo123", false},
		{"underscore start", "_Foo", false},
		{"single uppercase", "F", true},
		{"single lowercase", "f", false},
		{"empty string", "", false},
		{"unicode uppercase", "Über", true},
		{"unicode lowercase", "über", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExported(tt.input)
			if got != tt.expected {
				t.Errorf("isExported(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsFileLevelDecl(t *testing.T) {
	if isFileLevelDecl(nil) {
		t.Error("isFileLevelDecl(nil) = true, want false")
	}
}
