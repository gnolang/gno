package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"math"
	"path"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// typeExpansionGasPerNode prices one node of the go/types validType walk, which
// the deploy path charges for BEFORE that walk runs.
//
// go/types validates that named types do not "expand" indefinitely
// (src/go/types/validtype.go). Its walk follows value-containment edges only —
// struct fields, array elements, interface embeddeds and type-set terms, and a
// named type's underlying RHS — and crucially it does NOT memoize visited types
// (the optimization is commented out as a workaround for golang/go#65711). So the
// walk is exponential in the worst case: a "doubling" chain such as
//
//	type T0 struct{ x int }
//	type T1 struct{ a, b [0]T0 } // references the previous level TWICE by value
//	type T2 struct{ a, b [0]T1 }
//	// ... depth ~40 => 2^40 node visits
//
// hangs the type checker, and TypeCheckMemPackage runs unmetered at deploy time
// (VMKeeper.AddPackage / MsgRun), so a ~40-line package was a consensus DoS.
//
// typeExpansionCost below computes that same node-visit count WITH the memoization
// validType lacks, which makes computing it linear, and the deploy path charges
// the count to TypeCheckOptions.GasMeter before handing the package to go/types.
// The count is a deterministic node count, not a wall-clock measurement, so the
// charge is consensus-safe.
//
// The rate sits here, beside the cost model that produces the count, for the same
// reason tokenCostFactor and the OpCPU* tables do: it is a measured ns rate for
// work gnovm performs, not a governance knob. Promoting it to a vm Params field is
// a reasonable future step.
//
// DERIVATION. 1 gas == 1ns of wall time on reference hardware, which for this repo
// is an Intel Xeon Platinum 8168 @ 2.70GHz (see the OpCPU* table in machine.go and
// gnovm/cmd/calibrate/README.md). Two steps, and the second is easy to forget:
//
//  1. Measure the walk. BenchmarkValidTypeWalk reports ns/node over a doubling
//     chain: 30.1 / 30.4 / 34.9 at depth 18 / 20 / 22 on an Apple M5. The marginal
//     rate between successive depths keeps climbing — 34.3 / 35.8 / 37.3 / 38.6 /
//     40.3 from depth 22 to 26 — because the working set outgrows cache. A DoS is
//     the large-working-set end, so ~40 is the figure to price, not ~30.
//  2. Calibrate to reference hardware. gnovm/cmd/calibrate ships paired benchmark
//     output for the Xeon (bench_output_do_dedicated.txt) and Apple silicon; over
//     the 37 shared BenchmarkAlloc cases the Xeon is 2.96x slower (median), and
//     2.2-3.2x on the small allocations that most resemble validType's pointer
//     chasing. So 40 * ~2.5 = 100.
//
// At 100, a whole block of gas (tm2's MaxBlockMaxGas = 3e9) buys 3e7 nodes, about
// 3s of validType on reference hardware — which is what "1 gas == 1ns" is meant to
// mean for a full block.
//
// An earlier revision of this change used 25: the same walk measured on the
// development machine with step 2 omitted. That under-charged by ~4x, and with no
// ceiling behind the charge one block would have bought ~12s of walk. Re-measuring
// directly on reference hardware is how to tighten this — the calibration factor is
// the dominant uncertainty here, exactly as it is for PreprocessGasPerByte.
//
// PRICED, NOT CAPPED. There is deliberately no ceiling on the count. Any ceiling
// sits either below what a sender can pay, in which case it refuses packages that
// were paid for, or above it, where nothing is payable anyway and it merely
// relabels an out-of-gas as something else. It would also have to be per-package,
// a scope no budget has: Tx.Msgs is unbounded and one message re-checks every
// dependency it imports, whose source bytes earlier transactions paid for. An
// earlier revision of this change carried a 1_000_000 ceiling; it was removed for
// exactly those reasons. What that costs is off-chain: unmetered callers (gno test,
// gno lint, gnodev) churn on a pathological local package for as long as `go build`
// does on the same input, so the walk is bounded there only by the developer
// interrupting it.
//
// SOUNDNESS. The count is now the only thing between a deploy and an unmetered
// exponential walk, so it must never UNDER-count a live containment edge:
// under-counting is under-charging, which is the same denial of service at a
// discount. cost() mirrors validType's edges for the Go subset Gno accepts, and
// the two construct classes it cannot model are rejected outright rather than
// approximated — each rejection is a precondition of this pricing, not a nicety:
//
//   - go1.18 generic instantiation (type-argument substitution) and interface
//     type-set terms (unions, ~T). validType walks both, but cost() does not
//     follow them; checkNoUncountableGenerics rejects them.
//   - dot imports. cost() cannot see a dot-imported type's expansion (namedCost
//     resolves in the declaring package only); checkNoDotImports rejects them.
//
// A named type the guard cannot resolve — an imported stdlib type, or the
// `realm`/`address` names from the .gnobuiltins.gno shim, which is injected after
// these guards run — is scored at leafExpansionBound rather than 1, so those edges
// over-count instead of under-counting.
//
// Revisit this file if a toolchain upgrade adds a containment edge, if validType is
// finally memoized (golang/go#65711), or if Gno ever accepts generics, type sets or
// dot imports: they would then have to be counted here, not rejected.
const typeExpansionGasPerNode = 100

// gnoBuiltinShimExpansion is the exact expansion of the type names
// .gnobuiltins.gno injects into every package. That file is added AFTER these
// guards run, so cost() never sees the declarations and would otherwise score both
// as unresolved. It cannot: `address` is the type of most of the stdlib's public
// API surface, so charging it leafExpansionBound would tax honest deploys hundreds
// of times over for a two-node type. TestGnoBuiltinShimExpansion pins these
// against the real shim source, in case a future shim grows an edge.
var gnoBuiltinShimExpansion = map[string]uint64{
	// Both are one node for the named type plus one for its underlying type: an
	// interface declaring only methods (validType walks no method signature), and
	// a named string.
	"realm":   2,
	"address": 2,
}

// leafExpansionBound is the expansion charged for an imported stdlib type, which
// this guard scores as a leaf because expansionPkgResolver deliberately does not
// resolve stdlib: go/types serves those imports from its own result cache without
// a store read, so resolving them here would add store gas the deploy otherwise
// never pays. Predeclared names and the shim above are excluded — those are scored
// exactly.
//
// Scoring such a leaf as 1 would UNDER-count, and an under-count multiplies: a
// leaf at the base of a doubling chain of depth d is walked 2^d times, so an
// under-count of k there is a k-fold discount on the whole package. Charging the
// largest expansion any leaf can actually have makes the edge an over-count
// instead, which is the direction pricing has to err.
//
// It is safe as a constant because the set of stdlib types is fixed by the binary,
// not by transaction input: stdlib source ships with the node, and a stdlib type
// cannot reference a user package. TestLeafExpansionBound measures the real
// maximum over every exported stdlib type and fails if this no longer covers it.
// 32: the measured maximum over every exported stdlib type is 19 (regexp.Regexp),
// and the margin is small on purpose — every reference to a stdlib type in honest
// code pays this, so headroom here is a tax on ordinary deploys.
const leafExpansionBound = 32

// expansionGas converts a node count into gas, clamping at math.MaxInt64 instead
// of converting a saturated count straight to int64: math.MaxUint64 as an int64 is
// -1, and -1 * typeExpansionGasPerNode is a NEGATIVE charge, which would refund
// gas for the most expensive package a sender could submit. Clamped, an
// unrepresentable count is simply unaffordable and the deploy runs out of gas.
func expansionGas(nodes uint64) int64 {
	const max = uint64(math.MaxInt64) / typeExpansionGasPerNode
	if nodes > max {
		return math.MaxInt64
	}
	return int64(nodes) * typeExpansionGasPerNode
}

// pkgResolver returns the parsed Go source files of an already-deployed
// dependency package, or nil when the package should be treated as a leaf:
// stdlib (fixed, bounded source that no user chain can amplify), missing, or
// unparseable. It lets typeExpansionCost follow value-containment edges
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
		if types.Universe.Lookup(k.name) != nil {
			// Predeclared (int, string, error, any, ...). validType stops at these,
			// so 1 is exact, not an approximation.
			return 1
		}
		if v, ok := gnoBuiltinShimExpansion[k.name]; ok {
			return v
		}
		// An imported stdlib type: expansionPkgResolver does not resolve those, so
		// score the largest expansion any of them can have. Over-counting here is
		// the safe direction; see leafExpansionBound.
		return leafExpansionBound
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

// typeExpansionCost returns the number of nodes go/types' validType walk will
// visit for one package's named types, for the deploy path to charge before that
// walk runs. See typeExpansionGasPerNode.
//
// entryPath is the package being checked and resolve fetches the parsed source of
// its (already-deployed) dependencies, so value containment is followed across
// import boundaries — validType re-expands imported named types without
// memoizing, so a doubling chain split over several packages still doubles at
// every link. cache, if non-nil, is shared with the other type checks of the same
// importer so each dependency is parsed once rather than once per nesting level.
//
// The count saturates at math.MaxUint64 rather than wrapping; callers converting
// it to gas must clamp (see the charge site in gotypecheck.go).
func typeExpansionCost(entryPath string, gofs []*ast.File, resolve pkgResolver, cache *expansionPkgCache) uint64 {
	c := newExpansionChecker(entryPath, gofs, resolve, cache)

	// go/types runs validType once per declared named type, so the walk this
	// transaction pays for is the SUM over declarations. Cost every declaration,
	// not every name: a name declared more than once is validated once per
	// declaration. The total is order-independent.
	var total uint64
	for _, specs := range c.declsFor(entryPath).byName {
		for _, d := range specs {
			total = satAdd(total, satAdd(1, c.cost(d.spec.Type, entryPath, d.imports)))
		}
	}
	return total
}

// expansionPkgResolver returns a pkgResolver backed by the importer's getter,
// used to follow value-containment edges into already-deployed dependencies.
func (gimp *gnoImporter) expansionPkgResolver() pkgResolver {
	return func(pkgPath string) []*ast.File {
		// Do not fetch stdlib source: go/types serves stdlib imports from its own
		// result cache without a store read, so fetching here would add store gas
		// the deploy otherwise never pays. namedCost scores what it cannot resolve
		// at leafExpansionBound, which is the measured maximum over every exported
		// stdlib type (19, regexp.Regexp) — so a stdlib reference OVER-counts rather
		// than under-charging, and a user chain doubling over a stdlib type is
		// priced at least as high as the walk it causes.
		//
		// This holds only because a stdlib type's own expansion is fixed by the
		// binary: stdlib source ships with the node and cannot import user packages,
		// so no transaction can grow it. TestLeafExpansionBound pins the measurement.
		// Counting stdlib exactly is still possible via a table precomputed at
		// stdlib load — no per-deploy gas — but buys exactness for an edge that is
		// already priced above its cost. See adr/pr5826_typecheck_dos_guards.md.
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
// countable by the cost model above.
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
				// Reject only the type-set terms the cost model cannot count:
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
		return fmt.Errorf("%s: %s are not supported",
			fset.Position(off), what)
	}
	return nil
}

// errDotImports is the one wording for Gno's dot-import ban, shared with the
// preprocessor's own rejection so the two enforcement sites cannot drift.
//
// The ban is enforced (preprocess panics at two sites, reachable — `gno run` on
// a dot import fails there) but not documented, has no recorded rationale, and
// gnovm/tests/backup/import2.gno still asserts the opposite; see #6076. This
// guard does not depend on why Gno bans them — see checkNoDotImports.
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
