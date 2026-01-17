package lint

import (
	"encoding/json"
	"testing"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"info", SeverityInfo, "info"},
		{"warning", SeverityWarning, "warning"},
		{"error", SeverityError, "error"},
		{"unknown", Severity(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverity_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"info", SeverityInfo, `"info"`},
		{"warning", SeverityWarning, `"warning"`},
		{"error", SeverityError, `"error"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.severity)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSeverity_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Severity
		wantErr bool
	}{
		{"info", `"info"`, SeverityInfo, false},
		{"warning", `"warning"`, SeverityWarning, false},
		{"error", `"error"`, SeverityError, false},
		{"unknown defaults to info", `"unknown"`, SeverityInfo, false},
		{"invalid json", `not json`, SeverityInfo, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Severity
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
