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
		ChainDomain:     "example.com",
	}
	result := p.String()

	// Construct the expected string.
	expected := "Params: \n" +
		fmt.Sprintf("SysUsersPkgPath: %q\n", p.SysNamesPkgPath) +
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
		{
			name:        "invalid pkgpath panics",
			key:         "sysnames_pkgpath",
			value:       "path/to/pkg",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		{
			name:        "invalid domain panics",
			key:         "chain_domain",
			value:       "example/com",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		{
			name:        "invalid storage_price panics",
			key:         "storage_price",
			value:       "invalid",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		{
			name:        "invalid default_deposit panics",
			key:         "default_deposit",
			value:       "garbage",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
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

func TestParamsValidate(t *testing.T) {
	valid := DefaultParams()

	tests := []struct {
		name    string
		modify  func(p Params) Params
		wantErr bool
	}{
		{
			name:    "valid default params",
			modify:  func(p Params) Params { return p },
			wantErr: false,
		},
		{
			name:    "invalid storage_price",
			modify:  func(p Params) Params { p.StoragePrice = "invalid"; return p },
			wantErr: true,
		},
		{
			name:    "empty storage_price",
			modify:  func(p Params) Params { p.StoragePrice = ""; return p },
			wantErr: true,
		},
		{
			name:    "invalid chain_domain",
			modify:  func(p Params) Params { p.ChainDomain = "not/a/domain"; return p },
			wantErr: true,
		},
		{
			name:    "invalid default_deposit",
			modify:  func(p Params) Params { p.DefaultDeposit = "garbage"; return p },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.modify(valid)
			err := p.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
