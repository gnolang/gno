package common

import (
	"testing"

	"gno.land/p/demo/grc/grc721"
)

func TestPositionIdFrom(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		want      grc721.TokenID
		wantPanic bool
	}{
		{
			name:  "string input",
			input: "123",
			want:  grc721.TokenID("123"),
		},
		{
			name:  "int input",
			input: 123,
			want:  grc721.TokenID("123"),
		},
		{
			name:  "uint64 input",
			input: uint64(123),
			want:  grc721.TokenID("123"),
		},
		{
			name:  "grc721.TokenID input",
			input: grc721.TokenID("123"),
			want:  grc721.TokenID("123"),
		},
		{
			name:      "nil input",
			input:     nil,
			wantPanic: true,
		},
		{
			name:      "unsupported type (byte)",
			input:     []byte("123"),
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("PositionIdFrom() should have panicked")
					}
				}()
			}

			got := PositionIdFrom(tt.input)
			if got != tt.want {
				t.Errorf("PositionIdFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
