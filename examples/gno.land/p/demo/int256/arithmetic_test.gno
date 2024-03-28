package int256

import "testing"

type binOp2Test struct {
	x, y, want string
}

func TestAdd(t *testing.T) {
	tests := []binOp2Test{
		{"0", "1", "1"},
		{"1", "0", "1"},
		{"1", "1", "2"},
		{"1", "2", "3"},
		// NEGATIVE
		{"-1", "1", "-0"}, // TODO: remove negative sign for 0 ??
		{"1", "-1", "0"},
		{"-1", "-1", "-2"},
		{"-1", "-2", "-3"},
		{"-1", "3", "2"},
		{"3", "-1", "2"},
		// OVERFLOW
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "1", "0"},
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

		got := New()
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
		{"-1", "1", "-2"},
		{"1", "-1", "2"},
		{"-1", "-1", "-0"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", "-0"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "0", "-115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{x: "-115792089237316195423570985008687907853269984665640564039457584007913129639935", y: "1", want: "-0"},
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

		got := New()
		got.Sub(x, y)

		if got.Neq(want) {
			t.Errorf("Sub(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestMul(t *testing.T) {
	tests := []binOp2Test{
		{"5", "3", "15"},
		{"-5", "3", "-15"},
		{"5", "-3", "-15"},
		{"0", "3", "0"},
		{"3", "0", "0"},
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

		got := New()
		got.Mul(x, y)

		if got.Neq(want) {
			t.Errorf("Mul(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestDiv(t *testing.T) {
	tests := []binOp2Test{
		{"0", "1", "0"},
		{"0", "-1", "-0"},
		{"10", "0", "0"},
		{"10", "1", "10"},
		{"10", "-1", "-10"},
		{"-10", "0", "-0"},
		{"-10", "1", "-10"},
		{"-10", "-1", "10"},
		{"10", "-3", "-3"},
		{"10", "3", "3"},
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

		got := New()
		got.Div(x, y)

		if got.Neq(want) {
			t.Errorf("Div(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestRem(t *testing.T) {
	tests := []binOp2Test{
		{"0", "1", "0"},
		{"0", "-1", "0"},
		{"10", "0", "0"},
		{"10", "1", "0"},
		{"10", "-1", "0"},
		{"-10", "0", "0"},
		{"-10", "1", "0"},
		{"-10", "-1", "0"},
		{"10", "3", "1"},
		{"10", "-3", "1"},
		{"-10", "3", "-1"},
		{"-10", "-3", "-1"},
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

		got := New()
		got.Rem(x, y)

		if got.Neq(want) {
			t.Errorf("Rem(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}
