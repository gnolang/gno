package bitset_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/bitset"
)

func TestBitSetTest(t *testing.T) {
	var (
		b      bitset.BitSet
		values = []int{0, 5, 63, 64, 200}
	)

	for _, v := range values {
		b.Set(v)
	}

	for _, v := range values {
		if !b.Test(v) {
			t.Errorf("bit %d should be set", v)
		}
	}

	if b.Test(999) {
		t.Error("bit 999 should not be set")
	}
}

func TestBitSetClear(t *testing.T) {
	var b bitset.BitSet
	b.Set(10)
	b.Clear(10)

	if b.Test(10) {
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

	if b.Test(1) || b.Test(64) || b.Test(200) {
		t.Error("no bits should be set after ClearAll")
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
	// Arrange
	var a, b bitset.BitSet

	a.Set(1)
	a.Set(3)

	b.Set(2)
	b.Set(3)
	b.Set(200)

	// Act
	a.Or(b)

	// Assert
	for _, i := range []int{1, 2, 3, 200} {
		if !a.Test(i) {
			t.Errorf("bit %d should be set after union", i)
		}
	}

	if a.Test(0) {
		t.Error("bit 0 should not be set")
	}
}

func TestBitSetAnd(t *testing.T) {
	// Arrange
	var a, b bitset.BitSet

	a.Set(1)
	a.Set(2)
	a.Set(3)

	b.Set(2)
	b.Set(3)
	b.Set(4)

	// Act
	a.And(b)

	// Assert
	if a.Test(1) {
		t.Error("bit 1 should be cleared after AND")
	}

	if !a.Test(2) {
		t.Error("bit 2 should remain set")
	}

	if !a.Test(3) {
		t.Error("bit 3 should remain set")
	}

	if a.Test(4) {
		t.Error("bit 4 should not be set (was not in a)")
	}
}

func TestBitSetXor(t *testing.T) {
	// Arrange
	var a, b bitset.BitSet

	a.Set(1)
	a.Set(2)

	b.Set(2)
	b.Set(3)

	// Act
	a.Xor(b)

	// Assert
	if !a.Test(1) {
		t.Error("bit 1 should be set (only in a)")
	}

	if a.Test(2) {
		t.Error("bit 2 should be cleared (in both)")
	}

	if !a.Test(3) {
		t.Error("bit 3 should be set (only in b)")
	}
}

func TestBitSetEquals(t *testing.T) {
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
			got := a.Equals(b)

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
		values []int
		want   string
	}{
		{"empty", nil, ""},
		{"single bit", []int{0}, "1"},
		{"multiple sequential bits", []int{0, 1, 2, 3}, "1111"},
		{"multiple non sequential bits", []int{0, 2, 4}, "10101"},
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
