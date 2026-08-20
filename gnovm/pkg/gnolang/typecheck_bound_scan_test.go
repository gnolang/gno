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

// TestHonestTypeExpansionUnderBudget pins the measurement that typeExpansionCeiling
// is sized against: the largest expansion total of any real package. That number
// is the entire false-rejection argument for the budget, so without this test it
// would rot silently — a future stdlib or example package with a genuinely large
// type graph could creep toward the budget and nobody would notice until a
// legitimate deploy started failing.
//
// It scans every stdlib and example package INCLUDING test files (the guard runs
// on the whole file set for the package being checked) and asserts a wide margin
// rather than an exact figure, so ordinary growth does not churn the test.
func TestHonestTypeExpansionUnderBudget(t *testing.T) {
	t.Parallel()

	const (
		// The margin the budget's rationale claims. Well below the budget, but far
		// above today's max (181), so this fails long before a real deploy would.
		maxHonestTotal = typeExpansionCeiling / 100
		// Guard against the scan silently finding nothing (a moved directory would
		// otherwise make this test vacuously pass).
		minPackages = 100
	)

	fset := token.NewFileSet()
	roots := map[string]string{
		"":          "../../stdlibs",               // stdlib import paths are bare
		"gno.land/": "../../../examples/gno.land/", // examples live under gno.land/
	}

	// resolve parses a package's .gno files, mirroring what the deploy-path
	// resolver hands the guard, but reaching source on disk instead of the store.
	resolve := func(pkgPath string) []*ast.File {
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
				src, err := os.ReadFile(filepath.Join(dir, e.Name()))
				if err != nil {
					continue
				}
				// Skip anything that does not parse as Go (e.g. not-yet-transpiled
				// sources); the guard treats unparseable dependencies as leaves too.
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

	// Enumerate every directory holding .gno files, as an import path.
	var pkgPaths []string
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
					pkgPaths = append(pkgPaths, prefix+filepath.ToSlash(rel))
					return nil
				}
			}
			return nil
		}))
	}
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
		c := newExpansionChecker(pp, gofs, userResolve, shared)
		var total uint64
		for _, specs := range c.declsFor(pp).byName {
			for _, d := range specs {
				nTypes++
				total = satAdd(total, satAdd(1, c.cost(d.spec.Type, pp, d.imports)))
			}
		}
		if total > worstTotal {
			worstTotal, worstPkg = total, pp
		}
	}

	t.Logf("scanned %d packages, %d named types; largest per-package expansion total "+
		"is %d (%s), budget is %d", len(pkgPaths), nTypes, worstTotal, worstPkg, typeExpansionCeiling)
	assert.LessOrEqual(t, worstTotal, uint64(maxHonestTotal),
		"package %s expands to %d nodes, within %dx of typeExpansionCeiling (%d). "+
			"Real code is approaching the DoS budget: re-derive the budget (see its doc "+
			"comment) rather than just relaxing this test",
		worstPkg, worstTotal, typeExpansionCeiling/maxHonestTotal, typeExpansionCeiling)
}
