package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

// setupCodePolicyEnv builds a minimal VMKeeper wired to a params keeper — just
// enough for checkCodeSubmissionPolicy, which only reads vm params and the tx
// messages. Stdlibs are intentionally NOT loaded (not needed for the ante
// check), keeping the test fast.
func setupCodePolicyEnv(t *testing.T) (sdk.Context, *vm.VMKeeper) {
	t.Helper()

	db := memdb.NewMemDB()
	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseCapKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlCapKey, iavl.StoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id", Height: 1}, log.NewNoopLogger())

	prmk := params.NewParamsKeeper(iavlCapKey)
	acck := auth.NewAccountKeeper(iavlCapKey, prmk.ForModule(auth.ModuleName), std.ProtoBaseAccount, std.ProtoBaseSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName))
	vmk := vm.NewVMKeeper(baseCapKey, iavlCapKey, acck, bankk, prmk)
	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)
	prmk.Register(vm.ModuleName, vmk)

	return ctx, vmk
}

// addPkgMsg builds a minimal MsgAddPackage signed by creator.
func addPkgMsg(creator crypto.Address) std.Msg {
	files := []*std.MemFile{{Name: "a.gno", Body: "package a\n"}}
	return vm.NewMsgAddPackage(creator, "gno.land/r/test/a", files)
}

// runMsg builds a minimal MsgRun signed by caller.
func runMsg(caller crypto.Address) std.Msg {
	files := []*std.MemFile{{Name: "main.gno", Body: "package main\nfunc main() {}\n"}}
	return vm.NewMsgRun(caller, nil, files)
}

// callMsg builds a MsgCall (vm route, type "exec") — not gated by policy.
func callMsg(caller crypto.Address) std.Msg {
	return vm.NewMsgCall(caller, nil, "gno.land/r/test/a", "Foo", nil)
}

// sendMsg builds a bank MsgSend (non-vm route) — not gated by policy.
func sendMsg(from, to crypto.Address) std.Msg {
	return bank.NewMsgSend(from, to, std.Coins{std.NewCoin("ugnot", 1)})
}

func TestCheckCodeSubmissionPolicy(t *testing.T) {
	allowed := crypto.AddressFromPreimage([]byte("allowed"))
	other := crypto.AddressFromPreimage([]byte("other"))
	third := crypto.AddressFromPreimage([]byte("third"))

	tests := []struct {
		name       string
		policy     vm.CodeSubmissionPolicy
		submitters []crypto.Address
		msgs       []std.Msg
		wantAbort  bool
	}{
		{
			name:      "permissionless allows unlisted add_package",
			policy:    vm.CodeSubmissionPolicyPermissionless,
			msgs:      []std.Msg{addPkgMsg(other)},
			wantAbort: false,
		},
		{
			name:      "empty policy treated as permissionless",
			policy:    "", // zero value
			msgs:      []std.Msg{addPkgMsg(other)},
			wantAbort: false,
		},
		{
			name:       "permissioned allows listed add_package",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{addPkgMsg(allowed)},
			wantAbort:  false,
		},
		{
			name:       "permissioned rejects unlisted add_package",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{addPkgMsg(other)},
			wantAbort:  true,
		},
		{
			name:       "permissioned allows listed run",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{runMsg(allowed)},
			wantAbort:  false,
		},
		{
			name:       "permissioned rejects unlisted run",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{runMsg(other)},
			wantAbort:  true,
		},
		{
			name:       "permissioned ignores MsgCall from unlisted",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{callMsg(other)},
			wantAbort:  false,
		},
		{
			name:       "permissioned ignores bank send from unlisted",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{sendMsg(other, allowed)},
			wantAbort:  false,
		},
		{
			name:       "permissioned rejects if any code msg is unlisted",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{addPkgMsg(allowed), addPkgMsg(other)},
			wantAbort:  true,
		},
		{
			name:       "permissioned allows all-listed code msgs",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed, third},
			msgs:       []std.Msg{addPkgMsg(allowed), runMsg(third)},
			wantAbort:  false,
		},
		{
			name:       "permissioned with empty allowlist rejects add_package",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: nil,
			msgs:       []std.Msg{addPkgMsg(allowed)},
			wantAbort:  true,
		},
		{
			name:       "permissioned allows non-code msg mixed with listed code msg",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{sendMsg(other, allowed), addPkgMsg(allowed)},
			wantAbort:  false,
		},
		{
			name:       "permissioned rejects unlisted code msg mixed with non-code msg",
			policy:     vm.CodeSubmissionPolicyPermissioned,
			submitters: []crypto.Address{allowed},
			msgs:       []std.Msg{sendMsg(allowed, other), addPkgMsg(other)},
			wantAbort:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, vmk := setupCodePolicyEnv(t)
			p := vm.DefaultParams()
			p.CodeSubmissionPolicy = tt.policy
			p.CodeSubmitters = tt.submitters
			require.NoError(t, vmk.SetParams(ctx, p))

			tx := std.Tx{Msgs: tt.msgs}
			res, abort := checkCodeSubmissionPolicy(ctx, tx, vmk)

			assert.Equal(t, tt.wantAbort, abort, "abort mismatch")
			if tt.wantAbort {
				assert.False(t, res.IsOK(), "expected non-OK result on abort")
				assert.Contains(t, res.Log, "not authorized to submit code")
			} else {
				assert.True(t, res.IsOK(), "expected OK result, got: %s", res.Log)
			}
		})
	}
}

// TestCheckCodeSubmissionPolicyDefaultParams verifies that a freshly-defaulted
// vm keeper (permissionless) never aborts a code submission.
func TestCheckCodeSubmissionPolicyDefaultParams(t *testing.T) {
	ctx, vmk := setupCodePolicyEnv(t)
	require.NoError(t, vmk.SetParams(ctx, vm.DefaultParams()))

	tx := std.Tx{Msgs: []std.Msg{addPkgMsg(crypto.AddressFromPreimage([]byte("anyone")))}}
	res, abort := checkCodeSubmissionPolicy(ctx, tx, vmk)
	assert.False(t, abort)
	assert.True(t, res.IsOK(), res.Log)
}
