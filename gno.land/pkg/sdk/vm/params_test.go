package vm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParamsString verifies the output of the String method.
func TestParamsString(t *testing.T) {
	p := Params{
		SysNamesPkgPath: "gno.land/r/sys/names", // XXX what is this really for now
		SysNamesEnabled: true,
		ChainDomain:     "example.com",
	}
	result := p.String()

	// Construct the expected string.
	expected := "Params: \n" +
		fmt.Sprintf("SysNamesPkgPath: %q\n", p.SysNamesPkgPath) +
		fmt.Sprintf("SysNamesEnabled: %t\n", p.SysNamesEnabled) +
		fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain) +
		fmt.Sprintf("DefaultDeposit: %q\n", p.DefaultDeposit) +
		fmt.Sprintf("StoragePrice: %q\n", p.StoragePrice) +
		fmt.Sprintf("StorageFeeCollector: %q\n", p.StorageFeeCollector)

	// Assert: check if the result matches the expected string.
	if result != expected {
		t.Errorf("Params.String() = %q; want %q", result, expected)
	}
}

func TestWillSetParam(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	vmk := env.vmk
	prmk := env.prmk
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
			name:  "update sysnames_pkgpath",
			key:   "sysnames_pkgpath",
			value: "gno.land/r/foo",
			getExpectedValue: func(prms Params) string {
				return prms.SysNamesPkgPath
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		{
			name:  "update chain_domain",
			key:   "chain_domain",
			value: "example.com",
			getExpectedValue: func(prms Params) string {
				return prms.ChainDomain
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		/* unknown parameter keys are OK
		{
			name:             "unknown parameter key panics",
			key:              "unknown_key",
			value:            "some value",
			getExpectedValue: nil,
			shouldPanic:      true,
			isUpdated:        false,
			isEqual:          false, // Not applicable, but included for consistency
		},
		*/
		{
			name:  "non-empty realm does not update params",
			key:   "gno.land/r/sys/params.sysnames_pkgpath",
			value: "gno.land/r/foo",
			getExpectedValue: func(prms Params) string {
				return prms.SysNamesPkgPath // Expect unchanged value
			},
			shouldPanic: false,
			isUpdated:   false,
			isEqual:     false,
		},
		/* XXX add verification in willSetParam().
		{
			name:        "invalid pkgpath panics",
			key:         "sysusers_pkgpath",
			value:       "path/to/pkg",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false, // Not applicable
		},
		{
			name:        "invalid domain panics",
			key:         "chain_domain",
			value:       "example/com",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false, // Not applicable
		},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					prmk.SetString(ctx, "vm:p:"+tt.key, tt.value)
				}, "expected panic for test: %s", tt.name)
			} else {
				prmk.SetString(ctx, "vm:p:"+tt.key, tt.value)
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

func TestWillSetParamBool(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	vmk := env.vmk
	prmk := env.prmk

	// Default should be false
	prms := vmk.GetParams(ctx)
	assert.False(t, prms.SysNamesEnabled, "expected sysnames_enabled to default to false")

	// Set sysnames_enabled to true
	prmk.SetBool(ctx, "vm:p:sysnames_enabled", true)
	prms = vmk.GetParams(ctx)
	assert.True(t, prms.SysNamesEnabled, "expected sysnames_enabled to be true")
}
