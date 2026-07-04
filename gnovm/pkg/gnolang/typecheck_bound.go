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

// typeExpansionBudget bounds the number of nodes that go/types' validType walk
// would visit for any single named type in a package.
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
// go/types runs if any named type's count exceeds the budget. The budget is a
// deterministic count (not a wall-clock limit), so the check is consensus-safe.
//
// 100_000 is set from measurement, not guessed: across all stdlib and example
// packages (357 packages, 877 named types) the largest expansion count is 35, so
// this leaves ~3 orders of magnitude of headroom and never false-rejects honest
// code — even large generated packages stay in the low thousands. Because the
// DoS is exponential (counts of 2^40+), any budget between honest code and the
// blowup separates them cleanly. At the bound, go/types' validType walk takes
// ~8ms (measured, ~80ns/node on Apple Silicon go1.25; more on slower hardware) —
// bounded and one-time per gas-paying tx, not the unbounded hang it replaces.
// The budget is a deterministic node count (not a wall-clock limit), so the
// check is consensus-safe.
//
// MAINTENANCE: cost() below mirrors validType's containment edges for the go1.17
// subset Gno accepts. validType also walks two go1.18 edges — generic
// instantiation (type-argument substitution) and interface type-set terms
// (unions, ~T) — but cost() does NOT follow them: checkNoGenerics rejects those
// constructs before this runs, so they never reach validType. Revisit this file
// if a toolchain upgrade adds a go1.17-reachable edge, if validType is finally
// memoized (golang/go#65711), or if Gno ever accepts generics/type-sets (they
// would then have to be counted here, not rejected) — under-counting a live edge
// would silently reopen the DoS.
const typeExpansionBudget = 100_000

// pkgResolver returns the parsed Go source files of an already-deployed
// dependency package, or nil when the package should be treated as a leaf:
// stdlib (fixed, bounded source that no user chain can amplify), missing, or
// unparseable. It lets checkTypeExpansionBound follow value-containment edges
// across import boundaries, which is required because go/types' validType walk
// re-expands imported named types WITHOUT memoizing across packages
// (golang/go#65711) — so a doubling chain split over several packages stays
// under the per-package budget while the walk doubles at every link.
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

// expansionChecker computes, with the memoization validType lacks, the node
// count validType would visit — following value-containment edges within AND
// across packages. Memoization is keyed by (package, name), so the cross-package
// walk that validType runs exponentially is computed linearly here.
type expansionChecker struct {
	resolve  pkgResolver
	files    map[string][]*ast.File // parsed files per pkg path (entry pre-seeded)
	decls    map[string]*pkgDecls   // decls per pkg path
	memo     map[typeKey]uint64
	visiting map[typeKey]bool
}

func newExpansionChecker(resolve pkgResolver) *expansionChecker {
	if resolve == nil {
		resolve = func(string) []*ast.File { return nil }
	}
	return &expansionChecker{
		resolve:  resolve,
		files:    make(map[string][]*ast.File),
		decls:    make(map[string]*pkgDecls),
		memo:     make(map[typeKey]uint64),
		visiting: make(map[typeKey]bool),
	}
}

// filesFor returns a package's parsed files, resolving (and caching) on demand.
func (c *expansionChecker) filesFor(pkgPath string) []*ast.File {
	if fs, ok := c.files[pkgPath]; ok {
		return fs
	}
	fs := c.resolve(pkgPath)
	c.files[pkgPath] = fs
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
	if pd, ok := c.decls[pkgPath]; ok {
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
	c.decls[pkgPath] = pd
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
			if total > typeExpansionBudget {
				return total
			}
		}
		return total
	case *ast.InterfaceType:
		total := uint64(1)
		for _, f := range t.Methods.List {
			if len(f.Names) != 0 {
				continue // method: a func signature, not recursed
			}
			total = satAdd(total, c.cost(f.Type, pkgPath, imports)) // embedded type
			if total > typeExpansionBudget {
				return total
			}
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
	return checkTypeExpansionBoundImports(fset, "", gofs, nil)
}

// checkTypeExpansionBoundImports is checkTypeExpansionBound with cross-package
// resolution: entryPath is the deploying package's path, and resolve fetches the
// parsed source of its (already-deployed) dependencies.
func checkTypeExpansionBoundImports(fset *token.FileSet, entryPath string, gofs []*ast.File, resolve pkgResolver) error {
	c := newExpansionChecker(resolve)
	c.files[entryPath] = gofs // seed the entry package; do not fetch it

	// Report the earliest-declared offending type, so the error is deterministic
	// regardless of map iteration order (the message can reach consensus-visible
	// tx results).
	var off *ast.TypeSpec
	var offVal uint64
	for _, specs := range c.declsFor(entryPath).byName {
		for _, d := range specs {
			if v := satAdd(1, c.cost(d.spec.Type, entryPath, d.imports)); v > typeExpansionBudget &&
				(off == nil || d.spec.Name.Pos() < off.Name.Pos()) {
				off, offVal = d.spec, v
			}
		}
	}
	if off != nil {
		return fmt.Errorf(
			"%s: type %s expands to at least %d nodes during type validation, "+
				"exceeding the limit of %d (possible denial-of-service vector)",
			fset.Position(off.Name.Pos()), off.Name.Name, offVal, typeExpansionBudget)
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
		// and small (measured max ~29 across all stdlibs), independent of input.
		//
		// So this only under-counts by a bounded per-reference constant, never
		// hides a fan-out. We deliberately do NOT fetch stdlib source: go/types
		// serves stdlib imports from its result cache without a store read, so
		// fetching here would add store gas the deploy otherwise never pays.
		// (Counting stdlibs exactly is possible via a table precomputed at stdlib
		// load — no per-deploy gas — but the cross-module plumbing isn't worth it
		// for a leaf that is already bounded-safe. See adr/pr4264_lint_transpile.md.)
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

// checkNoGenerics rejects go1.18 generics syntax before the package reaches
// go/types. Gno targets go1.17 and does not support generics (the interpreter
// rejects them too, e.g. go2gno.go), but the go/types pass runs first and would
// still form and validate generic types. That matters for the DoS defense: a
// generic instantiation fans out via type-argument substitution and interface
// type sets fan out via their union terms, both of which drive go/types'
// unmetered validType walk exponential. Pinning types.Config.GoVersion does NOT
// help — go/types reports the version error but does not halt, so it still runs
// the walk — hence this must reject syntactically, before go/types is invoked.
//
// "Generics" here means the two go1.18 constructs that carry a fan-out:
//   - type parameters, on a type or func declaration (`type W[P any] ...`); and
//   - interface type sets: unions (`A | B`) and approximation terms (`~T`).
//
// It reports the earliest-positioned offending construct, so the error is
// deterministic (the message can reach consensus-visible tx results).
func checkNoGenerics(fset *token.FileSet, gofs []*ast.File) error {
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
