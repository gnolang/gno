package vm

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

const upgradeNewStdlib = "crypto/bls12381"

var upgradeStdlibDir = filepath.Join("..", "..", "..", "..", "gnovm", "stdlibs")

// storeWithoutStdlib returns a test environment whose store holds every
// stdlib but missing, as genesis on an earlier binary would have written it.
func storeWithoutStdlib(t *testing.T, missing string) testEnv {
	t.Helper()
	env, mcw := setupTestEnvNoStdlibs()
	tx := env.vmk.gnoStore.BeginTransaction(nil, nil, nil, nil)
	for _, lib := range stdlibs.InitOrder() {
		if lib != missing {
			loadStdlibPackage(lib, upgradeStdlibDir, tx)
		}
	}
	tx.Write()
	mcw.MultiWrite()
	return env
}

// A node restarted on a binary that lists a stdlib its store never held (a
// chain upgraded in place) loads it at boot, and realms can then import it.
func TestInitializeLoadsMissingStdlib(t *testing.T) {
	env := storeWithoutStdlib(t, upgradeNewStdlib)
	ms := env.ctx.MultiStore()
	require.Nil(t, env.vmk.gnoStore.GetMemPackage(upgradeNewStdlib))

	vmk2 := NewVMKeeper(env.vmk.baseKey, env.vmk.iavlKey, env.acck, env.bankk, env.prmk)
	vmk2.StdlibDir = upgradeStdlibDir
	mcw := ms.MultiCacheWrap()
	require.NotPanics(t, func() { vmk2.Initialize(log.NewNoopLogger(), mcw) })
	mcw.MultiWrite()
	require.NotNil(t, vmk2.gnoStore.GetMemPackage(upgradeNewStdlib))

	// A realm importing it deploys and runs.
	ctx := vmk2.MakeGnoTransactionStore(env.ctx)
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(10_000_000)))
	const pkgPath = "gno.land/r/test/blsupgrade"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
		{Name: "x.gno", Body: `package blsupgrade

import "crypto/bls12381"

func Size() int { return bls12381.PubKeySize }
`},
	}
	require.NoError(t, vmk2.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))
	vmk2.CommitGnoTransactionStore(ctx)
	res, err := vmk2.QueryEval(env.ctx, pkgPath, "Size()")
	require.NoError(t, err)
	require.Contains(t, res, "48")

	// A later boot finds it stored and writes nothing.
	vmk3 := NewVMKeeper(env.vmk.baseKey, env.vmk.iavlKey, env.acck, env.bankk, env.prmk)
	require.NotPanics(t, func() { vmk3.Initialize(log.NewNoopLogger(), ms.MultiCacheWrap()) })
}

// Without a StdlibDir the keeper names the missing stdlib instead of
// dereferencing a nil package.
func TestInitializeMissingStdlibWithoutDir(t *testing.T) {
	env := storeWithoutStdlib(t, upgradeNewStdlib)
	vmk2 := NewVMKeeper(env.vmk.baseKey, env.vmk.iavlKey, env.acck, env.bankk, env.prmk)
	require.PanicsWithError(t,
		`stdlib "crypto/bls12381" is missing from the store, and the keeper has no StdlibDir to load it from`,
		func() { vmk2.Initialize(log.NewNoopLogger(), env.ctx.MultiStore().MultiCacheWrap()) })
}
