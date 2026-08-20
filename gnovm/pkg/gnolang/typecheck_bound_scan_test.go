package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gnoSrcRoots maps an import-path prefix to the directory holding those packages.
// Shared by the scans below so "which directory is a package, and which of its
// files count" has one definition rather than one per test.
type gnoSrcRoots map[string]string

var (
	stdlibRoots = gnoSrcRoots{"": "../../stdlibs"}
	allGnoRoots = gnoSrcRoots{
		"":          "../../stdlibs",               // stdlib import paths are bare
		"gno.land/": "../../../examples/gno.land/", // examples live under gno.land/
	}
)

// resolver returns a pkgResolver reading source from disk, mirroring what the
// deploy-path resolver hands the guard but reaching the working tree instead of
// the store. Unreadable and unparseable files are skipped, as the guard also
// treats unparseable dependencies as leaves.
//
// prodOnly mirrors MPFProd, which is what the deploy path filters DEPENDENCIES
// through. Pass true when measuring what a user package can name: a _test.gno file
// may declare an exported type in an xxx_test package that nothing can import.
func (roots gnoSrcRoots) resolver(fset *token.FileSet, prodOnly bool) pkgResolver {
	return func(pkgPath string) []*ast.File {
		for prefix, root := range roots {
			if prefix != "" && !strings.HasPrefix(pkgPath, prefix) {
				continue
			}
			dir := filepath.Join(root, strings.TrimPrefix(pkgPath, prefix))
			ents, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			var out []*ast.File
			for _, e := range ents {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".gno") {
					continue
				}
				if prodOnly && (strings.HasSuffix(e.Name(), "_test.gno") ||
					strings.HasSuffix(e.Name(), "_filetest.gno")) {
					continue
				}
				src, err := os.ReadFile(filepath.Join(dir, e.Name()))
				if err != nil {
					continue
				}
				f, err := parser.ParseFile(fset, filepath.Join(pkgPath, e.Name()), src,
					parser.SkipObjectResolution)
				if err != nil {
					continue
				}
				out = append(out, f)
			}
			if len(out) > 0 {
				return out
			}
		}
		return nil
	}
}

// pkgPaths enumerates every directory holding .gno files, as an import path.
func (roots gnoSrcRoots) pkgPaths(t *testing.T) []string {
	t.Helper()
	var out []string
	for prefix, root := range roots {
		require.NoError(t, filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil //nolint:nilerr // unreadable dirs are simply not packages
			}
			ents, err := os.ReadDir(p)
			if err != nil {
				return nil //nolint:nilerr
			}
			for _, e := range ents {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".gno") {
					rel, err := filepath.Rel(root, p)
					require.NoError(t, err)
					out = append(out, prefix+filepath.ToSlash(rel))
					return nil
				}
			}
			return nil
		}))
	}
	return out
}

// shimPath/shimFiles expose the .gnobuiltins shim package to the scans, parsed by
// the same function the production guard is fed by.
const shimPath = "gnobuiltins/gno0p9"

func shimFiles(t *testing.T, fset *token.FileSet) []*ast.File {
	t.Helper()
	//nolint:dogsled // only allgofs is needed; the shim is a single prod file.
	_, gofs, _, _, _, err := GoParseMemPackage(gnoBuiltinsMemPackage(shimPath), fset)
	require.NoError(t, err)
	return gofs
}

// TestHonestTypeExpansionUnderBudget pins what real code actually costs: the
// largest expansion total of any stdlib or example package. That number is the
// whole argument that pricing the validType walk does not tax honest deploys, so
// without this test it would rot silently — a cost() change that over-counted an
// everyday shape would show up as gas on every deploy and nowhere else.
//
// It scans every stdlib and example package INCLUDING test files (the guard runs
// on the whole file set for the package being checked) and asserts a wide margin
// rather than an exact figure, so ordinary growth does not churn the test.
// See also TestLeafExpansionBound, which pins the other half: what an unresolved
// stdlib reference costs, and so how much of this total is over-count.
func TestHonestTypeExpansionUnderBudget(t *testing.T) {
	t.Parallel()

	const (
		// ~2x today's max (431, gno.land/p/nt/commondao/v0): loose enough that
		// ordinary growth does not churn the test, tight enough to catch a cost()
		// change that inflates real code. At typeExpansionGasPerNode this whole
		// budget is ~100k gas, against the ~5e7 GasWanted a real deploy uses, so
		// pricing the walk is not a meaningful tax on honest packages.
		maxHonestTotal = 1_000
		// Guard against the scan silently finding nothing (a moved directory would
		// otherwise make this test vacuously pass).
		minPackages = 100
	)

	fset := token.NewFileSet()
	resolve := allGnoRoots.resolver(fset, false)
	pkgPaths := allGnoRoots.pkgPaths(t)
	require.GreaterOrEqual(t, len(pkgPaths), minPackages,
		"scan found too few packages (%d) — did the source layout move?", len(pkgPaths))

	// Score each package exactly as the guard does. Stdlib references are leaves
	// on the deploy path, so treat them as such here too.
	shared := newExpansionPkgCache()
	userResolve := func(pkgPath string) []*ast.File {
		if IsStdlib(pkgPath) {
			return nil
		}
		return resolve(pkgPath)
	}

	var worstTotal uint64
	var worstPkg string
	nTypes := 0
	for _, pp := range pkgPaths {
		gofs := resolve(pp)
		if len(gofs) == 0 {
			continue
		}
		// Call the production function rather than re-summing here: a test that
		// reimplements the formula it is measuring stops measuring it the moment
		// the formula changes.
		total := typeExpansionCost(pp, gofs, userResolve, shared)
		for _, specs := range newExpansionChecker(pp, gofs, userResolve, shared).declsFor(pp) {
			nTypes += len(specs)
		}
		if total > worstTotal {
			worstTotal, worstPkg = total, pp
		}
	}

	t.Logf("scanned %d packages, %d named types; largest per-package expansion total "+
		"is %d (%s)", len(pkgPaths), nTypes, worstTotal, worstPkg)
	assert.LessOrEqual(t, worstTotal, uint64(maxHonestTotal),
		"package %s expands to %d nodes, over the %d expected of real code. Either "+
			"cost() now over-counts an honest shape, or a real type graph grew far "+
			"beyond anything measured — investigate which before relaxing this test, "+
			"since this measurement is the whole false-rejection argument for the "+
			"per-node gas charge",
		worstPkg, worstTotal, maxHonestTotal)
}

// TestLeafExpansionBound pins the constant that keeps the cost model an
// over-estimate. namedCost cannot resolve a stdlib type — expansionPkgResolver
// deliberately does not fetch stdlib, because go/types serves those imports from
// its own cache without a store read, so resolving them here would add store gas
// a deploy otherwise never pays — nor the .gnobuiltins.gno shim types, which are
// injected after the guards run. Both are therefore scored at leafExpansionBound.
//
// That is only sound if it is an OVER-count, and scoring such a leaf too low
// multiplies: a leaf at the base of a doubling chain of depth d is walked 2^d
// times, so an under-count of k there discounts the whole package k-fold. This
// measures the real maximum a user package can name and fails if the constant no
// longer covers it.
func TestLeafExpansionBound(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	// Unlike the deploy path, resolve stdlib in full: that is the number being
	// measured. Anything still unresolved inside stdlib falls back to
	// leafExpansionBound, so a measured max below the constant also proves no such
	// fallback contributed to it.
	resolve := stdlibRoots.resolver(fset, true)
	pkgPaths := stdlibRoots.pkgPaths(t)
	require.GreaterOrEqual(t, len(pkgPaths), 20,
		"scan found too few stdlib packages (%d) — did the source layout move?", len(pkgPaths))

	// The shim types are named by every package, so they belong in the same bound.
	shim := shimFiles(t, fset)

	shared := newExpansionPkgCache()
	full := func(pkgPath string) []*ast.File {
		if pkgPath == shimPath {
			return shim
		}
		return resolve(pkgPath)
	}

	var worst uint64
	var worstName string
	nTypes := 0
	for _, pp := range append(pkgPaths, shimPath) {
		gofs := full(pp)
		if len(gofs) == 0 {
			continue
		}
		c := newExpansionChecker(pp, gofs, full, shared)
		for name, specs := range c.declsFor(pp) {
			// Only exported names: an unexported stdlib type cannot be named from a
			// user package, so it can never be the leaf that gets scored.
			if !token.IsExported(name) {
				continue
			}
			nTypes++
			for _, d := range specs {
				if v := satAdd(1, c.cost(d.spec.Type, pp, d.imports)); v > worst {
					worst, worstName = v, pp+"."+name
				}
			}
		}
	}

	t.Logf("scanned %d stdlib packages, %d exported named types; largest expansion "+
		"is %d (%s), leafExpansionBound is %d",
		len(pkgPaths), nTypes, worst, worstName, leafExpansionBound)
	assert.LessOrEqual(t, worst, uint64(leafExpansionBound),
		"exported type %s expands to %d nodes, above leafExpansionBound (%d). The "+
			"cost model now UNDER-counts every reference to it, and an under-count at "+
			"the base of a doubling chain discounts the whole package: raise the "+
			"constant to cover this",
		worstName, worst, leafExpansionBound)
}

// TestGnoBuiltinShimExpansion pins gnoBuiltinShimExpansion against the real shim
// source. The two names are scored from a table because .gnobuiltins.gno is
// injected after the guards run, so if the shim ever grows a containment edge —
// an embedded interface in `realm`, a struct field in `address` — the table would
// silently under-count a type that appears in nearly every package.
func TestGnoBuiltinShimExpansion(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	c := newExpansionChecker(shimPath, shimFiles(t, fset), nil, nil)

	require.Len(t, gnoBuiltinShimExpansion, 2)
	for name, want := range gnoBuiltinShimExpansion {
		specs := c.declsFor(shimPath)[name]
		require.Len(t, specs, 1, "shim no longer declares %q", name)
		got := satAdd(1, c.cost(specs[0].spec.Type, shimPath, specs[0].imports))
		assert.LessOrEqual(t, got, want,
			"shim type %q now expands to %d, above the %d in gnoBuiltinShimExpansion: "+
				"every package names this type, so the table must not under-count it",
			name, got, want)
	}
}
