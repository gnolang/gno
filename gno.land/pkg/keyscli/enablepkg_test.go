package keyscli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnablePkgHashMatchesWhatTheChainWillCheck is the property the -pkgdir
// flag exists for: the hash an approver computes from a local copy has to be
// the one EnablePackage computes from the parked blob, or every hand-sent
// approval is refused.
//
// The two differ in one file. AddPackage rewrites gnomod.toml at submit with
// the creator, height and declared deposit, so the stored copy is not the
// submitted one -- which is why PackageContentHash excludes it. This test
// stages that difference rather than assuming it does not matter.
func TestEnablePkgHashMatchesWhatTheChainWillCheck(t *testing.T) {
	const pkgPath = "gno.land/r/test/enablecli"

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"),
		[]byte(gno.GenGnoModLatest(pkgPath)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "enablecli.gno"),
		[]byte("package enablecli\n\nfunc Who(cur realm) string { return \"live\" }\n"), 0o644))

	local, err := gno.ReadMemPackage(dir, pkgPath, gno.MPUserAll)
	require.NoError(t, err)

	// The same package as the chain stores it: gnomod.toml stamped, everything
	// else byte-identical.
	stored := &std.MemPackage{Name: local.Name, Path: local.Path}
	for _, f := range local.Files {
		body := f.Body
		if f.Name == "gnomod.toml" {
			body += "\n[addpkg]\ncreator = \"g1abc\"\nheight = 42\n"
		}
		stored.Files = append(stored.Files, &std.MemFile{Name: f.Name, Body: body})
	}

	assert.Equal(t, vm.PackageContentHash(stored), vm.PackageContentHash(local),
		"a stamped gnomod.toml must not change the hash, or -pkgdir can never match")

	// And a real source change must, or the flag would approve anything.
	changed := &std.MemPackage{Name: local.Name, Path: local.Path}
	for _, f := range local.Files {
		body := f.Body
		if f.Name == "enablecli.gno" {
			body = "package enablecli\n\nfunc Who(cur realm) string { return \"evil\" }\n"
		}
		changed.Files = append(changed.Files, &std.MemFile{Name: f.Name, Body: body})
	}
	assert.NotEqual(t, vm.PackageContentHash(local), vm.PackageContentHash(changed),
		"a source change must change the hash")
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
