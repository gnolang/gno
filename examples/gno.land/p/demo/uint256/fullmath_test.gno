package uint256

import (
	"testing"

	"gno.land/p/demo/uassert"
)

var Q128 *Uint

func init() {
	Q128 = MustFromDecimal("340282366920938463463374607431768211456") // 2**128
}

func TestMulDiv(t *testing.T) {
	tests := []struct {
		name        string
		a           *Uint
		b           *Uint
		denominator *Uint
		want        *Uint
		wantErr     bool
	}{
		{
			name:        "simple multiplication and division",
			a:           NewUint(100),
			b:           NewUint(200),
			denominator: NewUint(50),
			want:        NewUint(400),
			wantErr:     false,
		},
		{
			name:        "division by zero",
			a:           NewUint(100),
			b:           NewUint(200),
			denominator: Zero(),
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "zero numerator",
			a:           Zero(),
			b:           NewUint(200),
			denominator: NewUint(50),
			want:        Zero(),
			wantErr:     false,
		},
		{
			name:        "large numbers within bounds",
			a:           Q128,
			b:           NewUint(50),
			denominator: NewUint(100),
			want:        new(Uint).Div(new(Uint).Mul(Q128, NewUint(50)), NewUint(100)),
			wantErr:     false,
		},
		{
			name:        "max uint256 values",
			a:           MustFromDecimal(twoPow256Sub1),
			b:           One(),
			denominator: One(),
			want:        MustFromDecimal(twoPow256Sub1),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MulDiv(tt.a, tt.b, tt.denominator)
			if err != nil {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
				uassert.Equal(t, got.String(), tt.want.String())
			}
		})
	}
}

func TestMulDivRoundingUp(t *testing.T) {
	tests := []struct {
		name        string
		a           *Uint
		b           *Uint
		denominator *Uint
		want        *Uint
		wantErr     bool
	}{
		{
			name:        "exact division",
			a:           NewUint(100),
			b:           NewUint(200),
			denominator: NewUint(50),
			want:        NewUint(400),
			wantErr:     false,
		},
		{
			name:        "division with rounding up",
			a:           NewUint(101),
			b:           NewUint(200),
			denominator: NewUint(50),
			want:        NewUint(404),
			wantErr:     false,
		},
		{
			name:        "division by zero",
			a:           NewUint(100),
			b:           NewUint(200),
			denominator: Zero(),
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "zero numerator",
			a:           Zero(),
			b:           NewUint(200),
			denominator: NewUint(50),
			want:        Zero(),
			wantErr:     false,
		},
		{
			name:        "large numbers with remainder",
			a:           Q128,
			b:           NewUint(51),
			denominator: NewUint(100),
			want:        new(Uint).Add(new(Uint).Div(new(Uint).Mul(Q128, NewUint(51)), NewUint(100)), One()),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MulDivRoundingUp(tt.a, tt.b, tt.denominator)
			if err != nil {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
				uassert.Equal(t, got.String(), tt.want.String())
			}
		})
	}
}

func TestDivRoundingUp(t *testing.T) {
	tests := []struct {
		name    string
		x       *Uint
		y       *Uint
		want    *Uint
		wantErr bool
	}{
		{
			name:    "division by zero",
			x:       NewUint(5),
			y:       Zero(),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "simple division without remainder",
			x:       NewUint(10),
			y:       NewUint(2),
			want:    NewUint(5),
			wantErr: false,
		},
		{
			name:    "division with remainder should round up",
			x:       NewUint(11),
			y:       NewUint(3),
			want:    NewUint(4),
			wantErr: false,
		},
		{
			name:    "division of zero by non-zero",
			x:       Zero(),
			y:       NewUint(5),
			want:    Zero(),
			wantErr: false,
		},
		{
			name:    "division of max uint256 by one",
			x:       MustFromDecimal(twoPow256Sub1),
			y:       One(),
			want:    MustFromDecimal(twoPow256Sub1),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DivRoundingUp(tt.x, tt.y)
			if err != nil {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
				uassert.Equal(t, got.String(), tt.want.String())
			}
		})
	}
}
