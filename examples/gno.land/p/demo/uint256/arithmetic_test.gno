package uint256

import "testing"

type binOp2Test struct {
	x, y, want string
}

func TestAdd(t *testing.T) {
	tests := []binOp2Test{
		{"0", "1", "1"},
		{"1", "0", "1"},
		{"1", "1", "2"},
		{"1", "3", "4"},
		{"10", "10", "20"},
		{"18446744073709551615", "18446744073709551615", "36893488147419103230"}, // uint64 overflow
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Add(x, y)

		if got.Neq(want) {
			t.Errorf("Add(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestSub(t *testing.T) {
	tests := []binOp2Test{
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"10", "10", "0"},
		{"31337", "1337", "30000"},
		{"2", "3", "115792089237316195423570985008687907853269984665640564039457584007913129639935"}, // underflow
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Sub(x, y)

		if got.Neq(want) {
			t.Errorf("Sub(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestMul(t *testing.T) {
	tests := []binOp2Test{
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"10", "10", "100"},
		{"18446744073709551615", "2", "36893488147419103230"}, // uint64 overflow
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Mul(x, y)

		if got.Neq(want) {
			t.Errorf("Mul(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestDiv(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "10445"},
		{"31337", "0", "0"},
		{"0", "31337", "0"},
		{"1", "1", "1"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Div(x, y)

		if got.Neq(want) {
			t.Errorf("Div(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestMod(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "2"},
		{"31337", "0", "0"},
		{"0", "31337", "0"},
		{"2", "31337", "2"},
		{"1", "1", "0"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Mod(x, y)

		if got.Neq(want) {
			t.Errorf("Mod(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestDivMod(t *testing.T) {
	tests := []struct {
		x       string
		y       string
		wantDiv string
		wantMod string
	}{
		{"1", "1", "1", "0"},
		{"10", "10", "1", "0"},
		{"100", "10", "10", "0"},
		{"31337", "3", "10445", "2"},
		{"31337", "0", "0", "0"},
		{"0", "31337", "0", "0"},
		{"2", "31337", "0", "2"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		wantDiv, err := FromDecimal(tc.wantDiv)
		if err != nil {
			t.Error(err)
			continue
		}

		wantMod, err := FromDecimal(tc.wantMod)
		if err != nil {
			t.Error(err)
			continue
		}

		gotDiv := new(Uint)
		gotMod := new(Uint)
		gotDiv.DivMod(x, y, gotMod)

		for i := range gotDiv.arr {
			if gotDiv.arr[i] != wantDiv.arr[i] {
				t.Errorf("DivMod(%s, %s) got Div %v, want Div %v", tc.x, tc.y, gotDiv, wantDiv)
				break
			}
		}
		for i := range gotMod.arr {
			if gotMod.arr[i] != wantMod.arr[i] {
				t.Errorf("DivMod(%s, %s) got Mod %v, want Mod %v", tc.x, tc.y, gotMod, wantMod)
				break
			}
		}
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"31337", "115792089237316195423570985008687907853269984665640564039457584007913129608599"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129608599", "31337"},
		{"0", "0"},
		{"2", "115792089237316195423570985008687907853269984665640564039457584007913129639934"},
		{"1", "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Neg(x)

		if got.Neq(want) {
			t.Errorf("Neg(%s) = %v, want %v", tc.x, got.ToString(), want.ToString())
		}
	}
}

func TestExp(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "30773171189753"},
		{"31337", "0", "1"},
		{"0", "31337", "0"},
		{"1", "1", "1"},
		{"2", "3", "8"},
		{"2", "64", "18446744073709551616"},
		{"2", "128", "340282366920938463463374607431768211456"},
		{"2", "255", "57896044618658097711785492504343953926634992332820282019728792003956564819968"},
		{"2", "256", "0"}, // overflow
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Exp(x, y)

		if got.Neq(want) {
			t.Errorf("Exp(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}
