package vm

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// migratedStdlib stands for a stdlib a release adds: a leaf, which no other
// stdlib imports, so a store can hold every stdlib but it.
const migratedStdlib = "encoding/csv"

// storeWithoutStdlib returns a test environment whose store holds every
// stdlib but missing, as genesis on an earlier binary would have written it.
func storeWithoutStdlib(t *testing.T, missing string) testEnv {
	t.Helper()
	env, mcw := setupTestEnvNoStdlibs()
	tx := env.vmk.gnoStore.BeginTransaction(nil, nil, nil, nil)
	for _, lib := range stdlibs.InitOrder() {
		if lib != missing {
			loadStdlibPackage(lib, gnovm.Stdlibs(), tx)
		}
	}
	tx.Write()
	mcw.MultiWrite()
	return env
}

// boot starts a keeper on env's store, as a node restarting on a new binary
// does, and returns it with the cache-wrapped multistore it booted on.
func boot(env testEnv, fsys fs.FS, upgrade bool) (*VMKeeper, func()) {
	vmk := NewVMKeeper(env.vmk.baseKey, env.vmk.iavlKey, env.acck, env.bankk, env.prmk)
	vmk.SetStdlibs(fsys, upgrade)
	mcw := env.ctx.MultiStore().MultiCacheWrap()
	vmk.Initialize(log.NewNoopLogger(), mcw)
	return vmk, mcw.MultiWrite
}

// A node booting at a versioned halt on a binary that lists a stdlib its store
// never held (a chain upgraded in place) adds it, realms can then import it,
// and a later boot finds nothing to do.
func TestInitializeAddsMissingStdlib(t *testing.T) {
	env := storeWithoutStdlib(t, migratedStdlib)
	require.Nil(t, env.vmk.gnoStore.GetMemPackage(migratedStdlib))

	vmk, write := boot(env, gnovm.Stdlibs(), true)
	write()
	require.NotNil(t, vmk.gnoStore.GetMemPackage(migratedStdlib))

	// A realm importing it deploys and runs.
	ctx := vmk.MakeGnoTransactionStore(env.ctx)
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(10_000_000)))
	const pkgPath = "gno.land/r/test/csvuser"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
		{Name: "x.gno", Body: `package csvuser

import (
	"encoding/csv"
	"strings"
)

func Fields() int {
	records, _ := csv.NewReader(strings.NewReader("a,b,c")).ReadAll()
	return len(records[0])
}
`},
	}
	require.NoError(t, vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))
	vmk.CommitGnoTransactionStore(ctx)
	res, err := vmk.QueryEval(env.ctx, pkgPath, "Fields()")
	require.NoError(t, err)
	require.Contains(t, res, "3")

	// A later boot, at no halt, finds the stdlib stored and changes nothing.
	require.NotPanics(t, func() { boot(env, gnovm.Stdlibs(), false) })
}

// Outside a versioned halt the keeper refuses a store that lacks a stdlib,
// naming it, instead of writing it at whatever height the node happens to
// boot.
func TestInitializeRefusesMigrationOutsideHalt(t *testing.T) {
	env := storeWithoutStdlib(t, migratedStdlib)
	require.PanicsWithError(t,
		`stdlib "encoding/csv" in the store differs from this binary's; `+
			`it can only be brought in line by booting at a governance halt that names a version`,
		func() { boot(env, gnovm.Stdlibs(), false) })
}

// A stdlib whose stored files differ from the binary's, other than by a
// missing package, is refused: in-place changes are not supported.
func TestInitializeRefusesChangedStdlib(t *testing.T) {
	env := storeWithoutStdlib(t, "")
	changed := copyFS(t, gnovm.Stdlibs())
	f := changed["strings/builder.gno"]
	changed["strings/builder.gno"] = &fstest.MapFile{Data: append(append([]byte{}, f.Data...), "\n// changed\n"...)}
	require.PanicsWithError(t,
		`stdlib "strings" in the store differs from this binary's; changing an existing stdlib is not supported`,
		func() { boot(env, changed, true) })
}

// copyFS returns a writable copy of fsys.
func copyFS(t *testing.T, fsys fs.FS) fstest.MapFS {
	t.Helper()
	out := fstest.MapFS{}
	require.NoError(t, fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(fsys, path)
		out[path] = &fstest.MapFile{Data: data}
		return err
	}))
	return out
}
