package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"math"
	"path"
	"slices"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// go/types' validType walk is exponential on value-containment fan-out (it does not
// memoize, golang/go#65711) and runs unmetered at deploy time, so the deploy path
// counts the nodes it will visit and charges for them first. No ceiling backs that
// charge up, so cost() must never UNDER-count a live edge — that is why the guards
// below reject instead of approximating. 100 = ~40ns/node from BenchmarkValidTypeWalk
// x ~2.5 calibration to the Xeon that "1 gas == 1ns" means; adr/pr5826_typecheck_dos_guards.md.
const typeExpansionGasPerNode = 100

// gnoBuiltinShimExpansion is the exact expansion of the type names .gnobuiltins.gno
// injects into every package. cost() never sees their declarations (the shim is
// added after these guards run), and `address` is too common in the stdlib API to
// approximate. Pinned by TestGnoBuiltinShimExpansion.
//
// NOT keyed by gno version, while makeGnoBuiltins is (it panics on anything but
// GnoVerLatest). A future 0.10 shim would silently keep 0.9's numbers, and the
// pinning test only checks 0.9 — so key this by (version, name) when a second
// shim version appears. Failure mode is over- or under-charge, not unsoundness
// today, because 0.9 is the only shim that exists.
var gnoBuiltinShimExpansion = map[string]uint64{
	// One node for the named type plus one for its underlying type: an interface
	// declaring only methods (validType walks no method signature), and a string.
	"realm":   2,
	"address": 2,
}

// leafExpansionBound is charged for an imported stdlib type, which the guard scores
// as a leaf because expansionPkgResolver does not resolve stdlib. 32 covers the
// measured maximum over every exported stdlib type (19, regexp.Regexp) with
// deliberately thin margin, since honest code pays this per reference. Pinned by
// TestLeafExpansionBound; see the ADR for why over-counting is the safe direction.
const leafExpansionBound = 32

// expansionGas converts a node count to gas, clamping instead of wrapping:
// int64(math.MaxUint64) is -1, so a saturated count would charge NEGATIVE gas.
func expansionGas(nodes uint64) int64 {
	if g := satMul(nodes, typeExpansionGasPerNode); g <= math.MaxInt64 {
		return int64(g)
	}
	return math.MaxInt64
}

// pkgResolver returns the parsed Go source of an already-deployed dependency, or
// nil to score it as a leaf (stdlib, missing, unparseable). It exists because
// validType re-expands imported types without memoizing, so a doubling chain split
// over several packages still doubles at every link.
type pkgResolver func(pkgPath string) []*ast.File

// typeKey identifies a named type by its declaring package path and name.
// The entry (deploying) package uses the empty path "".
type typeKey struct {
	pkg  string
	name string
}

// declWithImports pairs a type declaration with the import aliases in scope for
// the file it was declared in, so a qualified reference (pkg.T) inside it can be
// resolved to the imported package's path.
type declWithImports struct {
	spec    *ast.TypeSpec
	imports map[string]string // selector name -> import path
}

// pkgDecls indexes a package's type declarations by name. A name can appear more
// than once: validType runs per declaration, so all of them are kept.
//
// names carries the same keys, SORTED, and is what callers must iterate. Two
// reasons it is not a map range and not declaration order:
//   - a map range is not reproducible, and the total is not order-independent
//     (see typeExpansionCost), so the charge would vary per run;
//   - declaration order is reproducible but layout-sensitive — writing the two
//     members of a containment cycle in the other order changes the price 2x
//     (measured: 3076 vs 6136 nodes). Sorting depends only on the set of names,
//     so moving a declaration between files or reordering for readability does
//     not move anyone's gas.
type pkgDecls struct {
	byName map[string][]declWithImports
	names  []string
}

// expansionPkgCache caches resolved DEPENDENCY sources and declarations, shared by
// every expansionChecker of one type check — without it, dependency parsing is
// quadratic in the import graph. A checker's own entry package stays outside it
// (see expansionChecker.own): the entry is seeded with whichever file set its caller
// is checking, so it must never stand in for a dependency's prod-only sources.
// Not safe for concurrent use; one cache per gnoImporter, whose checks are serial.
type expansionPkgCache struct {
	files map[string][]*ast.File
	decls map[string]pkgDecls
}

func newExpansionPkgCache() *expansionPkgCache {
	return &expansionPkgCache{
		files: make(map[string][]*ast.File),
		decls: make(map[string]pkgDecls),
	}
}

// expansionChecker computes, with the memoization validType lacks, the node
// count validType would visit — following value-containment edges within AND
// across packages. Memoization is keyed by (package, name), so the cross-package
// walk that validType runs exponentially is computed linearly here.
type expansionChecker struct {
	resolve   pkgResolver
	entryPath string
	own       *expansionPkgCache // entry package only; never shared with siblings
	shared    *expansionPkgCache // resolved dependencies
	memo      map[typeKey]uint64
	visiting  map[typeKey]bool
}

// newExpansionChecker returns a checker for entryPath, whose files are gofs.
// shared may be nil, in which case the checker gets a private cache.
func newExpansionChecker(entryPath string, gofs []*ast.File, resolve pkgResolver, shared *expansionPkgCache) *expansionChecker {
	if resolve == nil {
		resolve = func(string) []*ast.File { return nil }
	}
	if shared == nil {
		shared = newExpansionPkgCache()
	}
	own := newExpansionPkgCache()
	own.files[entryPath] = gofs // seed the entry package; never fetch it
	return &expansionChecker{
		resolve:   resolve,
		entryPath: entryPath,
		own:       own,
		shared:    shared,
		memo:      make(map[typeKey]uint64),
		visiting:  make(map[typeKey]bool),
	}
}

// cacheFor returns the cache owning pkgPath: this checker's private one for its
// entry package, the shared one for resolved dependencies. Keeping the entry out
// of the shared cache is what makes sharing safe — see expansionPkgCache.
func (c *expansionChecker) cacheFor(pkgPath string) *expansionPkgCache {
	if pkgPath == c.entryPath {
		return c.own
	}
	return c.shared
}

// filesFor returns a package's parsed files, resolving (and caching) on demand.
func (c *expansionChecker) filesFor(pkgPath string) []*ast.File {
	cache := c.cacheFor(pkgPath)
	if fs, ok := cache.files[pkgPath]; ok {
		return fs
	}
	fs := c.resolve(pkgPath)
	cache.files[pkgPath] = fs
	return fs
}

// pkgName returns a package's declared name (for resolving unaliased imports).
func (c *expansionChecker) pkgName(pkgPath string) string {
	fs := c.filesFor(pkgPath)
	if len(fs) == 0 || fs[0].Name == nil {
		return ""
	}
	return fs[0].Name.Name
}

// fileImports maps each selector-visible import name in a file to its path.
func (c *expansionChecker) fileImports(gof *ast.File) map[string]string {
	m := make(map[string]string, len(gof.Imports))
	for _, imp := range gof.Imports {
		impPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		name := ""
		if imp.Name != nil {
			// "_" (side-effect) and "." (dot) imports introduce no selector name.
			// (Dot imports are separately rejected; see checkNoDotImports.)
			if imp.Name.Name == "_" || imp.Name.Name == "." {
				continue
			}
			name = imp.Name.Name
		} else if name = c.pkgName(impPath); name == "" {
			// Unresolved (e.g. stdlib, treated as a leaf): fall back to the path's
			// last element, the conventional package name.
			name = path.Base(impPath)
		}
		m[name] = impPath
	}
	return m
}

// declsFor collects a package's type declarations (package-level and
// function-local; local types are validated too, so they are also a vector).
// On name collision we keep all candidates and take the highest cost — a safe
// over-approximation that never under-counts.
func (c *expansionChecker) declsFor(pkgPath string) pkgDecls {
	cache := c.cacheFor(pkgPath)
	if pd, ok := cache.decls[pkgPath]; ok {
		return pd
	}
	pd := pkgDecls{byName: make(map[string][]declWithImports)}
	for _, gof := range c.filesFor(pkgPath) {
		imports := c.fileImports(gof)
		ast.Inspect(gof, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			name := ts.Name.Name
			if _, seen := pd.byName[name]; !seen {
				pd.names = append(pd.names, name)
			}
			pd.byName[name] = append(pd.byName[name],
				declWithImports{spec: ts, imports: imports})
			// A type expression holds no further declarations, so stop here.
			return false
		})
	}
	slices.Sort(pd.names)
	cache.decls[pkgPath] = pd
	return pd
}

// namedCost returns the count for a named type: one node for the type itself
// plus the cost of its underlying type, memoized by (package, name).
func (c *expansionChecker) namedCost(k typeKey) uint64 {
	if v, ok := c.memo[k]; ok {
		return v
	}
	specs := c.declsFor(k.pkg).byName[k.name]
	if len(specs) == 0 {
		v := unresolvedCost(k.name)
		c.memo[k] = v // depends only on the key, so it is safe to remember
		return v
	}
	if c.visiting[k] {
		// A value-containment cycle: an invalid recursive type that go/types
		// detects and reports itself. Return a finite count so we neither loop
		// nor pre-empt go/types' diagnostic. Deliberately not memoized — the
		// truncated value is only valid inside this walk.
		return 1
	}
	c.visiting[k] = true
	var best uint64 = 1
	for _, d := range specs {
		if v := satAdd(1, c.cost(d.spec.Type, k.pkg, d.imports)); v > best {
			best = v
		}
	}
	delete(c.visiting, k)
	c.memo[k] = best
	return best
}

// unresolvedCost prices a named type the guard cannot resolve to a declaration.
// Every such case routes through here, so the file has one answer to "what does an
// unknown name cost" rather than one per call site.
func unresolvedCost(name string) uint64 {
	if types.Universe.Lookup(name) != nil {
		// Predeclared (int, string, error, any, ...). validType stops at these, so
		// 1 is exact, not an approximation.
		return 1
	}
	if v, ok := gnoBuiltinShimExpansion[name]; ok {
		return v
	}
	// An imported stdlib type: expansionPkgResolver does not resolve those, so
	// score the largest expansion any of them can have. Over-counting here is the
	// safe direction; see leafExpansionBound.
	return leafExpansionBound
}

// cost returns the number of validType nodes visited for a type expression,
// evaluated in the context of package pkgPath and the imports of the file the
// expression came from.
func (c *expansionChecker) cost(e ast.Expr, pkgPath string, imports map[string]string) uint64 {
	switch t := e.(type) {
	case *ast.ParenExpr:
		return c.cost(t.X, pkgPath, imports)
	case *ast.StarExpr:
		return 1 // pointer: validType does not recurse
	case *ast.ArrayType:
		if t.Len == nil {
			return 1 // slice: not recursed
		}
		return satAdd(1, c.cost(t.Elt, pkgPath, imports)) // array: recurse into element
	case *ast.MapType, *ast.ChanType, *ast.FuncType:
		return 1 // not recursed
	case *ast.StructType:
		total := uint64(1)
		for _, f := range t.Fields.List {
			mult := uint64(len(f.Names))
			if mult == 0 {
				mult = 1 // embedded field
			}
			total = satAdd(total, satMul(mult, c.cost(f.Type, pkgPath, imports)))
		}
		return total
	case *ast.InterfaceType:
		total := uint64(1)
		for _, f := range t.Methods.List {
			if len(f.Names) != 0 {
				continue // method: a func signature, not recursed
			}
			total = satAdd(total, c.cost(f.Type, pkgPath, imports)) // embedded type
		}
		return total
	case *ast.Ident:
		return c.namedCost(typeKey{pkg: pkgPath, name: t.Name})
	case *ast.SelectorExpr:
		// A qualified type pkg.T: resolve the import and cross into it. validType
		// re-walks imported types without memoizing, so this edge is real.
		if id, ok := t.X.(*ast.Ident); ok {
			if path := imports[id.Name]; path != "" {
				return c.namedCost(typeKey{pkg: path, name: t.Sel.Name})
			}
		}
		// Unresolvable qualifier. Priced like any other unresolvable name rather
		// than as a 1-node leaf: same situation, so the same (over-counting)
		// answer, or this becomes the one edge that under-charges.
		return unresolvedCost(t.Sel.Name)
	case *ast.IndexExpr:
		return c.cost(t.X, pkgPath, imports) // generic instantiation (rejected earlier)
	case *ast.IndexListExpr:
		return c.cost(t.X, pkgPath, imports)
	default:
		return 1
	}
}

// typeExpansionCost returns the number of nodes go/types' validType walk will visit
// for entryPath's named types, for the deploy path to charge before that walk runs.
// resolve follows value containment into already-deployed dependencies; cache, if
// non-nil, is shared with this importer's other type checks. Saturates at
// math.MaxUint64 rather than wrapping — convert with expansionGas.
func typeExpansionCost(entryPath string, gofs []*ast.File, resolve pkgResolver, cache *expansionPkgCache) uint64 {
	c := newExpansionChecker(entryPath, gofs, resolve, cache)

	// validType runs once per DECLARATION, so the charge is the sum over them, not
	// over names (a name declared twice is validated twice).
	//
	// SORTED, never map order. namedCost truncates a value-containment cycle at
	// whichever member it is already visiting, and correctly does not memoize that
	// 1 — but every ANCESTOR on the cycle memoizes a value DERIVED from the
	// truncation, and which member gets truncated depends on which root is walked
	// first. Ranging the map therefore priced the same source at 3076 or 6136 nodes,
	// 306k gas apart, at random. This count is charged as gas, so a node that
	// disagrees forks: ABCIResult.Error is hashed into LastResultsHash, and runTx
	// charges GasConsumedToLimit() to the BlockGasMeter.
	//
	// Only invalid recursive types are affected, and go/types rejects those a moment
	// later — but the charge lands BEFORE it does, so one cheap malformed package
	// was enough. TestTypeExpansionCostCyclicIsDeterministic pins this.
	var total uint64
	decls := c.declsFor(entryPath)
	for _, name := range decls.names {
		for _, d := range decls.byName[name] {
			total = satAdd(total, satAdd(1, c.cost(d.spec.Type, entryPath, d.imports)))
		}
	}
	return total
}

// expansionPkgResolver returns a pkgResolver backed by the importer's getter,
// used to follow value-containment edges into already-deployed dependencies.
func (gimp *gnoImporter) expansionPkgResolver() pkgResolver {
	return func(pkgPath string) []*ast.File {
		// Never fetch stdlib: go/types serves those imports from its own cache
		// without a store read, so fetching would add store gas the deploy otherwise
		// never pays. namedCost prices them at leafExpansionBound instead, which
		// over-counts; safe because stdlib source is fixed by the binary. See the ADR.
		if IsStdlib(pkgPath) {
			return nil
		}
		mpkg := gimp.getter.GetMemPackage(pkgPath)
		if mpkg == nil {
			return nil
		}
		mpkg = MPFProd.FilterMemPackage(mpkg)
		_, allgofs, _, _, _, err := GoParseMemPackage(mpkg, token.NewFileSet())
		if err != nil {
			return nil
		}
		return allgofs
	}
}

// satAdd returns a+b, saturating at math.MaxUint64 on overflow.
func satAdd(a, b uint64) uint64 {
	if s, ok := overflow.Add(a, b); ok {
		return s
	}
	return math.MaxUint64
}

// satMul returns a*b, saturating at math.MaxUint64 on overflow.
func satMul(a, b uint64) uint64 {
	if p, ok := overflow.Mul(a, b); ok {
		return p
	}
	return math.MaxUint64
}

// checkNoUncountableGenerics rejects exactly the go1.18 constructs cost() cannot
// model — type parameters, and interface type-set `|` / `~` terms — before the
// package reaches go/types. It is a COST guard, not a generics gate: do NOT widen
// it into one (generics completeness belongs in Go2Gno, #6059; see the ADR).
// Reports the earliest offending construct, so the message is deterministic.
func checkNoUncountableGenerics(fset *token.FileSet, gofs []*ast.File) error {
	var (
		off  token.Pos
		what string
	)
	note := func(pos token.Pos, kind string) {
		if !off.IsValid() || pos < off {
			off, what = pos, kind
		}
	}
	for _, gof := range gofs {
		ast.Inspect(gof, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.TypeSpec:
				if t.TypeParams != nil {
					note(t.Name.Pos(), "generic type declarations")
				}
			case *ast.FuncType:
				if t.TypeParams != nil {
					note(t.Pos(), "generic functions")
				}
			case *ast.InterfaceType:
				// Only `|` and `~`, which cost() treats as leaves. Bare and
				// `;`-separated type-set elements are ordinary containment edges the
				// cost model counts correctly — do NOT "fix" that by rejecting them.
				for _, f := range t.Methods.List {
					if len(f.Names) != 0 {
						continue // a method, not a type-set term
					}
					switch e := f.Type.(type) {
					case *ast.BinaryExpr:
						if e.Op == token.OR {
							note(e.Pos(), "interface type unions")
						}
					case *ast.UnaryExpr:
						if e.Op == token.TILDE {
							note(e.Pos(), "interface approximation (~) terms")
						}
					}
				}
			}
			return true
		})
	}
	if off.IsValid() {
		return fmt.Errorf("%s: %s are not supported",
			fset.Position(off), what)
	}
	return nil
}

// checkNoDotImports rejects dot imports before go/types. The preprocessor rejects
// them too, but on the deploy path it runs AFTER the type checker — and a
// dot-imported type is named by a bare identifier, which namedCost scores as a leaf
// while validType expands it in full. Rejected rather than counted; see the ADR.
// Reports the earliest offender, so the message does not depend on file order.
func checkNoDotImports(fset *token.FileSet, gofs []*ast.File) error {
	var off token.Pos
	for _, gof := range gofs {
		for _, imp := range gof.Imports {
			if imp.Name == nil || imp.Name.Name != "." {
				continue
			}
			if !off.IsValid() || imp.Name.Pos() < off {
				off = imp.Name.Pos()
			}
		}
	}
	if off.IsValid() {
		return fmt.Errorf("%s: %s", fset.Position(off), errDotImports)
	}
	return nil
}
