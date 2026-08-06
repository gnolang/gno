package gnolang

import (
	"runtime"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tcBuildTagBody(t *testing.T, body string) error {
	t.Helper()
	mp := &std.MemPackage{
		Type:  MPUserProd,
		Name:  "z",
		Path:  "gno.land/p/demo/z",
		Files: []*std.MemFile{{Name: "z.gno", Body: body}},
	}
	_, err := TypeCheckMemPackage(mp, TypeCheckOptions{Mode: TCLatestRelaxed})
	return err
}

// The consensus type-check pins types.Config.GoVersion to go1.18 so the
// accept/reject verdict is a function of the submitted package alone, never of
// the Go toolchain a given validator binary was built with.
//
// go/types honours a per-file //go:build go1.N line by upgrading that file's
// language version above the Config pin, and rejects a file whose version
// exceeds the version the binary was built with. Package bodies are
// attacker-supplied, so without stripping the per-file version a submitter
// could raise the gate on their own file, or a file tagged above one
// validator's toolchain could be accepted on another. GoParseMemPackage blanks
// ast.File.GoVersion on every parsed .gno file so the pin is the sole version
// authority.
func TestTypeCheckMemPackage_BuildTagCannotRaisePin(t *testing.T) {
	t.Parallel()

	// Guard: prove the pin is actually in effect in this build, so a failure
	// below cannot be misread as "the pin is simply missing".
	require.ErrorContains(t, tcBuildTagBody(t, "package z\nfunc F() { for range 10 {} }\n"),
		"go1.22", "precondition: the pin must reject range-over-int without a build tag")

	t.Run("build tag must not raise the pinned version", func(t *testing.T) {
		t.Parallel()

		// Identical package, plus one comment line. Must stay rejected.
		err := tcBuildTagBody(t, "//go:build go1.22\n\npackage z\nfunc F() { for range 10 {} }\n")
		assert.Error(t, err,
			"a //go:build line must not raise the pinned GoVersion: the verdict "+
				"for a submitted package must not be settable by the submitter")

		err = tcBuildTagBody(t, "//go:build go1.21\n\npackage z\nfunc F() int { return min(1, 2) }\n")
		assert.Error(t, err,
			"a //go:build line must not unlock go1.21 builtins the VM cannot run")

		err = tcBuildTagBody(t, "//go:build go1.23\n\npackage z\n"+
			"func F(p func(func(int) bool)) { for range p {} }\n")
		assert.Error(t, err,
			"a //go:build line must not unlock range-over-func")
	})

	t.Run("verdict must not depend on the building toolchain", func(t *testing.T) {
		t.Parallel()

		// go/types rejects a file version newer than the toolchain that built
		// the binary ("file requires newer Go version goX (application built
		// with goY)"). Y is the builder's Go version, so the same package gets
		// opposite verdicts on two honest validators running binaries built
		// with different Go releases. Body is valid go1.18 code; only the build
		// tag varies.
		const body = "//go:build go1.99\n\npackage z\nfunc F() int { return 1 }\n"
		err := tcBuildTagBody(t, body)
		assert.NoError(t, err,
			"verdict must not reference the building toolchain (runtime %s); "+
				"got: %v", runtime.Version(), err)
	})
}

// buildTagImportGetter serves one dependency by path so the import graph can
// carry the build tag instead of the root package.
type buildTagImportGetter map[string]*std.MemPackage

func (g buildTagImportGetter) GetMemPackage(path string) *std.MemPackage {
	return g[path]
}

func TestTypeCheckMemPackage_BuildTagOnImport(t *testing.T) {
	t.Parallel()

	// The blanking runs on every file GoParseMemPackage returns, and the
	// importer recurses through it, so a tag in a dependency is the same
	// vector as one in the package under check. Nothing else asserts that.
	dep := &std.MemPackage{
		Type: MPUserProd,
		Name: "dep",
		Path: "gno.land/p/demo/dep",
		Files: []*std.MemFile{{Name: "dep.gno", Body: "//go:build go1.22\n\n" +
			"package dep\nfunc G() { for range 10 {} }\n"}},
	}
	root := &std.MemPackage{
		Type: MPUserProd,
		Name: "z",
		Path: "gno.land/p/demo/z",
		Files: []*std.MemFile{{Name: "z.gno", Body: "package z\n" +
			"import \"gno.land/p/demo/dep\"\nfunc F() { dep.G() }\n"}},
	}

	getter := buildTagImportGetter{dep.Path: dep}
	_, err := TypeCheckMemPackage(root, TypeCheckOptions{
		Getter:     getter,
		TestGetter: getter,
		Mode:       TCLatestRelaxed,
	})
	assert.Error(t, err,
		"a //go:build line in an imported package must not raise the pinned "+
			"GoVersion for that import")
}
