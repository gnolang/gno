// ported from github.com/mempooler/int256
package int256

import (
	"testing"
)

func TestSign(t *testing.T) {
	tests := []struct {
		x    string
		want int
	}{
		{"0", 0},
		{"1", 1},
		{"-1", -1},
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)
		got := z.Sign()
		if got != tc.want {
			t.Errorf("Sign(%s) = %d, want %d", tc.x, got, tc.want)
		}
	}
}
