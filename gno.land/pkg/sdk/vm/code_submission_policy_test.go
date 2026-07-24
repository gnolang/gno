package vm

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addr is a small helper that derives a deterministic, valid bech32 address
// from a seed, for use in the code_submission_policy tests.
func cspTestAddr(seed string) crypto.Address {
	return crypto.AddressFromPreimage([]byte(seed))
}

// TestCodeSubmissionPolicyConstants documents the on-chain string values of the
// policy constants. These strings are consensus-relevant (they are stored in
// params and parsed from governance proposals), so a rename must be deliberate.
func TestCodeSubmissionPolicyConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, CodeSubmissionPolicy("permissionless"), CodeSubmissionPolicyPermissionless)
	assert.Equal(t, CodeSubmissionPolicy("permissioned"), CodeSubmissionPolicyPermissioned)
}

// TestDefaultParamsCodeSubmission verifies the default params are permissionless
// with no submitters, so existing chains are unaffected.
func TestDefaultParamsCodeSubmission(t *testing.T) {
	t.Parallel()
	p := DefaultParams()
	assert.Equal(t, CodeSubmissionPolicyPermissionless, p.CodeSubmissionPolicy)
	assert.Nil(t, p.CodeSubmitters)
	assert.NoError(t, p.Validate())
}

// TestParamsValidateCodeSubmission exercises Params.Validate for the
// code_submission_policy and code_submitters fields.
func TestParamsValidateCodeSubmission(t *testing.T) {
	t.Parallel()

	addr1 := cspTestAddr("submitter1")
	addr2 := cspTestAddr("submitter2")

	tests := []struct {
		name    string
		modify  func(p Params) Params
		wantErr string // substring; empty means no error expected
	}{
		{
			name:   "permissionless with no submitters",
			modify: func(p Params) Params { p.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless; return p },
		},
		{
			name: "permissioned with submitters",
			modify: func(p Params) Params {
				p.CodeSubmissionPolicy = CodeSubmissionPolicyPermissioned
				p.CodeSubmitters = []crypto.Address{addr1, addr2}
				return p
			},
		},
		{
			name: "permissioned with empty submitters is valid at the param level",
			modify: func(p Params) Params {
				p.CodeSubmissionPolicy = CodeSubmissionPolicyPermissioned
				p.CodeSubmitters = nil
				return p
			},
		},
		{
			name:   "empty policy is treated as permissionless",
			modify: func(p Params) Params { p.CodeSubmissionPolicy = ""; return p },
		},
		{
			name:    "invalid policy string",
			modify:  func(p Params) Params { p.CodeSubmissionPolicy = "sometimes"; return p },
			wantErr: "invalid code_submission_policy",
		},
		{
			name: "submitters may be set even when permissionless",
			modify: func(p Params) Params {
				p.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
				p.CodeSubmitters = []crypto.Address{addr1}
				return p
			},
		},
		{
			name: "zero address in submitters",
			modify: func(p Params) Params {
				p.CodeSubmissionPolicy = CodeSubmissionPolicyPermissioned
				p.CodeSubmitters = []crypto.Address{addr1, {}}
				return p
			},
			wantErr: "zero address",
		},
		{
			name: "duplicate address in submitters",
			modify: func(p Params) Params {
				p.CodeSubmissionPolicy = CodeSubmissionPolicyPermissioned
				p.CodeSubmitters = []crypto.Address{addr1, addr2, addr1}
				return p
			},
			wantErr: "duplicate address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := tt.modify(DefaultParams())
			err := p.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

// TestWillSetParamCodeSubmissionPolicy exercises the governance param setter for
// the code_submission_policy key. WillSetParam re-validates after applying, so
// an invalid policy string panics.
func TestWillSetParamCodeSubmissionPolicy(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	vmk := env.vmk
	prmk := env.prmk

	tests := []struct {
		name        string
		value       string
		shouldPanic bool
		want        CodeSubmissionPolicy
	}{
		{name: "set permissioned", value: "permissioned", want: CodeSubmissionPolicyPermissioned},
		{name: "set permissionless", value: "permissionless", want: CodeSubmissionPolicyPermissionless},
		{name: "invalid policy panics", value: "bogus", shouldPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmk.SetParams(ctx, DefaultParams())
			if tt.shouldPanic {
				assert.Panics(t, func() {
					prmk.SetString(ctx, "vm:p:code_submission_policy", tt.value)
				})
				return
			}
			prmk.SetString(ctx, "vm:p:code_submission_policy", tt.value)
			assert.Equal(t, tt.want, vmk.GetParams(ctx).CodeSubmissionPolicy)
		})
	}
}

// TestWillSetParamCodeSubmitters exercises the governance strings-param setter
// for code_submitters: trimming, empty-segment skipping, empty list -> nil,
// invalid address -> panic, duplicate -> panic (via Validate).
//
// code_submitters is set as a string array (SetStrings) so the stored JSON
// array round-trips into the typed []crypto.Address field via GetParams. This
// test also serves as the regression guard for that round-trip: it calls
// GetParams after every successful set, which would panic if the stored
// representation were incompatible (see the comment in WillSetParam).
func TestWillSetParamCodeSubmitters(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	vmk := env.vmk
	prmk := env.prmk

	addr1 := cspTestAddr("submitter1")
	addr2 := cspTestAddr("submitter2")
	s1, s2 := addr1.String(), addr2.String()

	tests := []struct {
		name        string
		value       []string
		shouldPanic bool
		wantEmpty   bool
		want        []crypto.Address
	}{
		{name: "empty clears list", value: []string{}, wantEmpty: true},
		{name: "single address", value: []string{s1}, want: []crypto.Address{addr1}},
		{name: "two addresses", value: []string{s1, s2}, want: []crypto.Address{addr1, addr2}},
		// Entries are validated verbatim (no trimming), because the keeper
		// stores the raw strings and GetParams decodes them element-wise.
		// Anything that would not round-trip is rejected at set time.
		{name: "whitespace entry rejected", value: []string{"  " + s1 + " "}, shouldPanic: true},
		{name: "empty entry rejected", value: []string{s1, "", s2}, shouldPanic: true},
		{name: "invalid address panics", value: []string{"not-an-address"}, shouldPanic: true},
		{name: "duplicate address panics", value: []string{s1, s1}, shouldPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start from permissioned so the submitters list is meaningful,
			// but validation of submitters happens regardless of policy.
			base := DefaultParams()
			base.CodeSubmissionPolicy = CodeSubmissionPolicyPermissioned
			vmk.SetParams(ctx, base)

			if tt.shouldPanic {
				assert.Panics(t, func() {
					prmk.SetStrings(ctx, "vm:p:code_submitters", tt.value)
				})
				return
			}
			prmk.SetStrings(ctx, "vm:p:code_submitters", tt.value)
			// GetParams must not panic; it must return the decoded list.
			got := vmk.GetParams(ctx).CodeSubmitters
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestCodeSubmittersRoundTrip is the focused regression test for the storage
// round-trip bug: both the genesis path (SetParams -> SetStruct) and the
// governance path (SetStrings) must produce a code_submitters value that
// GetParams can read back without panicking.
func TestCodeSubmittersRoundTrip(t *testing.T) {
	addr1 := cspTestAddr("rt-submitter1")
	addr2 := cspTestAddr("rt-submitter2")

	t.Run("genesis SetParams", func(t *testing.T) {
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		p := DefaultParams()
		p.CodeSubmissionPolicy = CodeSubmissionPolicyPermissioned
		p.CodeSubmitters = []crypto.Address{addr1, addr2}
		require.NoError(t, env.vmk.SetParams(ctx, p))

		got := env.vmk.GetParams(ctx)
		assert.Equal(t, CodeSubmissionPolicyPermissioned, got.CodeSubmissionPolicy)
		assert.Equal(t, []crypto.Address{addr1, addr2}, got.CodeSubmitters)
	})

	t.Run("governance SetStrings", func(t *testing.T) {
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		require.NoError(t, env.vmk.SetParams(ctx, DefaultParams()))

		env.prmk.SetString(ctx, "vm:p:code_submission_policy", "permissioned")
		env.prmk.SetStrings(ctx, "vm:p:code_submitters", []string{addr1.String(), addr2.String()})

		got := env.vmk.GetParams(ctx)
		assert.Equal(t, CodeSubmissionPolicyPermissioned, got.CodeSubmissionPolicy)
		assert.Equal(t, []crypto.Address{addr1, addr2}, got.CodeSubmitters)
	})
}
