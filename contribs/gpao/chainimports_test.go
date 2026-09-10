package main

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// The composable case "inert" exists to allow: a package importing another
// package that lives only on chain.
//
// It was broken in a way no existing test could see. The dependency's source
// was fetched over RPC and handed to AddMemPackage, which stores source and
// nothing else -- but the preprocessor resolves an import to a PackageValue,
// and the only thing that builds one is the store's package getter, which knew
// stdlibs and examples/ and nothing about the chain. So the typecheck resolved
// the import over RPC, preprocess then panicked with "unknown import path", and
// gpao recorded a rejection of a package the validator would have enabled --
// terminally, since a rejection is keyed on the bytes.
//
// Every other fixture in this package imports only stdlibs or examples/
// residents, which is precisely why it shipped. These import from the chain.
func TestVerifierResolvesChainOnlyImport(t *testing.T) {
	dep := chainPackage("gno.land/p/test/dep", "package dep\n\nfunc Add(a, b int) int { return a + b }\n")
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport \"gno.land/p/test/dep\"\n\nfunc Use() int { return dep.Add(1, 2) }\n")

	v := newRPCVerifier(t, dep)
	require.NoError(t, v.verifyPackage(mpkg),
		"a package importing an on-chain-only dependency must verify")
}

// Same property with a realistic dependency: one that ships tests.
//
// vm/qfile lists every stored file, so a dependency rebuilt from it carries its
// _test.gno and its flattened _filetest.gno. Stamping that MPUserProd -- as the
// getter did -- makes AddMemPackage reject it ("unexpected file given type
// MPUserProd"), which surfaces as a rejection of the IMPORTING package for
// files that are not even its own.
func TestVerifierResolvesChainOnlyImportWithTestFiles(t *testing.T) {
	dep := chainPackage("gno.land/p/test/tested", "package tested\n\nfunc Add(a, b int) int { return a + b }\n")
	dep.Files = append(dep.Files,
		&std.MemFile{Name: "tested_test.gno", Body: "package tested\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(1, 2) != 3 {\n\t\tt.Fatal(\"bad\")\n\t}\n}\n"},
		&std.MemFile{Name: "z_0_filetest.gno", Body: "package main\n\nfunc main() {\n\tprintln(\"ok\")\n}\n\n// Output:\n// ok\n"},
	)
	dep.Sort()
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport \"gno.land/p/test/tested\"\n\nfunc Use() int { return tested.Add(1, 2) }\n")

	v := newRPCVerifier(t, dep)
	require.NoError(t, v.verifyPackage(mpkg),
		"a dependency's own test files are not the importing package's problem")
}

// A diamond: user imports low and mid, and mid also imports low. So a
// dependency is reached both directly and through another chain-only package,
// and low must be resolved before mid can be.
//
// Ordering is the preprocessor's business, not the oracle's -- that is the
// point of resolving through the store's package getter rather than
// materializing the graph here, and this fixture is what says so. An earlier
// revision seeded the graph itself and had to get the order right; reversing
// its breadth-first walk failed exactly this shape.
func TestVerifierResolvesTransitiveChainOnlyImports(t *testing.T) {
	low := chainPackage("gno.land/p/test/low", "package low\n\nfunc One() int { return 1 }\n")
	mid := chainPackage("gno.land/p/test/mid",
		"package mid\n\nimport \"gno.land/p/test/low\"\n\nfunc Two() int { return low.One() + 1 }\n")
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport (\n\t\"gno.land/p/test/low\"\n\t\"gno.land/p/test/mid\"\n)\n\nfunc Use() int { return low.One() + mid.Two() }\n")

	v := newRPCVerifier(t, mid, low)
	require.NoError(t, v.verifyPackage(mpkg),
		"a transitive on-chain-only dependency must resolve too")
}

// A chain that does not deploy under gno.land.
//
// This is what a hardcoded "gno.land/" prefix got wrong: the typecheck routes on
// gno.IsUserlib, which accepts any domain, so it resolved the import over RPC
// and passed -- and then preprocess refused the same path and the submitter was
// told their code was bad. The getter now routes on that same predicate.
func TestVerifierResolvesImportOnAnotherDomain(t *testing.T) {
	dep := chainPackage("example.com/p/test/dep", "package dep\n\nfunc Add(a, b int) int { return a + b }\n")
	mpkg := chainPackage("example.com/p/test/user",
		"package user\n\nimport \"example.com/p/test/dep\"\n\nfunc Use() int { return dep.Add(1, 2) }\n")

	v := newRPCVerifier(t, dep)
	require.NoError(t, v.verifyPackage(mpkg),
		"a chain not called gno.land must verify its own packages")
}

// What prepare leaves for the budget: nothing.
//
// #6116 moved the closure FETCH before the budget, so the compile the validator
// pays for is the only thing measured. Building those dependencies belongs on
// the same side of that line: at enable time the validator's store already
// holds every active package's value, so it never compiles a dependency, and
// charging a candidate for that is charging it for the oracle's setup.
//
// The chain is taken away between the two calls, which is what tells the two
// apart -- a fetch would be served from the RPC cache either way, so only a
// build that already happened can survive this.
func TestPrepareBuildsDependenciesBeforeTheBudget(t *testing.T) {
	dep := chainPackage("gno.land/p/test/dep", "package dep\n\nfunc Add(a, b int) int { return a + b }\n")
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport \"gno.land/p/test/dep\"\n\nfunc Use() int { return dep.Add(1, 2) }\n")

	v := newRPCVerifier(t, dep)
	require.NoError(t, v.prepare(mpkg))

	v.rpc.cache = map[string]*std.MemPackage{}
	v.rpc.qfile = func(string) ([]byte, error) { return nil, errors.New("node is gone") }

	require.NoError(t, v.preprocess(mpkg),
		"preprocess must need neither the network nor a dependency compile")
}

// chainPackage is a submission as the chain stores one: MPUserAll, with a
// gnomod.toml, sorted.
func chainPackage(pkgPath, body string) *std.MemPackage {
	name := path.Base(pkgPath)
	mpkg := &std.MemPackage{
		Name: name,
		Path: pkgPath,
		Type: gno.MPUserAll,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
			{Name: name + ".gno", Body: body},
		},
	}
	mpkg.Sort()
	return mpkg
}

// newRPCVerifier builds a verifier whose chain serves pkgs and nothing else, so
// an import resolves only if it went to the chain: newTestVerifier points at a
// closed port, and the fake getter replaces it.
func newRPCVerifier(t *testing.T, pkgs ...*std.MemPackage) *verifier {
	t.Helper()
	v := newTestVerifier(t)
	v.rpc = &rpcGetter{cache: make(map[string]*std.MemPackage), qfile: fakeQFile(pkgs...)}
	return v
}

// fakeQFile answers like vm/qfile: a newline-separated file list for a package
// path, a body for a package path joined with a file name, an error for
// anything else -- which is also what a node returns for a package that is
// still parked.
func fakeQFile(pkgs ...*std.MemPackage) qfileFunc {
	return func(fpath string) ([]byte, error) {
		for _, mpkg := range pkgs {
			if fpath == mpkg.Path {
				names := make([]string, 0, len(mpkg.Files))
				for _, mfile := range mpkg.Files {
					names = append(names, mfile.Name)
				}
				return []byte(strings.Join(names, "\n")), nil
			}
			for _, mfile := range mpkg.Files {
				if fpath == path.Join(mpkg.Path, mfile.Name) {
					return []byte(mfile.Body), nil
				}
			}
		}
		return nil, errors.New("package is not available")
	}
}

// A dependency that will not build ends prepare, and does NOT become a verdict
// about the candidate.
//
// vm/qfile serves the normal package keyspace and a parked package lives in the
// inert one, invisible to that resolver -- so everything reachable through the
// getter is active on chain, and has already survived every panic buildOne
// catches on a validator. One here therefore says something about this oracle's
// tree. Letting the verification trip over it instead would settle a
// submitter's content hash on the operator's checkout, terminally.
//
// TestPrepareRefusesABrokenStdlibWhereverItIsImported is the same property for
// the other kind of import.
//
// prepare returning non-nil is the whole mechanism: the child exits
// exitResolverUnavailable on it, which the parent classifies as unavailability
// and records pending, uncounted, for a restart or a resubmission to retry.
func TestPrepareRefusesADependencyThatWillNotBuild(t *testing.T) {
	dep := chainPackage("gno.land/p/test/broken", "package broken\n\nfunc Add(a, b int) int { return a + }\n")
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport \"gno.land/p/test/broken\"\n\nfunc Use() int { return broken.Add(1, 2) }\n")

	v := newRPCVerifier(t, dep)
	err := v.prepare(mpkg)

	require.Error(t, err, "a dependency that will not build leaves no verdict to give")
	assert.Contains(t, err.Error(), "could not prebuild")
	assert.Contains(t, err.Error(), "gno.land/p/test/broken")
	assert.Contains(t, err.Error(), "gno.land/p/test/user",
		"the operator has to be told which candidate reached it")
}

// A transport fault outranks a build failure it caused.
//
// The walk fetches before it builds, so a node that dies partway leaves the
// getter unable to resolve a transitive import and the package importing it
// unbuildable. Both leave the candidate pending, so what is at stake is only
// which reason the operator and the status line get: "the node stopped
// answering" is the diagnosis, "a dependency would not compile" is the symptom
// of it, and reporting the symptom sends the operator to their gno checkout
// over a network fault.
func TestPrepareReportsTheTransportFaultUnderABuildFailure(t *testing.T) {
	mid := chainPackage("gno.land/p/test/mid",
		"package mid\n\nimport \"gno.land/p/test/low\"\n\nfunc Two() int { return low.One() + 1 }\n")
	low := chainPackage("gno.land/p/test/low", "package low\n\nfunc One() int { return 1 }\n")
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport \"gno.land/p/test/mid\"\n\nfunc Use() int { return mid.Two() }\n")

	// Serve mid, then go down: low is never fetched, so building mid panics on
	// an import the getter cannot resolve. Wrapped in errResolverUnavailable
	// like newRPCGetter's own qfile wraps an unreachable node, which is what
	// separates a transport fault from the node answering "nothing is there".
	faulting := func() *verifier {
		v := newRPCVerifier(t, mid, low)
		inner := v.rpc.qfile
		v.rpc.qfile = func(fpath string) ([]byte, error) {
			if strings.HasPrefix(fpath, low.Path) {
				return nil, fmt.Errorf("%w: connection refused", errResolverUnavailable)
			}
			return inner(fpath)
		}
		return v
	}

	// There is a build failure to outrank, asserted on the walk itself because
	// prepare returns the fault instead. Without this the test below passes
	// against a buildChainImports that does nothing: prepare's own LoadImports
	// reaches low first and sets transportErr before the walk ever runs.
	buildErr := faulting().buildChainImports(mpkg)
	require.Error(t, buildErr, "the walk must produce the failure being outranked")
	assert.Contains(t, buildErr.Error(), "could not prebuild")
	assert.NotErrorIs(t, buildErr, errResolverUnavailable,
		"the symptom does not carry the diagnosis, which is why the order matters")

	// A fresh verifier, so the choice prepare makes is made on this run's
	// state and not on a fault recorded by the assertions above.
	err := faulting().prepare(mpkg)

	require.Error(t, err)
	assert.ErrorIs(t, err, errResolverUnavailable,
		"the fault under the build failure is what the parent classifies on")
}

// An import the node does not serve is asked for once, in prepare, and not
// again by either stage the budget measures.
//
// That is prepare's whole claim -- it fetches the closure so the stages below
// read the store rather than the network -- and a miss used to escape it: the
// absence was not cached, so LoadImports and hybridGetter each asked again, on
// the clock. The candidate is one the typecheck rejects either way, but against
// a slow node the rejection arrived as an overrun instead, counted toward
// maxOverBudgetAttempts rather than settling the package.
func TestPrepareCachesAnImportTheNodeDoesNotServe(t *testing.T) {
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport \"gno.land/p/test/nothing\"\n\nfunc Use() int { return nothing.X }\n")

	v := newRPCVerifier(t) // a chain that serves nothing at all
	calls := 0
	inner := v.rpc.qfile
	v.rpc.qfile = func(fpath string) ([]byte, error) {
		calls++
		return inner(fpath)
	}

	require.NoError(t, v.prepare(mpkg),
		"an import nothing serves is the typecheck's verdict to give, not prepare's")
	onPrepare := calls
	require.Positive(t, onPrepare, "prepare must have asked, or this proves nothing")

	require.Error(t, v.verifyPackage(mpkg), "the unresolved import must still be refused")
	assert.Equal(t, onPrepare, calls,
		"the typecheck and the preprocess must not reach the node on the budget")
}

// A stdlib that will not build in the operator's tree leaves the candidate
// pending, whether it is imported directly or through a chain dependency.
//
// The asymmetry this pins out: buildChainImports used to skip non-userlib
// paths, and LoadImports is called with abortOnError=false, which recovers the
// disk getter's panic and prints it -- so prepare returned nil and the
// typecheck's own panic became the answer. A candidate importing a broken
// stdlib directly was rejected and marked seen; one importing a /p/ package
// that imports it stayed pending. One level of indirection decided whether the
// operator's checkout settled a stranger's bytes, terminally.
func TestPrepareRefusesABrokenStdlibWhereverItIsImported(t *testing.T) {
	// A tree that will not build, simulated at the seam a corrupt source file
	// panics through. The real fault was NUL bytes in a stdlib .gno file.
	breakStdlib := func(v *verifier, stdlib string) {
		prev := v.prodgs.GetPackageGetter()
		v.prodgs.SetPackageGetter(func(p string, st gno.Store) (*gno.PackageNode, *gno.PackageValue) {
			if p == stdlib {
				panic("broken stdlib tree: " + p)
			}
			return prev(p, st)
		})
	}

	t.Run("imported directly", func(t *testing.T) {
		mpkg := chainPackage("gno.land/p/test/direct",
			"package direct\n\nimport \"strings\"\n\nfunc Use() string { return strings.ToUpper(\"x\") }\n")

		v := newRPCVerifier(t)
		breakStdlib(v, "strings")

		err := v.prepare(mpkg)
		require.Error(t, err, "the operator's tree must not settle the candidate's bytes")
		assert.Contains(t, err.Error(), "could not prebuild")
		assert.Contains(t, err.Error(), "strings")
	})

	t.Run("imported by a chain dependency", func(t *testing.T) {
		dep := chainPackage("gno.land/p/test/viadep",
			"package viadep\n\nimport \"strings\"\n\nfunc Up(s string) string { return strings.ToUpper(s) }\n")
		mpkg := chainPackage("gno.land/p/test/indirect",
			"package indirect\n\nimport \"gno.land/p/test/viadep\"\n\nfunc Use() string { return viadep.Up(\"x\") }\n")

		v := newRPCVerifier(t, dep)
		breakStdlib(v, "strings")

		err := v.prepare(mpkg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not prebuild")
	})

	// And the boundary: a stdlib that is merely absent is a verdict, not a
	// broken tree. GetPackage returns nil for one without panicking, so the
	// walk passes it over and the typecheck reports the unresolved import.
	t.Run("a stdlib that does not exist stays the typecheck's verdict", func(t *testing.T) {
		mpkg := chainPackage("gno.land/p/test/typo",
			"package typo\n\nimport \"strimgs\"\n\nfunc Use() string { return strimgs.ToUpper(\"x\") }\n")

		v := newRPCVerifier(t)
		require.NoError(t, v.prepare(mpkg),
			"an import nothing serves is not the oracle's tree failing")
		assert.Error(t, v.verifyPackage(mpkg),
			"and the typecheck is what refuses it")
	})
}

// Where the chain and the operator's examples/ have drifted, preprocess must
// resolve the chain's copy -- the one the typecheck resolved.
//
// The getter used to try disk first and fall through only on a miss, so a
// dependency present in both came from the operator's checkout while the
// typecheck took the chain's -- injectChainGetter's doc has why that rejects
// packages a validator would enable.
//
// Shadowing a real examples/ resident is the point: an invented path would miss
// on disk and reach the chain either way, which is what the tests above cover.
// The chain's copy carries a function the disk copy does not, so only the
// routing can tell the two apart.
func TestVerifierPrefersTheChainOverADivergentDiskCopy(t *testing.T) {
	const shadowed = "gno.land/p/nt/ownable/v0"
	require.DirExists(t, filepath.Join(gnoenv.RootDir(), "examples", shadowed),
		"the fixture needs a path that really is on disk; examples/ paths move "+
			"(p/demo -> p/nt did), so repoint the constant rather than reading "+
			"this as a routing failure")

	dep := chainPackage(shadowed, "package ownable\n\nfunc OnlyOnChain() int { return 42 }\n")
	mpkg := chainPackage("gno.land/p/test/user",
		"package user\n\nimport \"gno.land/p/nt/ownable/v0\"\n\nfunc Use() int { return ownable.OnlyOnChain() }\n")

	v := newRPCVerifier(t, dep)
	require.NoError(t, v.prepare(mpkg))
	require.NoError(t, v.verifyPackage(mpkg),
		"preprocess must resolve the dependency the typecheck resolved")
}
