package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	allowedAddr = crypto.AddressFromPreimage([]byte("allowed"))
	otherAddr   = crypto.AddressFromPreimage([]byte("other"))
)

// TestCodePolicyMatrix pins the full (policy x message type) matrix documented
// on checkCodePolicy. Every cell is asserted, because the two rows are
// deliberately different -- add_package is gated only under "permissioned"
// while run is gated always -- and the previous shape of this check (gate both
// whenever policy != permissionless) reads as plausible until you enumerate.
func TestCodePolicyMatrix(t *testing.T) {
	t.Parallel()

	policies := []vm.CodeSubmissionPolicy{
		vm.CodeSubmissionPolicyPermissionless,
		vm.CodeSubmissionPolicyPermissioned,
		vm.CodeSubmissionPolicyInert,
		// A raw empty value and an unrecognised one. Neither can reach the ante
		// in production — GetParams normalises "" via applyLegacyDefaults, and
		// Params.Validate rejects anything else — so reaching this function with
		// either means an upstream invariant broke, and the answer to that is to
		// refuse rather than to guess.
		"",
		"someFuturePolicy",
	}

	// addPkgGated[policy] reports whether an off-list signer is refused for
	// vm/add_package. run is gated under every policy, so it needs no table.
	addPkgGated := map[vm.CodeSubmissionPolicy]bool{
		vm.CodeSubmissionPolicyPermissionless: false,
		vm.CodeSubmissionPolicyPermissioned:   true,
		vm.CodeSubmissionPolicyInert:          false,
		"":                                    true,
		"someFuturePolicy":                    true,
	}

	for _, policy := range policies {
		params := vm.Params{
			CodeSubmissionPolicy: policy,
			CodeSubmitters:       []crypto.Address{allowedAddr},
			RunSubmitters:        []crypto.Address{allowedAddr},
		}

		t.Run("policy="+string(policy), func(t *testing.T) {
			t.Parallel()

			// vm/run: always gated, under every policy value.
			_, abort := codePolicyResult(nil, []crypto.Address{otherAddr}, params)
			assert.True(t, abort,
				"off-list signer must be refused MsgRun under policy %q", policy)
			_, abort = codePolicyResult(nil, []crypto.Address{allowedAddr}, params)
			assert.False(t, abort,
				"listed signer must be allowed MsgRun under policy %q", policy)

			// vm/add_package: gated only under "permissioned".
			_, abort = codePolicyResult([]crypto.Address{otherAddr}, nil, params)
			assert.Equal(t, addPkgGated[policy], abort,
				"add_package gating under policy %q", policy)

			// A tx carrying no code-bearing message is never refused, whatever
			// the lists say.
			_, abort = codePolicyResult(nil, nil, params)
			assert.False(t, abort, "non-code tx must pass under policy %q", policy)
		})
	}
}

// TestEmptyListSemantics pins how each allowlist reads its own empty state.
// The two answers differ, and the difference is the whole design, so it is
// asserted rather than left to the field docs.
func TestEmptyListSemantics(t *testing.T) {
	t.Parallel()

	// run_submitters empty == gate OFF. It is consulted on every MsgRun with no
	// policy switch in front of it, so a fail-closed empty value would disable
	// MsgRun -- and with it GovDAO proposal creation, which is MsgRun-only -- on
	// every chain that upgrades without editing genesis. Under EVERY policy,
	// including the restrictive ones: nothing about code_submission_policy
	// changes what an unset run_submitters means.
	for _, policy := range []vm.CodeSubmissionPolicy{
		vm.CodeSubmissionPolicyPermissionless,
		vm.CodeSubmissionPolicyPermissioned,
		vm.CodeSubmissionPolicyInert,
	} {
		params := vm.Params{CodeSubmissionPolicy: policy}
		_, abort := codePolicyResult(nil, []crypto.Address{otherAddr}, params)
		assert.False(t, abort,
			"empty run_submitters must permit everyone under policy %q", policy)
	}

	// Populating it turns the gate on, for anyone not named.
	params := vm.Params{
		CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissionless,
		RunSubmitters:        []crypto.Address{allowedAddr},
	}
	res, abort := codePolicyResult(nil, []crypto.Address{otherAddr}, params)
	require.True(t, abort, "a populated run_submitters must refuse an off-list signer")
	assert.Contains(t, res.Log, "run_submitters",
		"the error should name the param an operator has to change")

	// code_submitters empty == refuse everyone, the opposite reading, and it is
	// safe precisely because it is unreachable until an operator has explicitly
	// moved the policy to "permissioned". Its empty state is a half-finished
	// opt-in, not a default.
	params = vm.Params{CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissioned}
	res, abort = codePolicyResult([]crypto.Address{allowedAddr}, nil, params)
	require.True(t, abort, "empty code_submitters must refuse everyone")
	assert.Contains(t, res.Log, "code_submitters")
}

// TestCodePolicyAllSignersMustBeListed covers a multi-signer tx: authorization
// is per-signer, so one listed signer must not carry an off-list co-signer.
func TestCodePolicyAllSignersMustBeListed(t *testing.T) {
	t.Parallel()

	params := vm.Params{RunSubmitters: []crypto.Address{allowedAddr}}
	_, abort := codePolicyResult(
		nil, []crypto.Address{allowedAddr, otherAddr}, params)
	assert.True(t, abort, "an off-list co-signer must not be carried by a listed one")
}

// TestCodePolicyRunGatedEvenWhenAddPkgIsNot is the regression pin for the bug
// this shape was written to prevent: a tx carrying both message types under
// "inert", where add_package is deliberately open. The run gate must still
// fire. Collapsing the two rules back into one policy branch passes every
// single-message test above and fails this one.
func TestCodePolicyRunGatedEvenWhenAddPkgIsNot(t *testing.T) {
	t.Parallel()

	params := vm.Params{
		CodeSubmissionPolicy: vm.CodeSubmissionPolicyInert,
		RunSubmitters:        []crypto.Address{allowedAddr},
	}
	_, abort := codePolicyResult([]crypto.Address{otherAddr}, []crypto.Address{otherAddr}, params)
	assert.True(t, abort, "run must stay gated under inert even alongside add_package")

	// And the converse: under inert, an off-list signer sending only
	// add_package is fine even though code_submitters is empty.
	_, abort = codePolicyResult([]crypto.Address{otherAddr}, nil, params)
	assert.False(t, abort, "inert must not gate add_package on code_submitters")
}

// TestCodePolicySignersArePerMessage pins the granularity of the check: the
// signers that must hold code-submission rights are the signers OF THE CODE
// MESSAGES, not every signer of the transaction.
//
// Adapted from the equivalent cases in #5885, which got this right and which
// the first version of this branch got wrong: it authorized from
// tx.GetSigners(), so bundling a code message with any message signed by a
// third party refused the whole tx and reported the third party as
// "not authorized to send MsgRun" when they had sent no such thing.
//
// The scan is exercised through txCodeMsgSigners on real messages rather than
// by passing slices directly, so a message kind dropped from the scan fails
// here too.
func TestCodePolicySignersArePerMessage(t *testing.T) {
	t.Parallel()

	third := crypto.AddressFromPreimage([]byte("third"))
	pkg := &std.MemPackage{
		Name:  "a",
		Files: []*std.MemFile{{Name: "a.gno", Body: "package a\n"}},
	}

	tests := []struct {
		name      string
		params    vm.Params
		msgs      []std.Msg
		wantAbort bool
	}{
		{
			// The bystander only moves coins. Requiring them to hold run
			// rights would refuse a legitimate bundle.
			name:   "bank send by an unlisted bystander alongside a listed run",
			params: vm.Params{RunSubmitters: []crypto.Address{allowedAddr}},
			msgs: []std.Msg{
				bank.MsgSend{FromAddress: otherAddr, ToAddress: allowedAddr},
				vm.MsgRun{Caller: allowedAddr, Package: pkg},
			},
			wantAbort: false,
		},
		{
			name: "bank send by an unlisted bystander alongside a listed addpkg",
			params: vm.Params{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissioned,
				CodeSubmitters:       []crypto.Address{allowedAddr},
			},
			msgs: []std.Msg{
				bank.MsgSend{FromAddress: otherAddr, ToAddress: allowedAddr},
				vm.MsgAddPackage{Creator: allowedAddr, Package: pkg},
			},
			wantAbort: false,
		},
		{
			// But an unlisted signer of a CODE message is still refused, even
			// when another code message in the same tx is fine.
			name: "one listed and one unlisted addpkg",
			params: vm.Params{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissioned,
				CodeSubmitters:       []crypto.Address{allowedAddr},
			},
			msgs: []std.Msg{
				vm.MsgAddPackage{Creator: allowedAddr, Package: pkg},
				vm.MsgAddPackage{Creator: otherAddr, Package: pkg},
			},
			wantAbort: true,
		},
		{
			// The two lists are independent, so a deploy-listed address is not
			// thereby allowed to run. This is the case #5885 asserted the other
			// way round, because there run was gated on code_submitters.
			name: "addpkg-listed signer is not thereby run-listed",
			params: vm.Params{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissioned,
				CodeSubmitters:       []crypto.Address{allowedAddr, third},
				RunSubmitters:        []crypto.Address{allowedAddr},
			},
			msgs: []std.Msg{
				vm.MsgAddPackage{Creator: allowedAddr, Package: pkg},
				vm.MsgRun{Caller: third, Package: pkg},
			},
			wantAbort: true,
		},
		{
			name: "both lists satisfied by their own signers",
			params: vm.Params{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissioned,
				CodeSubmitters:       []crypto.Address{allowedAddr},
				RunSubmitters:        []crypto.Address{third},
			},
			msgs: []std.Msg{
				vm.MsgAddPackage{Creator: allowedAddr, Package: pkg},
				vm.MsgRun{Caller: third, Package: pkg},
			},
			wantAbort: false,
		},
		{
			// MsgCall names a package but carries no source, so an unlisted
			// caller must not be gated even under permissioned.
			name: "MsgCall by an unlisted caller is never gated",
			params: vm.Params{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissioned,
				CodeSubmitters:       []crypto.Address{allowedAddr},
			},
			msgs:      []std.Msg{vm.MsgCall{Caller: otherAddr, PkgPath: "gno.land/r/a", Func: "F"}},
			wantAbort: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			addPkgSigners, runSigners := txCodeMsgSigners(std.Tx{Msgs: tt.msgs})
			_, abort := codePolicyResult(addPkgSigners, runSigners, tt.params)
			assert.Equal(t, tt.wantAbort, abort)
		})
	}
}
