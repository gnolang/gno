package vm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Under the "inert" policy AddPackage parks without type-checking, so
// EnablePackage is the only place a user package meets the reserved-name
// gate. Live it must refuse; replaying history it must not.
func TestVMKeeperEnableAppliesReservedNameGate(t *testing.T) {
	approver := crypto.AddressFromPreimage([]byte("oracle"))
	submitter := crypto.AddressFromPreimage([]byte("submitter"))
	env, ctx := inertEnv(t, approver, submitter)

	const pkgPath = "gno.land/r/test/gated"
	files := []*std.MemFile{
		{Name: "gated.gno", Body: "package gated\n\nfunc L(cur realm, len int) int { return len }\n"},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(submitter, pkgPath, files)))

	err := env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath))
	requireShadowRefusal(t, err)

	replayCtx := ctx.WithValue(auth.GenesisReplayKey{}, true)
	require.NoError(t, env.vmk.EnablePackage(replayCtx, approvalFor(t, env, ctx, approver, pkgPath)))
}

// The refusal text sits in the error's trace, not its message.
func requireShadowRefusal(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Contains(t, fmt.Sprintf("%+v", err), "builtin identifiers cannot be shadowed: len")
}

// MsgRun meets the same gate, with the same replay exemption.
func TestVMKeeperRunAppliesReservedNameGate(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))

	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/test")},
		{Name: "script.gno", Body: "package main\n\nfunc f(len int) int { return len }\n\nfunc main() { println(f(3)) }\n"},
	}
	msg := NewMsgRun(addr, std.MustParseCoins(""), files)

	_, err := env.vmk.Run(ctx, msg)
	requireShadowRefusal(t, err)

	res, err := env.vmk.Run(ctx.WithValue(auth.GenesisReplayKey{}, true), NewMsgRun(addr, std.MustParseCoins(""), files))
	require.NoError(t, err)
	assert.Equal(t, "3\n", res)
}
