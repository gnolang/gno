package uint256

import (
	"testing"
)

type binOpTest struct {
	x, y, want string
}

func TestAdd(t *testing.T) {
	tests := []binOpTest{
		{"0", "1", "1"},
		{"1", "0", "1"},
		{"1", "1", "2"},
		{"1", "3", "4"},
		{"10", "10", "20"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := parseUintFromString(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Add(x, y)

		for i := range got.arr {
			if got.arr[i] != want.arr[i] {
				t.Errorf("Add(%s, %s) = %v, want %v", tc.x, tc.y, got, want)
				break
			}
		}
	}
}

// TODO: need to test negative results
func TestSub(t *testing.T) {
	tests := []binOpTest{
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"10", "10", "0"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := parseUintFromString(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Sub(x, y)

		for i := range got.arr {
			if got.arr[i] != want.arr[i] {
				t.Errorf("Sub(%s, %s) = %v, want %v", tc.x, tc.y, got, want)
				break
			}
		}
	}
}

func TestMul(t *testing.T) {
	tests := []binOpTest{
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"10", "10", "100"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := parseUintFromString(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Mul(x, y)

		for i := range got.arr {
			if got.arr[i] != want.arr[i] {
				t.Errorf("Mul(%s, %s) = %v, want %v", tc.x, tc.y, got, want)
				break
			}
		}
	}
}

func TestDivMod(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		wantDiv string
		wantMod string
	}{
		{"1", "1", "1", "0"},
		{"10", "10", "1", "0"},
		{"100", "10", "10", "0"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		wantDiv, err := parseUintFromString(tc.wantDiv)
		if err != nil {
			t.Error(err)
			continue
		}

		wantMod, err := parseUintFromString(tc.wantMod)
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
