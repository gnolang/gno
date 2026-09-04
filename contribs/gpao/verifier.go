package main

import (
	"fmt"
	"go/token"
	"io"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/test"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// verifier holds everything one verification needs: stores that resolve stdlib
// and examples/ imports from the local filesystem, and an RPC fallback for
// packages that exist only on chain.
//
// "Local" rather than "persistent": the stores are backed by a memdb built at
// construction (see test.StoreWithOptions), so every package they materialize
// is read from gnoRoot and preprocessed again for the next verification.
//
// It lives in the child process (see verifyone.go), one per verification. That
// is what makes the budget enforceable. Previously these stores hung off the
// long-lived daemon and were reused across candidates, so an abandoned attempt
// kept mutating state the next one read — which is precisely why the budget
// could only be measured rather than enforced. A process boundary gives
// per-attempt isolation for free, so there is nothing left to race and a
// deadline can simply kill the work.
type verifier struct {
	prodbs storetypes.CommitStore
	prodgs gno.Store

	// rpc resolves on-chain-only imports the disk store cannot (falls back to
	// vm/qfile queries against the watched node). Nil when no remote is given,
	// in which case such an import resolves nowhere and fails the typecheck.
	rpc *rpcGetter

	// errw takes notes an operator needs on the SUCCESS path -- currently a
	// dependency the oracle could not prebuild, which points at its own tree
	// rather than at the candidate. The parent tees a child's stderr for
	// exactly this, so anything written here is seen.
	errw io.Writer
}

// newVerifier builds the stores for one verification. errw takes the store's
// own diagnostics and this verifier's; the parent tees a child's stderr, so
// anything landing there is seen by an operator.
//
// Deliberately no signer and no keystore: a verifier handles untrusted input
// and cannot approve anything, so the approver key never enters the process
// that compiles a stranger's code.
func newVerifier(gnoRoot, remote string, errw io.Writer) (*verifier, error) {
	var rpc *rpcGetter
	if remote != "" {
		c, err := rpcclient.NewHTTPClient(remote)
		if err != nil {
			return nil, fmt.Errorf("build RPC client: %w", err)
		}
		rpc = newRPCGetter(c)
	}
	// Production files against two directories under gnoRoot: gnovm/stdlibs
	// and examples. No test-stdlib overlay and the production native resolver,
	// which is right for this daemon -- the chain does not evaluate test files
	// at enable, so a test-only definition must not resolve here either.
	//
	// PreprocessOnly so imported code is preprocessed rather than executed: we
	// need type information, not side effects.
	prodbs, prodgs := test.StoreWithOptions(gnoRoot, errw,
		test.StoreOptions{PreprocessOnly: true, WithExamples: true})

	v := &verifier{
		prodbs: prodbs, prodgs: prodgs,
		rpc:  rpc,
		errw: errw,
	}
	// On the base store, so every transaction begun from it inherits the
	// getter: BeginTransaction copies pkgGetter. prepare materializes the chain
	// closure through it before the budget starts, and the verification then
	// finds those packages already built.
	v.injectChainGetter(prodgs)
	return v, nil
}

// injectChainGetter extends st's package getter with the chain: on a miss for a
// path that lives on this chain, fetch the source over RPC and preprocess it
// into st. That is what the disk getter does for an examples/ import, so the
// preprocessor sees one uniform way to resolve.
//
// Preprocessed, never run: init() stays unexecuted, as it does for the
// candidate itself. Saved, though: GetPackage caches a getter's result in the
// calling store alone, and prepare's work has to be visible to the
// verification's own transaction, which reads through the base store.
func (v *verifier) injectChainGetter(st gno.Store) {
	disk := st.GetPackageGetter()
	st.SetPackageGetter(func(pkgPath string, store gno.Store) (
		*gno.PackageNode, *gno.PackageValue,
	) {
		if disk != nil {
			if pn, pv := disk(pkgPath, store); pv != nil {
				return pn, pv
			}
		}
		// IsUserlib, the same predicate hybridGetter routes the typecheck on:
		// refusing anything the typecheck resolved fails preprocess with
		// "unknown import path", which gpao reports as bad code.
		if v.rpc == nil || !gno.IsUserlib(pkgPath) {
			return nil, nil
		}
		dep := v.rpc.GetMemPackage(pkgPath)
		if dep == nil {
			return nil, nil
		}
		// Under the dependency's OWN declared type, which AddMemPackage
		// requires to match.
		mptype, ok := dep.Type.(gno.MemPackageType)
		if !ok {
			return nil, nil
		}
		store.AddMemPackage(dep, mptype)
		m2 := gno.NewMachineWithOptions(gno.MachineOptions{
			PkgPath:     pkgPath,
			Output:      io.Discard,
			Store:       store,
			SkipPackage: true,
		})
		defer m2.Release()
		return m2.PreprocessFiles(dep.Name, dep.Path,
			m2.ParseMemPackageAsType(dep, gno.MPUserProd), true, false)
	})
}

// verifyPackage typechecks and then preprocesses mpkg, mirroring the two stages
// the validator re-runs at MsgEnablePackage time. Both matter: a package that
// typechecks quickly but preprocesses slowly costs the chain just as much, and
// checking only the first would leave that half unmeasured.
//
// Any panic from either stage becomes an error, so one bad package cannot take
// the daemon down -- and the preprocessor reports errors by panicking, so this
// is its error path rather than only a safety net.
func (v *verifier) verifyPackage(mpkg *std.MemPackage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("verification panicked: %v", r)
		}
		// A failure reached while the resolver under it was failing is not a
		// verdict: the unresolved import may exist and be unfetchable.
		if err != nil && v.rpc != nil && v.rpc.transportErr != nil {
			err = fmt.Errorf("%w (verification said: %w)", v.rpc.transportErr, err)
		}
	}()

	// Best-effort preload of imports resolvable from disk (stdlibs, examples).
	// On-chain-only imports are left to the RPC fallback below, and to the
	// package getter at preprocess; anything unresolvable surfaces as a
	// typecheck error. Inlined rather than left as a separate step a caller has
	// to remember: two of the tests used to omit it and so exercised a sequence
	// the child never runs.
	_ = test.LoadImports(v.prodgs, mpkg, false)

	// Fresh transactions isolate each typecheck from the persistent base stores.
	newProdGnoStore := func() gno.Store {
		cw := v.prodbs.CacheWrap()
		return v.prodgs.BeginTransaction(cw, cw, nil, nil)
	}
	// Wrap the disk getters with an RPC fallback so imports of on-chain-only
	// packages (not present under examples/) still resolve.
	// Options match EnablePackage's exactly (see VMKeeper.EnablePackage):
	// Getter only, ProdOnly, TCLatestStrict. That is the point — gpao exists to
	// predict what the validator will do, so any divergence here is a way to
	// approve something the chain then rejects, or to reject something it would
	// have accepted. In particular there is no TestGetter: the chain never
	// evaluates test files at enable, so a test-file error must not be a
	// verdict here either.
	//
	// No cache, either. The permanent typecheck cache only pays off across
	// calls, and this process makes exactly one.
	if _, errs := gno.TypeCheckMemPackage(mpkg, gno.TypeCheckOptions{
		Getter:   hybridGetter{disk: newProdGnoStore(), rpc: v.rpc},
		Mode:     gno.TCLatestStrict,
		ProdOnly: true,
	}); errs != nil {
		return errs
	}

	return v.preprocess(mpkg)
}

// preprocess runs the preprocessor over mpkg's production files.
//
// This is the second half of what the validator does at MsgEnablePackage, and
// PreprocessFiles is the right primitive for it. RunMemPackage is not: it
// executes init(), which off-chain and unmetered means a hostile package could
// hang this daemon forever.
//
// Not RunMemPackage, and not the shared store either: the transaction store is
// discarded, so nothing a candidate does here is persisted or visible to the
// next one.
func (v *verifier) preprocess(mpkg *std.MemPackage) error {
	cw := v.prodbs.CacheWrap()
	st := v.prodgs.BeginTransaction(cw, cw, nil, nil)

	// Chain imports resolve through the getter newVerifier installed on the
	// base store, which this transaction inherits. prepare has already
	// materialized them, so this stage normally finds them in the store and the
	// getter is a fallback. AddMemPackage is NOT the seam for any of it: an
	// import resolves to a PackageValue, and storing source builds none.

	tm := test.Machine(st, io.Discard, mpkg.Path, false, nil)
	defer tm.Release()
	// Add under the package's OWN declared type. AddMemPackage requires the
	// two to agree, so hardcoding either constant is wrong: production sends
	// MPUserAll (what AddPackage stamps and EnablePackage reads back), and a
	// package legitimately containing _test.gno files can only be MPUserAll —
	// MPUserProd's validation rejects those outright.
	//
	// Restricting to production files is the FILESET's job just below, which
	// is what the chain actually preprocesses.
	mptype, ok := mpkg.Type.(gno.MemPackageType)
	if !ok {
		return fmt.Errorf("mempackage %q has no type", mpkg.Path)
	}
	st.AddMemPackage(mpkg, mptype)
	// MPUserProd, so production files only -- matching the chain, which
	// type-checks and runs the prod subset (TypeCheckOptions.ProdOnly).
	tm.PreprocessFiles(mpkg.Name, mpkg.Path,
		tm.ParseMemPackageAsType(mpkg, gno.MPUserProd), false, false)
	return nil
}

// prepare does the work verification needs that is not the package's cost: the
// disk imports, and the chain import closure, fetched into the RPC cache and
// built into the store so that neither stage below asks the network or compiles
// a dependency again. The validator pays for the candidate's typecheck and
// preprocess; it never pays for the oracle's node being slow, nor for building
// dependencies -- at enable time every active package's value is already in its
// store. So this runs before the budget starts, and the budget then measures
// only what the validator will.
//
// A transport fault is returned as such: no verdict was possible. An import
// that nothing serves is not an error here; the typecheck judges that.
func (v *verifier) prepare(mpkg *std.MemPackage) error {
	_ = test.LoadImports(v.prodgs, mpkg, false)
	if v.rpc == nil {
		return nil
	}
	v.buildChainImports(mpkg)
	if v.rpc.transportErr != nil {
		return v.rpc.transportErr
	}
	return nil
}

// buildChainImports fetches mpkg's chain import closure into the RPC cache and
// builds each member into the base store, through the getter newVerifier
// installed there.
//
// Transitive because a fetched dependency has imports of its own, and both the
// typechecker and the preprocessor need the whole graph present, not just the
// first level.
//
// Best-effort per package, like the disk preload before it: a dependency that
// cannot be built is left for the verification to trip over, where the failure
// gets a verdict and a reason. Refusing here would turn it into an exit status
// the parent reads as unavailability, retried forever. The note is worth an
// operator's attention either way -- a dependency live on chain has already
// been type-checked, preprocessed and init()-run by a validator, so a build
// failure here points at this oracle's own tree.
//
// A path nothing serves is skipped rather than ending the walk: it is the
// typecheck's verdict to give, and the rest of the closure still has to be
// fetched, or the typecheck goes to the network on the clock.
func (v *verifier) buildChainImports(mpkg *std.MemPackage) {
	seen := map[string]bool{mpkg.Path: true}
	queue := []*std.MemPackage{mpkg}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		imports, err := packages.Imports(cur, token.NewFileSet())
		if err != nil {
			// Unparseable source is the typecheck's verdict to give, not ours.
			return
		}
		for _, imp := range imports.Merge(packages.FileKindPackageSource) {
			path := imp.PkgPath
			if seen[path] || !gno.IsUserlib(path) {
				// Stdlibs and already-queued paths need nothing: the store
				// resolves stdlibs itself. IsUserlib is what hybridGetter
				// routes the typecheck on and what the preprocess getter asks
				// too, so what is fetched here is exactly what they will ask
				// for -- a path either stage would fetch but this walk skipped
				// would reach the network inside the budget.
				continue
			}
			seen[path] = true

			// Fetch first: the source is what the walk descends through, and a
			// transport fault lands on the getter for prepare to read.
			dep := v.rpc.GetMemPackage(path)
			if dep == nil {
				continue
			}
			v.buildOne(path, cur.Path)
			queue = append(queue, dep)
		}
	}
}

// buildOne builds one fetched dependency into the base store, recovering so a
// dependency that will not compile costs the rest of the closure nothing.
func (v *verifier) buildOne(path, importedBy string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(v.errw, "gpao: could not prebuild %q, imported by %q: %v\n",
				path, importedBy, r)
		}
	}()
	v.prodgs.GetPackage(path, false)
}
