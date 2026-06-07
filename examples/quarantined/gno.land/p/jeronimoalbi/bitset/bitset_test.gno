package bitset_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/bitset"
)

func TestBitSetSet(t *testing.T) {
	cases := []struct {
		name string
		bits []uint64
		want string
	}{
		{
			name: "single bit 0",
			bits: []uint64{0},
			want: "1",
		},
		{
			name: "single bit 5",
			bits: []uint64{5},
			want: "100000",
		},
		{
			name: "multiple bits in same word",
			bits: []uint64{0, 3, 7},
			want: "10001001",
		},
		{
			name: "bit at word boundary",
			bits: []uint64{63},
			want: "1000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name: "bit across word boundary",
			bits: []uint64{63, 64},
			want: "11000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name: "set same bit twice",
			bits: []uint64{2, 2},
			want: "100",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var b bitset.BitSet

			// Act
			for _, v := range tc.bits {
				b.Set(v)
			}

			// Assert
			if got := b.String(); got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestBitSetIsSet(t *testing.T) {
	var (
		b      bitset.BitSet
		values = []uint64{0, 5, 63, 64, 200}
	)

	for _, v := range values {
		b.Set(v)
	}

	for _, v := range values {
		if !b.IsSet(v) {
			t.Errorf("bit %d should be set", v)
		}
	}

	if b.IsSet(999) {
		t.Error("bit 999 should not be set")
	}
}

func TestBitSetClear(t *testing.T) {
	var b bitset.BitSet
	b.Set(10)
	b.Clear(10)

	if b.IsSet(10) {
		t.Error("bit 10 should be cleared")
	}

	// Clearing a bit beyond the slice should not panic
	b.Clear(999)
}

func TestBitSetClearAll(t *testing.T) {
	// Arrange
	var b bitset.BitSet
	b.Set(1)
	b.Set(64)
	b.Set(200)

	// Act
	b.ClearAll()

	// Assert
	if b.Len() != 0 {
		t.Errorf("expected len 0 after ClearAll, got %d", b.Len())
	}

	if b.IsSet(1) || b.IsSet(64) || b.IsSet(200) {
		t.Error("no bits should be set after ClearAll")
	}
}

func TestBitSetCompact(t *testing.T) {
	cases := []struct {
		name     string
		setup    func() bitset.BitSet
		wantSize int
	}{
		{
			name: "clear all",
			setup: func() bitset.BitSet {
				var b bitset.BitSet
				b.Set(1)
				b.Set(64)
				b.Set(200)
				b.ClearAll()
				return b
			},
			wantSize: 0,
		},
		{
			name: "trailing zeros",
			setup: func() bitset.BitSet {
				var a, other bitset.BitSet
				a.Set(1)
				a.Set(128)
				other.Set(1)
				a.And(other)
				return a
			},
			wantSize: 64,
		},
		{
			name: "no trailing zeros",
			setup: func() bitset.BitSet {
				var b bitset.BitSet
				b.Set(1)
				b.Set(64)
				return b
			},
			wantSize: 128,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			b := tc.setup()

			// Act
			b.Compact()

			// Assert
			if got := b.Size(); got != tc.wantSize {
				t.Errorf("expected size %d after Compact, got %d", tc.wantSize, got)
			}
		})
	}
}

func TestBitSetSize(t *testing.T) {
	cases := []struct {
		name  string
		setup func() bitset.BitSet
		want  int
	}{
		{
			name:  "empty",
			setup: func() bitset.BitSet { return bitset.BitSet{} },
			want:  0,
		},
		{
			name:  "new 128",
			setup: func() bitset.BitSet { return bitset.New(128) },
			want:  128,
		},
		{
			name: "set bit 0",
			setup: func() bitset.BitSet {
				var b bitset.BitSet
				b.Set(0)
				return b
			},
			want: 64,
		},
		{
			name: "set bit 200",
			setup: func() bitset.BitSet {
				var b bitset.BitSet
				b.Set(200)
				return b
			},
			want: 256,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			b := tc.setup()

			// Act
			got := b.Size()

			// Assert
			if got != tc.want {
				t.Errorf("expected size %d, got %d", tc.want, got)
			}
		})
	}
}

func TestBitSetLen(t *testing.T) {
	var b bitset.BitSet
	if b.Len() != 0 {
		t.Error("empty bitset should have len 0")
	}

	b.Set(0)
	b.Set(1)
	b.Set(100)
	if b.Len() != 3 {
		t.Errorf("expected len 3, got %d", b.Len())
	}
}

func TestBitSetNew(t *testing.T) {
	b := bitset.New(0)
	if b.Size() != 0 {
		t.Errorf("expected size 0, got %d", b.Size())
	}

	b = bitset.New(128)
	if b.Size() != 128 {
		t.Errorf("expected size 128, got %d", b.Size())
	}
}

func TestBitSetOr(t *testing.T) {
	cases := []struct {
		name    string
		current []uint64
		other   []uint64
		want    string
	}{
		{
			name:    "empty sets",
			current: []uint64{},
			other:   []uint64{},
			want:    "",
		},
		{
			name:    "equal sets",
			current: []uint64{1, 4, 8},
			other:   []uint64{1, 4, 8},
			want:    "100010010",
		},
		{
			name:    "different sets",
			current: []uint64{1, 4, 8},
			other:   []uint64{2, 3, 5},
			want:    "100111110",
		},
		{
			name:    "current smaller than other",
			current: []uint64{1, 4},
			other:   []uint64{2, 3, 5, 8},
			want:    "100111110",
		},
		{
			name:    "current bigger than other",
			current: []uint64{2, 3, 5, 8},
			other:   []uint64{1, 4},
			want:    "100111110",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var a, b bitset.BitSet

			for _, v := range tc.current {
				a.Set(v)
			}

			for _, v := range tc.other {
				b.Set(v)
			}

			// Act
			a.Or(b)

			// Assert
			if got := a.String(); got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestBitSetAnd(t *testing.T) {
	cases := []struct {
		name    string
		current []uint64
		other   []uint64
		want    string
	}{
		{
			name:    "empty sets",
			current: []uint64{},
			other:   []uint64{},
			want:    "",
		},
		{
			name:    "equal sets",
			current: []uint64{1, 4, 8},
			other:   []uint64{1, 4, 8},
			want:    "100010010",
		},
		{
			name:    "different sets",
			current: []uint64{1, 4, 8},
			other:   []uint64{2, 3, 5},
			want:    "0",
		},
		{
			name:    "current smaller than other",
			current: []uint64{1, 4},
			other:   []uint64{1, 4, 5, 8},
			want:    "10010",
		},
		{
			name:    "current bigger than other",
			current: []uint64{0, 3, 5, 8},
			other:   []uint64{0, 3},
			want:    "1001",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var a, b bitset.BitSet

			for _, v := range tc.current {
				a.Set(v)
			}

			for _, v := range tc.other {
				b.Set(v)
			}

			// Act
			a.And(b)

			// Assert
			if got := a.String(); got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestBitSetXor(t *testing.T) {
	cases := []struct {
		name    string
		current []uint64
		other   []uint64
		want    string
	}{
		{
			name:    "empty sets",
			current: []uint64{},
			other:   []uint64{},
			want:    "",
		},
		{
			name:    "equal sets",
			current: []uint64{1, 4, 8},
			other:   []uint64{1, 4, 8},
			want:    "0",
		},
		{
			name:    "different sets",
			current: []uint64{1, 3, 8},
			other:   []uint64{2, 3, 5},
			want:    "100100110",
		},
		{
			name:    "current smaller than other",
			current: []uint64{1, 3, 4},
			other:   []uint64{2, 3, 5, 8},
			want:    "100110110",
		},
		{
			name:    "current bigger than other",
			current: []uint64{2, 3, 5, 8},
			other:   []uint64{1, 3},
			want:    "100100110",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var a, b bitset.BitSet

			for _, v := range tc.current {
				a.Set(v)
			}

			for _, v := range tc.other {
				b.Set(v)
			}

			// Act
			a.Xor(b)

			// Assert
			if got := a.String(); got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestBitSetEqual(t *testing.T) {
	cases := []struct {
		name  string
		setup func() (a, b bitset.BitSet)
		want  bool
	}{
		{
			name:  "empty",
			setup: func() (a, b bitset.BitSet) { return a, b },
			want:  true,
		},
		{
			name: "equal",
			setup: func() (a, b bitset.BitSet) {
				a.Set(1)
				a.Set(100)
				b.Set(1)
				b.Set(100)
				return a, b
			},
			want: true,
		},
		{
			name: "equal with different sizes",
			setup: func() (a, b bitset.BitSet) {
				b = bitset.New(256)
				b.Set(5)
				a.Set(5)
				return a, b
			},
			want: true,
		},
		{
			name: "not equal",
			setup: func() (a, b bitset.BitSet) {
				a.Set(1)
				b.Set(1)
				b.Set(50)
				return a, b
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			a, b := tc.setup()

			// Act
			got := a.Equal(b)

			// Assert
			if tc.want == got {
				return
			}

			if tc.want {
				t.Error("bitsets should be equal")
			} else {
				t.Error("bitsets should not be equal")
			}
		})
	}
}

func TestBitSetString(t *testing.T) {
	cases := []struct {
		name   string
		values []uint64
		want   string
	}{
		{"empty", nil, ""},
		{"single bit", []uint64{0}, "1"},
		{"max word value", []uint64{63}, "1000000000000000000000000000000000000000000000000000000000000000"},
		{"multiple sequential bits", []uint64{0, 1, 2, 3}, "1111"},
		{"multiple non sequential bits", []uint64{0, 2, 4}, "10101"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var b bitset.BitSet
			for _, v := range tc.values {
				b.Set(v)
			}

			// Act
			got := b.String()

			// Assert
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestBitSetPaddedString(t *testing.T) {
	cases := []struct {
		name   string
		values []uint64
		want   string
	}{
		{"empty", nil, ""},
		{"single bit", []uint64{0}, "0000000000000000000000000000000000000000000000000000000000000001"},
		{"max word value", []uint64{63}, "1000000000000000000000000000000000000000000000000000000000000000"},
		{"multiple sequential bits", []uint64{0, 1, 2, 3}, "0000000000000000000000000000000000000000000000000000000000001111"},
		{"multiple non sequential bits", []uint64{0, 2, 4}, "0000000000000000000000000000000000000000000000000000000000010101"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var b bitset.BitSet
			for _, v := range tc.values {
				b.Set(v)
			}

			// Act
			got := b.PaddedString()

			// Assert
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
