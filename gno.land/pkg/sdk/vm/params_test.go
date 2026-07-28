package vm

import (
	"fmt"
	"reflect"
	"strings"
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
		fmt.Sprintf("SysCLAPkgPath: %q\n", p.SysCLAPkgPath) +
		fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain) +
		fmt.Sprintf("DefaultDeposit: %q\n", p.DefaultDeposit) +
		fmt.Sprintf("StoragePrice: %q\n", p.StoragePrice) +
		fmt.Sprintf("StorageFeeCollector: %q\n", p.StorageFeeCollector) +
		fmt.Sprintf("MinGetReadDepth100: %d\n", p.MinGetReadDepth100) +
		fmt.Sprintf("MinSetReadDepth100: %d\n", p.MinSetReadDepth100) +
		fmt.Sprintf("MinWriteDepth100: %d\n", p.MinWriteDepth100) +
		fmt.Sprintf("FixedGetReadDepth100: %d\n", p.FixedGetReadDepth100) +
		fmt.Sprintf("FixedSetReadDepth100: %d\n", p.FixedSetReadDepth100) +
		fmt.Sprintf("FixedWriteDepth100: %d\n", p.FixedWriteDepth100) +
		fmt.Sprintf("IterNextCostFlat: %d\n", p.IterNextCostFlat) +
		fmt.Sprintf("CodeSubmissionPolicy: %q\n", p.CodeSubmissionPolicy) +
		fmt.Sprintf("CodeSubmitters: %v\n", p.CodeSubmitters) +
		fmt.Sprintf("PreprocessGasPerByte: %d\n", p.PreprocessGasPerByte)

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

	typ := reflect.TypeFor[Params]()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag, _, _ := strings.Cut(field.Tag.Get("json"), ",")

		t.Run(jsonTag, func(t *testing.T) {
			assert.NotEqual(t, fmt.Sprintf(format, "p:"+jsonTag), call("p:"+jsonTag))
		})
	}
}

// A vm params blob written before PreprocessGasPerByte existed decodes with
// the field zero (simulated here by writing 0 directly, past Validate).
// GetParams must default it so that (a) the type-check/preprocess charge
// stays active on legacy state, and (b) WillSetParam's whole-struct
// re-validation does not reject updates of unrelated params on such state.
func TestGetParamsDefaultsPreprocessGasPerByte(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	legacy := DefaultParams()
	legacy.PreprocessGasPerByte = 0
	env.prmk.SetStruct(ctx, "vm:p", legacy) // direct write: no Validate, no hooks

	assert.Equal(t, preprocessGasPerByteDefault, env.vmk.GetParams(ctx).PreprocessGasPerByte)

	// The trap the defaulting closes: a params-keeper write of an unrelated
	// param runs WillSetParam, which re-validates the whole struct read via
	// GetParams and would panic on a zero PreprocessGasPerByte.
	assert.NotPanics(t, func() {
		env.prmk.SetString(ctx, "vm:p:chain_domain", "example.com")
	})
}

// A relaunch genesis exported by a binary predating PreprocessGasPerByte omits
// the field, so it decodes as zero. ValidateGenesis and InitGenesis must
// tolerate that (defaulting it) rather than rejecting the genesis — matching
// GetParams' runtime behavior — while a genesis with an explicitly invalid
// value still fails.
func TestGenesisToleratesLegacyPreprocessGasPerByte(t *testing.T) {
	legacy := DefaultParams()
	legacy.PreprocessGasPerByte = 0 // field absent in a pre-field export
	assert.NoError(t, ValidateGenesis(NewGenesisState(legacy)))

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	assert.NotPanics(t, func() { env.vmk.InitGenesis(ctx, NewGenesisState(legacy)) })
	assert.Equal(t, preprocessGasPerByteDefault, env.vmk.GetParams(ctx).PreprocessGasPerByte)

	// An explicitly out-of-range value is still rejected (not treated as legacy).
	bad := DefaultParams()
	bad.PreprocessGasPerByte = -1
	assert.Error(t, ValidateGenesis(NewGenesisState(bad)))
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

// TestDefaultParams pins the depth-gas defaults for the mounted reference
// store (B+32 + fast index) and guards a keeper SetParams→GetParams
// round-trip of the exact default values.
func TestDefaultParams(t *testing.T) {
	p := DefaultParams()

	assert.Equal(t, int64(100), p.FixedGetReadDepth100, "GET pinned: 1.0 flat read via the fast index")
	assert.Equal(t, int64(200), p.FixedSetReadDepth100, "SET-read pinned at the measured-with-cache descent")
	assert.Equal(t, int64(540), p.FixedWriteDepth100, "WRITE pinned: 4.4 batched COW + 1.0 index write")
	assert.Equal(t, p.FixedGetReadDepth100, p.MinGetReadDepth100, "NewParams pins Fixed = Min")
	assert.Equal(t, p.FixedSetReadDepth100, p.MinSetReadDepth100, "NewParams pins Fixed = Min")
	assert.Equal(t, p.FixedWriteDepth100, p.MinWriteDepth100, "NewParams pins Fixed = Min")
	assert.Equal(t, int64(1_000), p.IterNextCostFlat)
	assert.NoError(t, p.Validate())

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	env.vmk.SetParams(ctx, p)
	got := env.vmk.GetParams(ctx)
	assert.Equal(t, p, got, "defaults must round-trip through the params keeper unchanged")
}
