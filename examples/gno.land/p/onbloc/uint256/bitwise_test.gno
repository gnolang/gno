package uint256

import "testing"

type logicOpTest struct {
	name string
	x    Uint
	y    Uint
	want Uint
}

func TestOr(t *testing.T) {
	tests := []logicOpTest{
		{
			name: "all zeros",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "mixed",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "one operand all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := new(Uint).Or(&tt.x, &tt.y)
			if *res != tt.want {
				t.Errorf(
					"Or(%s, %s) = %s, want %s",
					tt.x.String(), tt.y.String(), res.String(), (tt.want).String(),
				)
			}
		})
	}
}

func TestAnd(t *testing.T) {
	tests := []logicOpTest{
		{
			name: "all zeros",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "mixed",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "mixed 2",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "mixed 3",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "one operand zero",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "one operand all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}},
			want: Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := new(Uint).And(&tt.x, &tt.y)
			if *res != tt.want {
				t.Errorf(
					"And(%s, %s) = %s, want %s",
					tt.x.String(), tt.y.String(), res.String(), (tt.want).String(),
				)
			}
		})
	}
}

func TestNot(t *testing.T) {
	tests := []struct {
		name string
		x    Uint
		want Uint
	}{
		{
			name: "all zeros",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "mixed",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}},
			want: Uint{arr: [4]uint64{0, 0, ^uint64(0), ^uint64(0)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := new(Uint).Not(&tt.x)
			if *res != tt.want {
				t.Errorf(
					"Not(%s) = %s, want %s",
					tt.x.String(), res.String(), (tt.want).String(),
				)
			}
		})
	}
}

func TestAndNot(t *testing.T) {
	tests := []logicOpTest{
		{
			name: "all zeros",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "mixed",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "mixed 2",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "mixed 3",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}},
		},
		{
			name: "one operand zero",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "one operand all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}},
			want: Uint{arr: [4]uint64{0xAAAAAAAAAAAAAAAA, 0x5555555555555555, 0x0000000000000000, ^uint64(0)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := new(Uint).AndNot(&tt.x, &tt.y)
			if *res != tt.want {
				t.Errorf(
					"AndNot(%s, %s) = %s, want %s",
					tt.x.String(), tt.y.String(), res.String(), (tt.want).String(),
				)
			}
		})
	}
}

func TestXor(t *testing.T) {
	tests := []logicOpTest{
		{
			name: "all zeros",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "mixed",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "mixed 2",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "mixed 3",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}},
			y:    Uint{arr: [4]uint64{0, 0, ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
		},
		{
			name: "one operand zero",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}},
			want: Uint{arr: [4]uint64{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}},
		},
		{
			name: "one operand all ones",
			x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}},
			want: Uint{arr: [4]uint64{0xAAAAAAAAAAAAAAAA, 0x5555555555555555, 0x0000000000000000, ^uint64(0)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := new(Uint).Xor(&tt.x, &tt.y)
			if *res != tt.want {
				t.Errorf(
					"Xor(%s, %s) = %s, want %s",
					tt.x.String(), tt.y.String(), res.String(), (tt.want).String(),
				)
			}
		})
	}
}

func TestLsh(t *testing.T) {
	tests := []struct {
		x    string
		y    uint
		want string
	}{
		{"0", 0, "0"},
		{"0", 1, "0"},
		{"0", 64, "0"},
		{"1", 0, "1"},
		{"1", 1, "2"},
		{"1", 64, "18446744073709551616"},
		{"1", 128, "340282366920938463463374607431768211456"},
		{"1", 192, "6277101735386680763835789423207666416102355444464034512896"},
		{"1", 255, "57896044618658097711785492504343953926634992332820282019728792003956564819968"},
		{"1", 256, "0"},
		{"31337", 0, "31337"},
		{"31337", 1, "62674"},
		{"31337", 64, "578065619037836218990592"},
		{"31337", 128, "10663428532201448629551770073089320442396672"},
		{"31337", 192, "196705537081812415096322133155058642481399512563169449530621952"},
		{"31337", 193, "393411074163624830192644266310117284962799025126338899061243904"},
		{"31337", 255, "57896044618658097711785492504343953926634992332820282019728792003956564819968"},
		{"31337", 256, "0"},
		// 64 < n < 128
		{"1", 65, "36893488147419103232"},
		{"31337", 100, "39724366859352024754702188346867712"},

		// 128 < n < 192
		{"1", 129, "680564733841876926926749214863536422912"},
		{"31337", 150, "44725660946326664792723507424638829088826130956288"},

		// 192 < n < 256
		{"1", 193, "12554203470773361527671578846415332832204710888928069025792"},
		{"31337", 200, "50356617492943978264658466087695012475238275216171379079839219712"},

		// n > 256
		{"1", 257, "0"},
		{"31337", 300, "0"},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		want := MustFromDecimal(tt.want)

		got := new(Uint).Lsh(x, tt.y)

		if got.Neq(want) {
			t.Errorf("Lsh(%s, %d) = %s, want %s", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestRsh(t *testing.T) {
	tests := []struct {
		x    string
		y    uint
		want string
	}{
		{"0", 0, "0"},
		{"0", 1, "0"},
		{"0", 64, "0"},
		{"1", 0, "1"},
		{"1", 1, "0"},
		{"1", 64, "0"},
		{"1", 128, "0"},
		{"1", 192, "0"},
		{"1", 255, "0"},
		{"57896044618658097711785492504343953926634992332820282019728792003956564819968", 255, "1"},
		{"6277101735386680763835789423207666416102355444464034512896", 192, "1"},
		{"340282366920938463463374607431768211456", 128, "1"},
		{"18446744073709551616", 64, "1"},
		{"393411074163624830192644266310117284962799025126338899061243904", 193, "31337"},
		{"196705537081812415096322133155058642481399512563169449530621952", 192, "31337"},
		{"10663428532201448629551770073089320442396672", 128, "31337"},
		{"578065619037836218990592", 64, "31337"},
		{twoPow256Sub1, 256, "0"},
		// outliers
		{"340282366920938463463374607431768211455", 129, "0"},
		{"18446744073709551615", 65, "0"},
		{twoPow256Sub1, 1, "57896044618658097711785492504343953926634992332820282019728792003956564819967"},

		// n > 256
		{"1", 257, "0"},
		{"31337", 300, "0"},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)

		want := MustFromDecimal(tt.want)
		got := new(Uint).Rsh(x, tt.y)

		if got.Neq(want) {
			t.Errorf("Rsh(%s, %d) = %s, want %s", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestSRsh(t *testing.T) {
	tests := []struct {
		x    string
		y    uint
		want string
	}{
		// Positive numbers (behaves like Rsh)
		{"0x0", 0, "0x0"},
		{"0x0", 1, "0x0"},
		{"0x1", 0, "0x1"},
		{"0x1", 1, "0x0"},
		{"0x31337", 0, "0x31337"},
		{"0x31337", 4, "0x3133"},
		{"0x31337", 8, "0x313"},
		{"0x31337", 16, "0x3"},
		{"0x10000000000000000", 64, "0x1"}, // 2^64 >> 64

		// // Numbers with MSB set (negative numbers in two's complement)
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 1, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 4, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 64, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 128, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 192, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 255, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},

		// Large positive number close to max value
		{"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 1, "0x3fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 2, "0x1fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 64, "0x7fffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 128, "0x7fffffffffffffffffffffffffffffff"},
		{"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 192, "0x7fffffffffffffff"},
		{"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 255, "0x0"},

		// Specific cases
		{"0x8000000000000000000000000000000000000000000000000000000000000000", 1, "0xc000000000000000000000000000000000000000000000000000000000000000"},
		{"0x8000000000000000000000000000000000000000000000000000000000000001", 1, "0xc000000000000000000000000000000000000000000000000000000000000000"},

		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 65, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 127, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 129, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 193, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},

		// n > 256
		{"0x1", 257, "0x0"},
		{"0x31337", 300, "0x0"},
	}

	for _, tt := range tests {
		x := MustFromHex(tt.x)
		want := MustFromHex(tt.want)

		got := new(Uint).SRsh(x, tt.y)

		if !got.Eq(want) {
			t.Errorf("SRsh(%s, %d) = %s, want %s", tt.x, tt.y, got.String(), want.String())
		}
	}
}
