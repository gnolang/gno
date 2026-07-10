package fork

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePkg creates a minimal gno package on disk: gnomod.toml + a
// single .gno file with the right package decl.
func writePkg(t *testing.T, root, pkgPath, body string) string {
	t.Helper()
	dir := filepath.Join(root, pkgPath)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"),
		[]byte("module = \"gno.land/"+pkgPath+"\"\ngno = \"0.9\"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.gno"),
		[]byte(body), 0o644))
	return dir
}

func runAddpkg(t *testing.T, args ...string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.jsonl")
	cfg := &addpkgCfg{output: outPath, deployerStr: defaultDeployerAddr}
	io := commands.NewTestIO()
	if err := execAddpkg(t.Context(), cfg, io, args); err != nil {
		return "", err
	}
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	return string(data), nil
}

func TestAddpkg_HappyPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	pkgDir := writePkg(t, root, "r/test/foo",
		"package foo\n\nfunc Hello() string { return \"hi\" }\n")

	out, err := runAddpkg(t, pkgDir)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Len(t, lines, 1)

	var at AnnotatedTx
	require.NoError(t, amino.UnmarshalJSON([]byte(lines[0]), &at))
	require.Len(t, at.Tx.Msgs, 1)
	msg, ok := at.Tx.Msgs[0].(vm.MsgAddPackage)
	require.True(t, ok, "msg is MsgAddPackage")
	assert.Equal(t, "gno.land/r/test/foo", msg.Package.Path)
	assert.Equal(t, defaultDeployerAddr, msg.Creator.String())
	require.NotNil(t, at.Metadata)
	assert.Equal(t, int64(0), at.Metadata.BlockHeight)
	assert.Empty(t, at.Tx.Signatures, "signatures stripped (consumer skips sig verification)")
	assert.Equal(t, "addpkg: gno.land/r/test/foo", at.Reason)
}

func TestAddpkg_MultiplePackages(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := writePkg(t, root, "r/test/foo", "package foo\n")
	b := writePkg(t, root, "r/test/bar", "package bar\n")

	out, err := runAddpkg(t, a, b)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Len(t, lines, 2)
}

func TestAddpkg_RejectsMissingOutput(t *testing.T) {
	t.Parallel()
	cfg := &addpkgCfg{output: "", deployerStr: defaultDeployerAddr}
	err := execAddpkg(t.Context(), cfg, commands.NewTestIO(), []string{"/tmp/dummy"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--output is required")
}

func TestAddpkg_RejectsNoArgs(t *testing.T) {
	t.Parallel()
	cfg := &addpkgCfg{output: filepath.Join(t.TempDir(), "out.jsonl"), deployerStr: defaultDeployerAddr}
	err := execAddpkg(t.Context(), cfg, commands.NewTestIO(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one pkgdir")
}

func TestAddpkg_RejectsBadDeployer(t *testing.T) {
	t.Parallel()
	cfg := &addpkgCfg{output: filepath.Join(t.TempDir(), "out.jsonl"), deployerStr: "not-bech32"}
	err := execAddpkg(t.Context(), cfg, commands.NewTestIO(), []string{"/tmp/dummy"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --deployer")
}
