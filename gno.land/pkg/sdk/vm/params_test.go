package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
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

func TestWillSetParam_ValsetUpdate(t *testing.T) {
	t.Parallel()

	t.Run("no valset update", func(t *testing.T) {
		t.Parallel()

		var (
			env = setupTestEnv()
			ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
		)

		assert.NotPanics(t, func() {
			env.vmk.WillSetParam(ctx, "random key", nil)
		})
	})

	t.Run("malformed valset update key", func(t *testing.T) {
		t.Parallel()

		var (
			env = setupTestEnv()
			ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
		)

		assert.Panics(t, func() {
			env.vmk.WillSetParam(ctx, valsetUpdatesParam+":10-10", []string{"value"})
		})
	})

	t.Run("invalid valset value type", func(t *testing.T) {
		t.Parallel()

		var (
			env = setupTestEnv()
			ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
		)

		assert.Panics(t, func() {
			env.vmk.WillSetParam(ctx, valsetUpdatesParam+":10", "single value")
		})
	})

	t.Run("invalid valset update values", func(t *testing.T) {
		t.Parallel()

		var (
			key = secp256k1.GenPrivKey()

			pubKey  = key.PubKey()
			address = pubKey.Address()
			power   = 10
		)

		var (
			env = setupTestEnv()
			ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
		)

		testTable := []struct {
			name   string
			change string
		}{
			{
				"invalid serialization",
				"addr:pubkey:power:random", // 4 parts
			},
			{
				"invalid bech address",
				fmt.Sprintf(
					"%s:%s:%d",
					"validvalidvalid", // invalid address
					pubKey,
					power,
				),
			},
			{
				"invalid public key",
				fmt.Sprintf(
					"%s:%s:%d",
					address,
					"validvalidvalid", // invalid pubkey
					power,
				),
			},
			{
				"address / public key mismatch",
				fmt.Sprintf(
					"%s:%s:%d",
					address,
					secp256k1.GenPrivKey().PubKey(), // different pubkey
					power,
				),
			},
			{
				"invalid voting power",
				fmt.Sprintf(
					"%s:%s:%s",
					address,
					pubKey,
					"abc123", // invalid voting power
				),
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				assert.Panics(t, func() {
					env.vmk.WillSetParam(
						ctx,
						fmt.Sprintf("%s:%d", valsetUpdatesParam, 10),
						[]string{testCase.change},
					)
				})
			})
		}
	})

	t.Run("valid valset update values", func(t *testing.T) {
		t.Parallel()

		var (
			key = secp256k1.GenPrivKey()

			pubKey  = key.PubKey()
			address = pubKey.Address()
			power   = 10
		)

		var (
			env = setupTestEnv()
			ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
		)

		var (
			paramKey   = fmt.Sprintf("%s:%d", valsetUpdatesParam, 10)
			paramValue = []string{
				fmt.Sprintf(
					"%s:%s:%d",
					address,
					pubKey,
					power,
				),
			}
		)

		assert.NotPanics(t, func() {
			env.vmk.WillSetParam(
				ctx,
				paramKey,
				paramValue,
			)
		})
	})
}
