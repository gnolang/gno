package vm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParamsString verifies the output of the String method.
func TestParamsString(t *testing.T) {
	p := Params{
		SysUsersPkgPath: "gno.land/r/sys/users",
		ChainDomain:     "example.com",
	}
	result := p.String()

	// Construct the expected string.
	expected := "Params: \n" +
		fmt.Sprintf("SysUsersPkgPath: %q\n", p.SysUsersPkgPath) +
		fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain)

	// Assert: check if the result matches the expected string.
	if result != expected {
		t.Errorf("Params.String() = %q; want %q", result, expected)
	}
}

func TestWillSetParam(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	vmk := env.vmk
	dps := DefaultParams()

	tests := []struct {
		name             string
		key              string
		value            string
		getExpectedValue func(prms Params) string
		shouldPanic      bool
		isUpdated        bool
		isEqual          bool
	}{
		{
			name:  "update sysusers_pkgpath",
			key:   "sysusers_pkgpath.string",
			value: "gno.land/r/foo",
			getExpectedValue: func(prms Params) string {
				return prms.SysUsersPkgPath
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		{
			name:  "update chain_domain",
			key:   "chain_domain.string",
			value: "example.com",
			getExpectedValue: func(prms Params) string {
				return prms.ChainDomain
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		{
			name:             "unknown parameter key panics",
			key:              "unknown_key",
			value:            "some value",
			getExpectedValue: nil,
			shouldPanic:      true,
			isUpdated:        false,
			isEqual:          false, // Not applicable, but included for consistency
		},
		{
			name:  "non-empty realm does not update params",
			key:   "gno.land/r/sys/params.sysusers_pkgpath.string",
			value: "gno.land/r/foo",
			getExpectedValue: func(prms Params) string {
				return prms.SysUsersPkgPath // Expect unchanged value
			},
			shouldPanic: false,
			isUpdated:   false,
			isEqual:     false,
		},
		{
			name:        "error from SetParams panics",
			key:         "sysusers_pkgpath.string",
			value:       "path/to/pkg",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false, // Not applicable
		},
		{
			name:        "error from prmk.SetParams panics",
			key:         "chain_domain.string",
			value:       "example/com",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false, // Not applicable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					vmk.WillSetParam(ctx, tt.key, tt.value)
				}, "expected panic for test: %s", tt.name)
			} else {
				vmk.WillSetParam(ctx, tt.key, tt.value)
				if tt.getExpectedValue != nil {
					prms := vmk.GetParams(ctx)
					if tt.isUpdated {
						assert.False(t, prms.Equals(dps), "expected values to be different")
					}
					if tt.isEqual {
						actual := tt.getExpectedValue(prms)
						assert.Equal(t, tt.value, actual, "expected values to match")
					}
				}
			}
		})
	}
}
