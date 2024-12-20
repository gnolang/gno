package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAppConfig(t *testing.T) {
	c := DefaultAppConfig()
	c.MinGasPrices = "" // empty

	testCases := []struct {
		testName     string
		minGasPrices string
		expectErr    bool
	}{
		{"invalid min gas prices invalid gas", "10token/1", true},
		{"invalid min gas prices invalid gas denom", "9token/0gs", true},
		{"invalid min gas prices zero gas", "10token/0gas", true},
		{"invalid min gas prices no gas", "10token/gas", true},
		{"invalid min gas prices negtive gas", "10token/-1gas", true},
		{"invalid min gas prices invalid denom", "10$token/2gas", true},
		{"invalid min gas prices invalid second denom", "10token/2gas;10/3gas", true},
		{"valid min gas prices", "10foo/3gas;5bar/3gas", false},
	}

	cfg := DefaultAppConfig()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			cfg.MinGasPrices = tc.minGasPrices
			assert.Equal(t, tc.expectErr, cfg.ValidateBasic() != nil)
		})
	}
}
