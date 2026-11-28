package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	tmerrors "github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// directiveFiles is a package with a build constraint master admitted: #5978's
// txtar deployed this exact tag, so a chain's history can hold it.
func directiveFiles(pkgPath, name string) []*std.MemFile {
	return []*std.MemFile{
		{Name: "gnomod.toml", Body: `module = "` + pkgPath + `"
gno = "0.9"`},
		{Name: name + ".gno", Body: "//go:build go1.99\n\npackage " + name + "\n\nfunc F() int { return 1 }\n"},
	}
}

// TestDirectiveRuleWaivedOnGenesisReplay pins that the directive rule
// (adr/pr6078_forbid_directives.md) gates live admission only. A genesis replay
// re-delivers deploys the source chain already accepted, under the binary that
// accepted them; refusing one stops InitChain, or under
// -skip-failing-genesis-txs brings the fork up diverged. The type-check's
// AlreadyOnChain waiver exists for the same reason.
func TestDirectiveRuleWaivedOnGenesisReplay(t *testing.T) {
	t.Run("AddPackage", func(t *testing.T) {
		const (
			pkgPath = "gno.land/r/test/zdirreplay"
			name    = "zdirreplay"
		)
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		addr := crypto.AddressFromPreimage([]byte("zdirreplay"))
		env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
		require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))

		err := env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, directiveFiles(pkgPath, name)))
		require.Error(t, err)
		require.Contains(t, fmt.Sprintf("%+v", err), "directives are not supported",
			"a live deploy must still be refused")
		require.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

		replay := ctx.WithValue(auth.GenesisReplayKey{}, true)
		require.NoError(t, env.vmk.AddPackage(replay, NewMsgAddPackage(addr, pkgPath, directiveFiles(pkgPath, name))),
			"replay must reproduce a deploy the source chain accepted")
		require.NotNil(t, env.vmk.getGnoTransactionStore(replay).GetPackage(pkgPath, false))
	})

	t.Run("replay still applies the GoVersion pin", func(t *testing.T) {
		// With the directive rule waived, a replayed //go:build line reaches the
		// type-check again, and #5978's pin must still hold there: the tag must
		// not unlock go1.22 syntax. This is the coverage
		// addpkg_build_tag_goversion.txtar gave before the directive gate
		// pre-empted it on live submissions.
		const (
			pkgPath = "gno.land/r/test/zdirreplaypin"
			name    = "zdirreplaypin"
		)
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		addr := crypto.AddressFromPreimage([]byte("zdirreplaypin"))
		env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
		require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))
		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: `module = "` + pkgPath + `"
gno = "0.9"`},
			{Name: name + ".gno", Body: "//go:build go1.22\n\npackage " + name + "\n\nfunc F() {\n\tfor range 10 {\n\t}\n}\n"},
		}

		err := env.vmk.AddPackage(ctx.WithValue(auth.GenesisReplayKey{}, true), NewMsgAddPackage(addr, pkgPath, files))
		require.Error(t, err)
		assert.Equal(t, TypeCheckError{}, tmerrors.Cause(err), "refused by the type-check, not earlier")
		full := fmt.Sprintf("%+v", err)
		assert.Contains(t, full, "go1.22", "the type-check must refuse go1.22 syntax")
		assert.NotContains(t, full, "directives are not supported", "the directive rule is waived on replay")
	})

	t.Run("Run", func(t *testing.T) {
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		addr := crypto.AddressFromPreimage([]byte("zdirreplayrun"))
		env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
		require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))
		msg := NewMsgRun(addr, std.Coins{}, []*std.MemFile{{
			Name: "script.gno",
			Body: "//go:build go1.99\n\npackage main\n\nfunc main() {}\n",
		}})

		_, err := env.vmk.Run(ctx, msg)
		require.Error(t, err)
		require.Contains(t, fmt.Sprintf("%+v", err), "directives are not supported",
			"a live run must still be refused")

		_, err = env.vmk.Run(ctx.WithValue(auth.GenesisReplayKey{}, true), msg)
		require.NoError(t, err, "replay must reproduce a run the source chain accepted")

		// And #5978's pin still holds at this entry point on replay
		// (TCLatestRelaxed mode): the tag must not unlock go1.22 syntax.
		pinned := NewMsgRun(addr, std.Coins{}, []*std.MemFile{{
			Name: "script.gno",
			Body: "//go:build go1.22\n\npackage main\n\nfunc main() {\n\tfor range 10 {\n\t}\n}\n",
		}})
		_, err = env.vmk.Run(ctx.WithValue(auth.GenesisReplayKey{}, true), pinned)
		require.Error(t, err)
		assert.Equal(t, TypeCheckError{}, tmerrors.Cause(err), "refused by the type-check, not earlier")
		assert.Contains(t, fmt.Sprintf("%+v", err), "go1.22")
	})

	t.Run("EnablePackage", func(t *testing.T) {
		const (
			pkgPath = "gno.land/r/test/zdirreplayenable"
			name    = "zdirreplayenable"
		)
		approver := crypto.AddressFromPreimage([]byte("zdiroracle"))
		owner := crypto.AddressFromPreimage([]byte("zdirowner"))
		env, ctx := inertEnv(t, approver, owner)
		replay := ctx.WithValue(auth.GenesisReplayKey{}, true)

		// Parked as a binary from before the rule would have parked it:
		// replay waives the rule, and the inert policy parks.
		require.NoError(t, env.vmk.AddPackage(replay, NewMsgAddPackage(owner, pkgPath, directiveFiles(pkgPath, name))))
		require.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false), "must be parked, not live")
		approval := approvalFor(t, env, ctx, approver, pkgPath)

		// A live enable re-validates the parked blob under the current rules,
		// as validateParkedBlob documents, so it is refused...
		err := env.vmk.EnablePackage(ctx, approval)
		require.Error(t, err)
		require.Contains(t, fmt.Sprintf("%+v", err), "directives are not supported")
		require.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

		// ...while a replayed one reproduces what the source chain did.
		require.NoError(t, env.vmk.EnablePackage(replay, approval),
			"replay must reproduce an enable the source chain accepted")
		assert.NotNil(t, env.vmk.getGnoTransactionStore(replay).GetPackage(pkgPath, false))
	})
}
