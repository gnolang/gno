package vm

import (
	"fmt"
	"reflect"
	"strings"
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
		fmt.Sprintf("SysCLAPkgPath: %q\n", p.SysCLAPkgPath) +
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
		// sysnames_pkgpath
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
			name:        "invalid pkgpath panics",
			key:         "sysnames_pkgpath",
			value:       "path/to/pkg",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		// syscla_pkgpath
		{
			name:  "update syscla_pkgpath",
			key:   "syscla_pkgpath",
			value: "gno.land/r/sys/newcla",
			getExpectedValue: func(prms Params) string {
				return prms.SysCLAPkgPath
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		{
			name:        "invalid syscla_pkgpath panics",
			key:         "syscla_pkgpath",
			value:       "path/to/pkg",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		// chain_domain
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
		{
			name:        "invalid domain panics",
			key:         "chain_domain",
			value:       "example/com",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		// storage_price
		{
			name:  "update storage_price",
			key:   "storage_price",
			value: "200ugnot",
			getExpectedValue: func(prms Params) string {
				return prms.StoragePrice
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		{
			name:        "invalid storage_price panics",
			key:         "storage_price",
			value:       "invalid",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		// default_deposit
		{
			name:  "update default_deposit",
			key:   "default_deposit",
			value: "500000000ugnot",
			getExpectedValue: func(prms Params) string {
				return prms.DefaultDeposit
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		{
			name:        "invalid default_deposit panics",
			key:         "default_deposit",
			value:       "garbage",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		// storage_fee_collector
		{
			name:  "update storage_fee_collector",
			key:   "storage_fee_collector",
			value: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			getExpectedValue: func(prms Params) string {
				return prms.StorageFeeCollector.String()
			},
			shouldPanic: false,
			isUpdated:   true,
			isEqual:     true,
		},
		{
			name:        "invalid storage_fee_collector panics",
			key:         "storage_fee_collector",
			value:       "invalid_address",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
		// unknown
		{
			name:        "unknown param panics",
			key:         "unknown_param",
			value:       "gno.land/r/foo",
			shouldPanic: true,
			isUpdated:   false,
			isEqual:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmk.SetParams(ctx, dps)
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

// TestWillSetParamExhaustive ensures every Params field has a WillSetParam case.
func TestWillSetParamExhaustive(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	vmk := env.vmk
	vmk.SetParams(ctx, DefaultParams())

	call := func(param string) (pnc any) {
		defer func() {
			pnc = recover()
		}()
		vmk.WillSetParam(ctx, param, "")
		return nil
	}

	// baseline: ensure a non-existent key has the expected error.
	const format = "unknown vm param key: %q"
	assert.Equal(t, fmt.Sprintf(format, "p:doesnotexist"), call("p:doesnotexist"))

	typ := reflect.TypeOf(Params{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag, _, _ := strings.Cut(field.Tag.Get("json"), ",")

		t.Run(jsonTag, func(t *testing.T) {
			assert.NotEqual(t, fmt.Sprintf(format, "p:"+jsonTag), call("p:"+jsonTag))
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
