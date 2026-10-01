package client

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestParseExpiresAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	maxSecs := int64(std.MaxSessionDuration)

	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "empty errors",
			input:   "",
			wantErr: true,
		},
		{
			name:  "none returns 0",
			input: "none",
			want:  0,
		},
		{
			name:  "24 hours",
			input: "24h",
			want:  now.Unix() + 24*3600,
		},
		{
			name:  "30 minutes",
			input: "30m",
			want:  now.Unix() + 30*60,
		},
		{
			name:  "compound duration",
			input: "1h30m",
			want:  now.Unix() + 90*60,
		},
		{
			name:  "7 days",
			input: "7d",
			want:  now.Unix() + 7*86400,
		},
		{
			name:  "4 weeks",
			input: "4w",
			want:  now.Unix() + 4*7*86400,
		},
		{
			name:  "fractional days",
			input: "0.5d",
			want:  now.Unix() + 12*3600,
		},
		{
			name:  "fractional weeks",
			input: "1.5w",
			want:  now.Unix() + (7*86400 + 3*86400 + 12*3600),
		},
		{
			name:  "future unix timestamp within cap",
			input: strconv.FormatInt(now.Unix()+24*3600, 10),
			want:  now.Unix() + 24*3600,
		},
		{
			name:    "duration way over cap",
			input:   "300w",
			wantErr: true,
		},
		{
			name:    "duration just over cap",
			input:   "1461d",
			wantErr: true,
		},
		{
			name:  "duration at exact cap",
			input: "1460d",
			want:  now.Unix() + maxSecs,
		},
		{
			name:  "30d under cap accepted",
			input: "30d",
			want:  now.Unix() + 30*86400,
		},
		{
			name:    "garbage",
			input:   "tomorrow",
			wantErr: true,
		},
		{
			name:    "negative duration",
			input:   "-1h",
			wantErr: true,
		},
		{
			name:    "zero duration",
			input:   "0h",
			wantErr: true,
		},
		{
			name:    "bare integer rejected as ambiguous",
			input:   "7",
			wantErr: true,
		},
		{
			name:    "bare past timestamp rejected",
			input:   "100",
			wantErr: true,
		},
		{
			name:    "bare future-but-too-far timestamp rejected",
			input:   strconv.FormatInt(now.Unix()+maxSecs+1, 10),
			wantErr: true,
		},
		{
			name:    "overflow w clamped and rejected",
			input:   "1e20w",
			wantErr: true,
		},
		{
			name:    "negative bare integer rejected",
			input:   "-1",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseExpiresAt(tc.input, now)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil (got=%d)", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseExpiresAt(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseDurationSecondsOverflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  int64
	}{
		// Values that overflow int64 seconds when multiplied by the unit.
		{"1e20w", math.MaxInt64},
		{"1e308d", math.MaxInt64},
	}
	for _, tc := range tests {
		got, ok := parseDurationSeconds(tc.input)
		if !ok {
			t.Errorf("parseDurationSeconds(%q) returned ok=false", tc.input)
			continue
		}
		if got != tc.want {
			t.Errorf("parseDurationSeconds(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
