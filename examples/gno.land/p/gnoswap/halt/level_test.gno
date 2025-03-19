package halt

import (
	"testing"
)

func TestCompositeOpFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CompositeOp
		wantErr bool
	}{
		{
			name:    "should parse 'and' operator correctly",
			input:   "and",
			want:    CompositeOpAnd,
			wantErr: false,
		},
		{
			name:    "should parse 'or' operator correctly",
			input:   "or",
			want:    CompositeOpOr,
			wantErr: false,
		},
		{
			name:    "should return error for invalid operator",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "should return error for empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompositeOpFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompositeOpFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompositeOpFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
