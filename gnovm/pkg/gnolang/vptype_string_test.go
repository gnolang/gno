package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVPTypeString(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		desc     string
		in       VPType
		expected string
	}{
		{"VPUverse", 0, "VPUverse"},
		{"VPBlock", 1, "VPBlock"},
		{"VPField", 2, "VPField"},
		{"VPValMethod", 3, "VPValMethod"},
		{"VPPtrMethod", 4, "VPPtrMethod"},
		{"VPInterface", 5, "VPInterface"},
		{"VPSubrefField", 6, "VPSubrefField"},
		{"VPType(7)", 7, "VPType(7)"},
		{"VPType(8)", 8, "VPType(8)"},
		{"VPType(9)", 9, "VPType(9)"},
		{"VPType(10)", 10, "VPType(10)"},
		{"VPType(11)", 11, "VPType(11)"},
		{"VPType(12)", 12, "VPType(12)"},
		{"VPType(13)", 13, "VPType(13)"},
		{"VPType(14)", 14, "VPType(14)"},
		{"VPType(15)", 15, "VPType(15)"},
		{"VPType(16)", 16, "VPType(16)"},
		{"VPType(17)", 17, "VPType(17)"},
		{"VPDerefField", 18, "VPDerefField"},
		{"VPDerefValMethod", 19, "VPDerefValMethod"},
		{"VPDerefPtrMethod", 20, "VPDerefPtrMethod"},
		{"VPDerefInterface", 21, "VPDerefInterface"},
		{"VPType(22)", 22, "VPType(22)"},
		{"VPType(23)", 23, "VPType(23)"},
		{"VPType(24)", 24, "VPType(24)"},
		{"VPType(25)", 25, "VPType(25)"},
		{"VPType(26)", 26, "VPType(26)"},
		{"VPType(27)", 27, "VPType(27)"},
		{"VPType(28)", 28, "VPType(28)"},
		{"VPType(29)", 29, "VPType(29)"},
		{"VPType(30)", 30, "VPType(30)"},
		{"VPType(31)", 31, "VPType(31)"},
		{"VPNative", 32, "VPNative"},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			actual := tt.in.String()
			if actual != tt.expected {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
