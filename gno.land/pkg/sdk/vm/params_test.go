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
		fmt.Sprintf("StorageFeeCollector: %q\n", p.StorageFeeCollector) +
		fmt.Sprintf("CLAHash: %q\n", p.CLAHash) +
		fmt.Sprintf("CLADocURL: %q\n", p.CLADocURL)

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

func TestValidateCLAHash(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	vmk := env.vmk
	prmk := env.prmk

	tests := []struct {
		name        string
		claParam    string // CLA hash in params
		msgCLAHash  string // CLA hash in message
		expectError bool
		errorType   error
	}{
		{
			name:        "enforcement disabled (empty param), empty message hash",
			claParam:    "",
			msgCLAHash:  "",
			expectError: false,
		},
		{
			name:        "enforcement disabled (empty param), message has hash",
			claParam:    "",
			msgCLAHash:  "somehash12345678",
			expectError: false, // hash is ignored when enforcement is disabled
		},
		{
			name:        "enforcement enabled, matching hash",
			claParam:    "a3d74e2544d091e8",
			msgCLAHash:  "a3d74e2544d091e8",
			expectError: false,
		},
		{
			name:        "enforcement enabled, missing hash in message",
			claParam:    "a3d74e2544d091e8",
			msgCLAHash:  "",
			expectError: true,
			errorType:   CLAHashMissingError{},
		},
		{
			name:        "enforcement enabled, mismatched hash",
			claParam:    "a3d74e2544d091e8",
			msgCLAHash:  "wronghash1234567",
			expectError: true,
			errorType:   CLAHashMismatchError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the CLA hash param
			prmk.SetString(ctx, "vm:p:cla_hash", tt.claParam)

			// Call validateCLAHash
			err := vmk.validateCLAHash(ctx, tt.msgCLAHash)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCLAErrors(t *testing.T) {
	t.Run("CLAHashMismatchError", func(t *testing.T) {
		err := ErrCLAHashMismatch("expected123", "actual456")
		assert.Error(t, err)
		assert.ErrorIs(t, err, CLAHashMismatchError{})

		// Test the base error message
		baseErr := CLAHashMismatchError{}
		assert.Equal(t, "CLA hash mismatch", baseErr.Error())
	})

	t.Run("CLAHashMissingError", func(t *testing.T) {
		err := ErrCLAHashMissing()
		assert.Error(t, err)
		assert.ErrorIs(t, err, CLAHashMissingError{})

		// Test the base error message
		baseErr := CLAHashMissingError{}
		assert.Equal(t, "CLA hash missing", baseErr.Error())
	})
}
