package reporters

import (
	"bytes"
	"testing"
)

func TestNewReporter(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name     string
		format   string
		wantType string
		wantErr  bool
	}{
		{"text format", "text", "*reporters.TextReporter", false},
		{"json format", "json", "*reporters.JSONReporter", false},
		{"empty format defaults to text", "", "*reporters.TextReporter", false},
		{"unknown format", "xml", "", true},
		{"unknown format sarif", "sarif", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewReporter(tt.format, &buf)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if r == nil {
				t.Fatal("reporter should not be nil")
			}

			switch tt.wantType {
			case "*reporters.TextReporter":
				if _, ok := r.(*TextReporter); !ok {
					t.Errorf("expected TextReporter, got %T", r)
				}
			case "*reporters.JSONReporter":
				if _, ok := r.(*JSONReporter); !ok {
					t.Errorf("expected JSONReporter, got %T", r)
				}
			}
		})
	}
}
