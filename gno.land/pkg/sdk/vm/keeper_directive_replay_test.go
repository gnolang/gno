package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tmerrors "github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// A build constraint master admitted: #5978's txtar deployed this exact
	// tag, so a chain's history can hold it.
	tagFuture = "//go:build go1.99\n\n"
	// #5978's pinned vector: the tag must not unlock range-over-int.
	tagGo122 = "//go:build go1.22\n\n"
	loop122  = "\tfor range 10 {\n\t}\n"
)

// directiveFiles is a one-file package at gno.land/r/test/<name>.
func directiveFiles(name, tag, body string) (string, []*std.MemFile) {
	pkgPath := "gno.land/r/test/" + name
	return pkgPath, sortMemFiles([]*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: name + ".gno", Body: tag + "package " + name + "\n\nfunc F() {\n" + body + "}\n"},
	})
}

// directiveScript is a MsgRun from addr whose script carries tag.
func directiveScript(addr crypto.Address, tag, body string) MsgRun {
	return NewMsgRun(addr, std.Coins{}, []*std.MemFile{{
		Name: "script.gno",
		Body: tag + "package main\n\nfunc main() {\n" + body + "}\n",
	}})
}

// fundedReplayEnv returns a permissionless chain with one funded account, its
// live context and the same context marked as a genesis replay.
func fundedReplayEnv(t *testing.T, seed string) (env testEnv, live, replay sdk.Context, addr crypto.Address) {
	t.Helper()
	env = setupTestEnv()
	live = env.vmk.MakeGnoTransactionStore(env.ctx)
	addr = crypto.AddressFromPreimage([]byte(seed))
	env.acck.SetAccount(live, env.acck.NewAccountWithAddress(live, addr))
	require.NoError(t, env.bankk.SetCoins(live, addr, initialBalance))
	return env, live, live.WithValue(auth.GenesisReplayKey{}, true), addr
}

func requireDirectiveRefusal(t *testing.T, err error, msg string) {
	t.Helper()
	require.Error(t, err)
	require.Contains(t, fmt.Sprintf("%+v", err), "directives are not supported", msg)
}

// requireGoVersionPin asserts the type-check, not the directive rule, refused
// go1.22 syntax: with the rule waived on replay, a //go:build line reaches the
// type-check again and #5978's pin must still hold there. This is the coverage
// addpkg_build_tag_goversion.txtar gave before the directive gate pre-empted it
// on live submissions.
func requireGoVersionPin(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	assert.Equal(t, TypeCheckError{}, tmerrors.Cause(err), "refused by the type-check, not earlier")
	full := fmt.Sprintf("%+v", err)
	assert.Contains(t, full, "go1.22", "the type-check must refuse go1.22 syntax")
	assert.NotContains(t, full, "directives are not supported", "the directive rule is waived on replay")
}

// TestDirectiveRuleWaivedOnGenesisReplay pins that the directive rule
// (adr/pr6078_forbid_directives.md) gates live admission only. A genesis replay
// re-delivers deploys the source chain already accepted, under the binary that
// accepted them; refusing one stops InitChain, or under
// -skip-failing-genesis-txs brings the fork up diverged. The type-check's
// AlreadyOnChain waiver exists for the same reason.
func TestDirectiveRuleWaivedOnGenesisReplay(t *testing.T) {
	t.Run("AddPackage", func(t *testing.T) {
		env, live, replay, addr := fundedReplayEnv(t, "zdirreplay")
		pkgPath, files := directiveFiles("zdirreplay", tagFuture, "")

		requireDirectiveRefusal(t, env.vmk.AddPackage(live, NewMsgAddPackage(addr, pkgPath, files)),
			"a live deploy must still be refused")
		require.Nil(t, env.vmk.getGnoTransactionStore(live).GetPackage(pkgPath, false))

		require.NoError(t, env.vmk.AddPackage(replay, NewMsgAddPackage(addr, pkgPath, files)),
			"replay must reproduce a deploy the source chain accepted")
		require.NotNil(t, env.vmk.getGnoTransactionStore(replay).GetPackage(pkgPath, false))

		pinPath, pinFiles := directiveFiles("zdirreplaypin", tagGo122, loop122)
		requireGoVersionPin(t, env.vmk.AddPackage(replay, NewMsgAddPackage(addr, pinPath, pinFiles)))
	})

	t.Run("Run", func(t *testing.T) {
		env, live, replay, addr := fundedReplayEnv(t, "zdirreplayrun")
		msg := directiveScript(addr, tagFuture, "")

		_, err := env.vmk.Run(live, msg)
		requireDirectiveRefusal(t, err, "a live run must still be refused")

		_, err = env.vmk.Run(replay, msg)
		require.NoError(t, err, "replay must reproduce a run the source chain accepted")

		// TCLatestRelaxed mode, the other type-check entry point.
		_, err = env.vmk.Run(replay, directiveScript(addr, tagGo122, loop122))
		requireGoVersionPin(t, err)
	})

	t.Run("EnablePackage", func(t *testing.T) {
		approver := crypto.AddressFromPreimage([]byte("zdiroracle"))
		owner := crypto.AddressFromPreimage([]byte("zdirowner"))
		env, ctx := inertEnv(t, approver, owner)
		replay := ctx.WithValue(auth.GenesisReplayKey{}, true)
		pkgPath, files := directiveFiles("zdirreplayenable", tagFuture, "")

		// Parked as a binary from before the rule would have parked it:
		// replay waives the rule, and the inert policy parks.
		require.NoError(t, env.vmk.AddPackage(replay, NewMsgAddPackage(owner, pkgPath, files)))
		require.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false), "must be parked, not live")
		approval := approvalFor(t, env, ctx, approver, pkgPath)

		// A live enable re-validates the parked blob under the current rules,
		// as validateParkedBlob documents, so it is refused...
		requireDirectiveRefusal(t, env.vmk.EnablePackage(ctx, approval), "")
		require.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

		// ...while a replayed one reproduces what the source chain did.
		require.NoError(t, env.vmk.EnablePackage(replay, approval),
			"replay must reproduce an enable the source chain accepted")
		assert.NotNil(t, env.vmk.getGnoTransactionStore(replay).GetPackage(pkgPath, false))
	})
}
