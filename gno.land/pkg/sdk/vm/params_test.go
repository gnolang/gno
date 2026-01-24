package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	assert.Equal(t, sysNamesPkgDefault, params.SysNamesPkgPath)
	assert.Equal(t, chainDomainDefault, params.ChainDomain)
	assert.Equal(t, depositDefault, params.DefaultDeposit)
	assert.Equal(t, storagePriceDefault, params.StoragePrice)
	assert.Equal(t, crypto.AddressFromPreimage([]byte(storageFeeCollectorNameDefault)), params.StorageFeeCollector)
}

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

func TestParamsValidate(t *testing.T) {
	validAddr := crypto.AddressFromPreimage([]byte("valid"))
	zeroAddr := crypto.Address{}

	tests := []struct {
		name    string
		params  Params
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default params",
			params:  DefaultParams(),
			wantErr: false,
		},
		{
			name:    "valid params with empty sysnames path",
			params:  NewParams("", "example.com", "1000ugnot", "10ugnot", validAddr),
			wantErr: false,
		},
		{
			name:    "valid params with empty chain domain",
			params:  NewParams("gno.land/r/sys/names", "", "1000ugnot", "10ugnot", validAddr),
			wantErr: false,
		},
		{
			name:    "invalid sysnames package path - not userlib",
			params:  NewParams("invalid/path", "example.com", "1000ugnot", "10ugnot", validAddr),
			wantErr: true,
			errMsg:  "invalid user package path",
		},
		{
			name:    "invalid chain domain - special characters",
			params:  NewParams("gno.land/r/sys/names", "invalid@domain", "1000ugnot", "10ugnot", validAddr),
			wantErr: true,
			errMsg:  "invalid chain domain",
		},
		{
			name:    "invalid chain domain - no TLD",
			params:  NewParams("gno.land/r/sys/names", "invalid", "1000ugnot", "10ugnot", validAddr),
			wantErr: true,
			errMsg:  "invalid chain domain",
		},
		{
			name:    "invalid default deposit - empty",
			params:  NewParams("gno.land/r/sys/names", "example.com", "", "10ugnot", validAddr),
			wantErr: true,
			errMsg:  "invalid default storage deposit",
		},
		{
			name:    "invalid default deposit - malformed",
			params:  NewParams("gno.land/r/sys/names", "example.com", "invalid", "10ugnot", validAddr),
			wantErr: true,
			errMsg:  "invalid default storage deposit",
		},
		{
			name:    "invalid storage price - empty",
			params:  NewParams("gno.land/r/sys/names", "example.com", "1000ugnot", "", validAddr),
			wantErr: true,
			errMsg:  "invalid storage price",
		},
		{
			name:    "invalid storage price - malformed",
			params:  NewParams("gno.land/r/sys/names", "example.com", "1000ugnot", "invalid", validAddr),
			wantErr: true,
			errMsg:  "invalid storage price",
		},
		{
			name:    "invalid storage price - different denomination",
			params:  NewParams("gno.land/r/sys/names", "example.com", "1000ugnot", "10uatom", validAddr),
			wantErr: true,
			errMsg:  "storage price \"10uatom\" coins must be a subset of default deposit \"1000ugnot\" coins",
		},
		{
			name:    "invalid storage fee collector - zero address",
			params:  NewParams("gno.land/r/sys/names", "example.com", "1000ugnot", "10ugnot", zeroAddr),
			wantErr: true,
			errMsg:  "invalid storage fee collector, cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVMKeeperWillSetParam(t *testing.T) {
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
