package keyscli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnablePkgHashMatchesWhatTheChainWillCheck is the property the -pkgdir
// flag exists for: the hash an approver computes from a local copy has to be
// the digest AddPackage records for the submission, or every hand-sent
// approval is refused.
//
// AddPackage records the digest of the package exactly as addpkg sends it, so
// both commands run here against one directory and the approval is compared
// with the submission they print. A test file rides along because the two
// commands have to agree on the file set, not only on the bytes.
func TestEnablePkgHashMatchesWhatTheChainWillCheck(t *testing.T) {
	const (
		pkgPath = "gno.land/r/test/enablecli"
		keyName = "approver"
		// Any valid mnemonic: nothing here is signed.
		mnemonic = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
	)

	dir := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	write("gnomod.toml", gno.GenGnoModLatest(pkgPath))
	write("enablecli.gno", "package enablecli\n\nfunc Who(cur realm) string { return \"live\" }\n")
	write("enablecli_test.gno", "package enablecli\n\nimport \"testing\"\n\nfunc TestWho(t *testing.T) {}\n")

	home := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	_, err = kb.CreateAccount(keyName, mnemonic, "", "", 0, 0)
	require.NoError(t, err)
	txCfg := &client.MakeTxCfg{
		RootCfg:   &client.BaseCfg{BaseOptions: client.BaseOptions{Home: home}},
		GasWanted: 1,
		GasFee:    "1ugnot",
	}

	// printed runs a command without -broadcast and returns the one message of
	// the transaction it prints.
	printed := func(t *testing.T, run func(commands.IO) error) std.Msg {
		t.Helper()
		var out bytes.Buffer
		cio := commands.NewTestIO()
		cio.SetOut(commands.WriteNopCloser(&out))
		require.NoError(t, run(cio))
		var tx std.Tx
		require.NoError(t, amino.UnmarshalJSON(out.Bytes(), &tx))
		require.Len(t, tx.Msgs, 1)
		return tx.Msgs[0]
	}
	approve := func(t *testing.T) vm.MsgEnablePackage {
		t.Helper()
		return printed(t, func(cio commands.IO) error {
			return execMakeEnablePkg(&MakeEnablePkgCfg{RootCfg: txCfg, PkgPath: pkgPath, PkgDir: dir}, []string{keyName}, cio)
		}).(vm.MsgEnablePackage)
	}

	add := printed(t, func(cio commands.IO) error {
		return execMakeAddPkg(&MakeAddPkgCfg{RootCfg: txCfg, PkgPath: pkgPath, PkgDir: dir}, []string{keyName}, cio)
	}).(vm.MsgAddPackage)
	require.NotNil(t, add.Package.GetFile("enablecli_test.gno"), "premise: addpkg submits test files")

	approval := approve(t)
	assert.Equal(t, vm.PackageContentHash(add.Package), approval.PkgHash,
		"-pkgdir names a digest other than the one AddPackage records for what addpkg submits")

	// And a real source change must move it, or the flag would approve anything.
	write("enablecli.gno", "package enablecli\n\nfunc Who(cur realm) string { return \"evil\" }\n")
	assert.NotEqual(t, approval.PkgHash, approve(t).PkgHash, "a source change must change the hash")
}

// TestEnablePkgRequiresASource pins that the command refuses to build an
// approval that names no source. Defaulting to "whatever is parked" would undo
// the guard the hash exists for.
func TestEnablePkgRequiresASource(t *testing.T) {
	cfg := &MakeEnablePkgCfg{PkgPath: "gno.land/r/test/x"}
	err := execMakeEnablePkg(cfg, []string{"key"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "has to name the source")

	cfg = &MakeEnablePkgCfg{PkgPath: "gno.land/r/test/x", PkgDir: "/tmp/x", PkgHash: "abc"}
	err = execMakeEnablePkg(cfg, []string{"key"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one of")
}
