package router

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestTrySwapTypeFromStr(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    SwapType
		wantErr bool
	}{
		{
			name:    "valid EXACT_IN",
			input:   "EXACT_IN",
			want:    ExactIn,
			wantErr: false,
		},
		{
			name:    "valid EXACT_OUT",
			input:   "EXACT_OUT",
			want:    ExactOut,
			wantErr: false,
		},
		{
			name:    "invalid empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid swap type",
			input:   "INVALID_TYPE",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := trySwapTypeFromStr(tt.input)

			if !tt.wantErr {
				uassert.NoError(t, err)
			}

			if got != tt.want {
				t.Errorf("trySwapTypeFromStr() = %v, want %v", got, tt.want)
			}
		})
	}
}