package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
	"path"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// typeExpansionBudget bounds the total number of nodes that go/types' validType
// walk would visit for a package's named types.
//
// go/types validates that named types do not "expand" indefinitely
// (src/go/types/validtype.go). Its walk follows value-containment edges only —
// struct fields, array elements, interface embeddeds and type-set terms, and a
// named type's underlying RHS — and crucially it does NOT memoize visited types (the
// optimization is commented out as a workaround for golang/go#65711). As a
// result the walk is exponential in the worst case: a "doubling" chain such as
//
//	type T0 struct{ x int }
//	type T1 struct{ a, b [0]T0 } // references the previous level TWICE by value
//	type T2 struct{ a, b [0]T1 }
//	// ... depth ~40 => 2^40 node visits
//
// hangs the type checker. Because TypeCheckMemPackage runs unmetered at deploy
// time (VMKeeper.AddPackage / MsgRun), a ~40-line package is a consensus DoS.
//
// checkTypeExpansionBound computes that same node-visit count, but WITH the
// memoization validType lacks, making it linear, and rejects the package before
// go/types runs if the count exceeds the budget.
//
// The budget is the TOTAL over the package's declared named types, because
// go/types calls validType once per declaration: the walk a transaction pays for
// is the sum, not the largest single type. Bounding only the largest leaves the
// sum unbounded, which is how this guard was first written and what it cost.
//
// 1_000_000 is set from measurement, not guessed: the largest per-package total
// in real code is 181, measured over all stdlibs and examples including their
// test files, and pinned by TestHonestTypeExpansionUnderBudget so it cannot rot.
// At the budget the walk costs ~21ms (~21ns/node on Apple Silicon go1.25; more on
// slower hardware). Because the bound is a total it also caps value-containment
// DEPTH near 1000, since a linear chain of depth d totals ~d^2 nodes; that is
// honest arithmetic, not a fan-out special case, and it is orders of magnitude
// past any real type.
//
// Note the budget is PER PACKAGE, which is not the same as per transaction: one
// MsgAddPackage re-type-checks (and so re-guards) every transitive user
// dependency, because the keeper's type-check cache holds only stdlibs and is
// cloned per tx; and a deploying package's prod declarations are walked twice,
// once for the prod Check and once with its in-package tests. So the walk a
// single tx can buy is ~2 * (packages checked) * budget, and the dependency count
// is bounded only by what was deployed earlier — bytes the importing tx does not
// pay for. Bounding the per-transaction sum is the next step up and wants its own
// calibration; see adr/pr5826_typecheck_dos_guards.md, which also records the
// alternatives weighed (gas-metering the walk, a governance Param).
//
// The budget is a deterministic node count, not a wall-clock limit, so the check
// is consensus-safe.
//
// MAINTENANCE: cost() below mirrors validType's containment edges for the go1.17
// subset Gno accepts. Two classes of construct are NOT modelled here because
// they are rejected before this runs; each rejection is therefore a precondition
// of this guard's soundness, not an independent nicety:
//
//   - go1.18 generic instantiation (type-argument substitution) and interface
//     type-set terms (unions, ~T). validType walks both, but cost() does not
//     follow them; checkNoUncountableGenerics rejects them.
//   - dot imports. cost() cannot see a dot-imported type's expansion (namedCost
//     resolves in the declaring package only); checkNoDotImports rejects them.
//
// Two further edges are under-counted rather than rejected, and stay safe only
// because their source is fixed and cannot grow with input: imported stdlib
// types (see expansionPkgResolver) and the `realm`/`address` names from the
// .gnobuiltins.gno shim, which is injected AFTER these guards run and so is
// scored as a leaf. Both are bounded by construction; see the ADR.
//
// Revisit this file if a toolchain upgrade adds a go1.17-reachable edge, if
// validType is finally memoized (golang/go#65711), or if Gno ever accepts
// generics/type-sets/dot imports (they would then have to be counted here, not
// rejected) — under-counting a live edge would silently reopen the DoS.
const typeExpansionBudget = 1_000_000

// pkgResolver returns the parsed Go source files of an already-deployed
// dependency package, or nil when the package should be treated as a leaf:
// stdlib (fixed, bounded source that no user chain can amplify), missing, or
// unparseable. It lets checkTypeExpansionBound follow value-containment edges
// across import boundaries, which is required because go/types' validType walk
// re-expands imported named types WITHOUT memoizing across packages
// (golang/go#65711) — so a doubling chain split over several packages stays
// under each package's own budget while the walk doubles at every link.
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

// pkgDecls indexes a package's type declarations by name.
type pkgDecls struct {
	byName map[string][]declWithImports
}

// expansionPkgCache caches the parsed sources and type declarations of resolved
// DEPENDENCY packages. One cache is shared by every expansionChecker of a single
// type check, because typeCheckMemPackage — and so this guard — runs once per
// imported package: without sharing, each nesting level re-fetches and re-parses
// its own dependencies, making dependency parsing quadratic in the import graph.
//
// It holds resolved sources only. A checker's own entry package is kept outside
// the cache (see expansionChecker.own): the entry is seeded with whichever
// file set its caller is type-checking, which for a top-level package includes
// test files, and must never be served in place of a dependency's prod-only
// sources, nor vice versa.
//
// Not safe for concurrent use; a cache belongs to one gnoImporter, whose type
// checks are sequential.
type expansionPkgCache struct {
	files map[string][]*ast.File
	decls map[string]*pkgDecls
}

func newExpansionPkgCache() *expansionPkgCache {
	return &expansionPkgCache{
		files: make(map[string][]*ast.File),
		decls: make(map[string]*pkgDecls),
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
func (c *expansionChecker) declsFor(pkgPath string) *pkgDecls {
	cache := c.cacheFor(pkgPath)
	if pd, ok := cache.decls[pkgPath]; ok {
		return pd
	}
	pd := &pkgDecls{byName: make(map[string][]declWithImports)}
	for _, gof := range c.filesFor(pkgPath) {
		imports := c.fileImports(gof)
		ast.Inspect(gof, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok {
				pd.byName[ts.Name.Name] = append(pd.byName[ts.Name.Name],
					declWithImports{spec: ts, imports: imports})
			}
			return true
		})
	}
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
		// Builtin, type parameter, or otherwise unresolved: validType leaf.
		return 1
	}
	if c.visiting[k] {
		// A value-containment cycle: an invalid recursive type that go/types
		// detects and reports itself. Return a finite count so we neither loop
		// nor pre-empt go/types' diagnostic.
		return 1
	}
	c.visiting[k] = true
	var best uint64 = 1
	for _, d := range specs {
		if v := satAdd(1, c.cost(d.spec.Type, k.pkg, d.imports)); v > best {
			best = v
		}
	}
	c.visiting[k] = false
	c.memo[k] = best
	return best
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
		return 1 // unresolvable qualifier: treat as leaf
	case *ast.IndexExpr:
		return c.cost(t.X, pkgPath, imports) // generic instantiation (rejected earlier)
	case *ast.IndexListExpr:
		return c.cost(t.X, pkgPath, imports)
	default:
		return 1
	}
}

// checkTypeExpansionBound rejects packages whose named types would cause
// go/types' validType walk to run super-linearly. Imports are treated as leaves;
// use checkTypeExpansionBoundImports to follow value-containment across packages.
// See typeExpansionBudget.
func checkTypeExpansionBound(fset *token.FileSet, gofs []*ast.File) error {
	return checkTypeExpansionBoundImports(fset, "", gofs, nil, nil)
}

// checkTypeExpansionBoundImports is checkTypeExpansionBound with cross-package
// resolution: entryPath is the deploying package's path, and resolve fetches the
// parsed source of its (already-deployed) dependencies. cache, if non-nil, is
// shared with the other type checks of the same importer so each dependency is
// parsed once rather than once per nesting level.
func checkTypeExpansionBoundImports(fset *token.FileSet, entryPath string, gofs []*ast.File, resolve pkgResolver, cache *expansionPkgCache) error {
	c := newExpansionChecker(entryPath, gofs, resolve, cache)

	// go/types runs validType once per declared named type, so the walk this
	// transaction pays for is the SUM over declarations — that is what the budget
	// bounds. Cost every declaration (not every name: a name declared more than
	// once is validated once per declaration). The total is order-independent, and
	// so is the verdict; for the message we name the costliest declaration, ties
	// broken by position, which is both deterministic (positions are unique) and
	// more useful than whichever declaration happened to cross the running total.
	var total, worstCost uint64
	var worst *ast.TypeSpec
	for _, specs := range c.declsFor(entryPath).byName {
		for _, d := range specs {
			cost := satAdd(1, c.cost(d.spec.Type, entryPath, d.imports))
			total = satAdd(total, cost)
			if worst == nil || cost > worstCost ||
				(cost == worstCost && d.spec.Name.Pos() < worst.Name.Pos()) {
				worst, worstCost = d.spec, cost
			}
		}
	}
	if total > typeExpansionBudget {
		return fmt.Errorf(
			"%s: this package's named types expand to %d nodes during type "+
				"validation, exceeding the limit of %d (largest: type %s at %d) "+
				"(possible denial-of-service vector)",
			fset.Position(worst.Name.Pos()), total, typeExpansionBudget,
			worst.Name.Name, worstCost)
	}
	return nil
}

// expansionPkgResolver returns a pkgResolver backed by the importer's getter,
// used to follow value-containment edges into already-deployed dependencies.
func (gimp *gnoImporter) expansionPkgResolver() pkgResolver {
	return func(pkgPath string) []*ast.File {
		// Treat stdlib types as leaves (count 1) instead of fetching+parsing them.
		// This is safe: the exponential vector is value-containment FAN-OUT, which
		// lives in user types and is fully counted — a user chain doubling over a
		// stdlib type still explodes the user-side count and trips the budget. A
		// stdlib type cannot import user packages, so its own expansion is fixed
		// and small, independent of input: measured max 28 over all stdlib types,
		// and only 19 (regexp.Regexp) over the EXPORTED ones, which are the only
		// ones a user package can name.
		//
		// So this only under-counts by a bounded per-reference constant, never
		// hides a fan-out, and does not compound with the aggregate budget (see
		// the ADR). We deliberately do NOT fetch stdlib source: go/types
		// serves stdlib imports from its result cache without a store read, so
		// fetching here would add store gas the deploy otherwise never pays.
		// (Counting stdlibs exactly is possible via a table precomputed at stdlib
		// load — no per-deploy gas — but the cross-module plumbing isn't worth it
		// for a leaf that is already bounded-safe. See adr/pr5826_typecheck_dos_guards.md.)
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

// checkNoUncountableGenerics rejects exactly the go1.18 constructs that cost()
// cannot model, before the package reaches go/types. It is a COST guard, not a
// generics gate: its job is to keep every construct that survives into go/types
// countable by the expansion bound above.
//
//   - type parameters, on a type or func declaration (`type W[P any] ...`):
//     cost() scores an instantiation by its base type and drops the type
//     arguments, so the doubling can hide there.
//   - interface type-set unions (`A | B`) and approximation terms (`~T`):
//     cost() falls through to a leaf for both, while validType walks every term.
//
// Bare and `;`-separated type-set elements are deliberately NOT rejected: they
// are ordinary containment edges the bound counts correctly, so they cannot hide
// a fan-out. Rejecting them would be a language-subset decision, not a cost one.
//
// Completeness with respect to generics is therefore NOT a goal here, and this
// guard must not be widened into one. Gno rejects unsupported syntax in Go2Gno,
// the single traversal every consumer shares — including the deploy path, and
// including gno run / the REPL / ParseFile, where go/types never runs at all.
// Tracked in #6059. Keep this function derivable from one rule: reject what
// cost() cannot count, nothing more.
//
// This must reject syntactically, before go/types is invoked: pinning
// types.Config.GoVersion does not help, because go/types reports the version
// error but does not halt, so it still runs the unmetered validType walk.
//
// It reports the earliest-positioned offending construct, so the error is
// deterministic (the message can reach consensus-visible tx results).
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
				// Reject only the type-set terms the expansion bound cannot count:
				// a union (`|`) and an approximation (`~`), which cost() treats as
				// leaves. `|` is only a type union in this position (elsewhere it is
				// bitwise-or) and `~` is exclusively type-approximation, so both are
				// unambiguous here. Other type-set elements (a bare or `;`-separated
				// type, also go1.18) are deliberately NOT rejected here: they are
				// ordinary containment edges the bound already counts, so they cannot
				// hang go/types — do not "fix" that by rejecting them.
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
		return fmt.Errorf("%s: %s are not supported (Gno targets go1.17)",
			fset.Position(off), what)
	}
	return nil
}

// errDotImports is the one wording for Gno's dot-import ban, shared with the
// preprocessor's own rejection so the two enforcement sites cannot drift.
const errDotImports = "dot imports are not allowed in Gno"

// checkNoDotImports rejects dot imports before the package reaches go/types.
// The preprocessor already rejects them, but on the deploy path it runs AFTER
// the type checker, so a dot import would spend the unmetered validType CPU
// first. It is also a hole in the bound above: a dot-imported type is named by a
// bare identifier, which namedCost scores as a leaf while validType expands it
// in full across the import boundary. Rejecting is preferred to counting because
// dot-import visibility is per file while namedCost memoizes per (package,
// name); see adr/pr5826_typecheck_dos_guards.md for the full argument.
//
// Like the guards above it reports the earliest-positioned offender, so the
// error does not depend on file order (it can reach consensus-visible results).
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
