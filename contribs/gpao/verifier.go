package main

import (
	"fmt"
	"go/token"
	"io"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/test"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// verifier holds everything one verification needs: the disk-backed stores that
// resolve stdlib and examples/ imports, a typecheck cache, and an RPC fallback
// for packages that exist only on chain.
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
	// in which case such imports simply stay unresolved.
	rpc *rpcGetter

	// errw receives advisory notes. A child's stderr is captured by the parent
	// and surfaces as the rejection reason, so anything written here must be
	// worth an operator seeing.
	errw io.Writer
}

// newVerifier builds the stores for one verification.
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
	// Mirrors `gno lint`: production files against stdlibs + examples, with a
	// test-stdlib overlay. PreprocessOnly so imported code is preprocessed
	// rather than executed — we need type information, not side effects.
	prodbs, prodgs := test.StoreWithOptions(gnoRoot, errw,
		test.StoreOptions{PreprocessOnly: true, WithExamples: true})

	return &verifier{
		prodbs: prodbs, prodgs: prodgs,
		rpc:  rpc,
		errw: errw,
	}, nil
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
	// On-chain-only imports are left to the RPC fallback below and to
	// seedChainImports; anything still unresolved surfaces as a typecheck
	// error, or skips preprocess with a note. Inlined rather than left as a
	// separate step a caller has to remember: two of the tests used to omit it
	// and so exercised a sequence the child never runs.
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

	// The preprocessor resolves imports through the store, which has no RPC
	// fallback of its own -- unlike the typecheck above, which reaches the
	// chain through hybridGetter. Seed what the disk cannot supply, or a
	// package importing an on-chain-only dependency would fail preprocess
	// while type-checking fine: precisely the composable case "inert" exists
	// to allow.
	//
	// If something is still unresolvable, skip preprocess rather than reject.
	// Refusing would be a regression -- before this stage existed such a
	// package was approved on the typecheck alone -- and it would refuse for a
	// limitation of the oracle rather than a fault of the package. Silence is
	// the thing to avoid, so it is logged.
	if missing := v.seedChainImports(st, mpkg); missing != "" {
		fmt.Fprintf(v.errw,
			"gpao: %q imports %q, which this oracle cannot resolve; "+
				"approving on the typecheck alone, preprocess NOT measured\n",
			mpkg.Path, missing)
		return nil
	}

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

// prepare does the work verification needs that is not the package's cost:
// the disk imports, and the chain-domain import closure fetched into the RPC
// cache so that neither stage below asks the network again. The validator pays
// for the typecheck and the preprocess; it never pays for the oracle's node
// being slow. So this runs before the budget starts, and the budget then
// measures only what the validator will.
//
// A transport fault is returned as such: no verdict was possible. An import
// that nothing serves is not an error here; the typecheck judges that.
func (v *verifier) prepare(mpkg *std.MemPackage) error {
	_ = test.LoadImports(v.prodgs, mpkg, false)
	if v.rpc == nil {
		return nil
	}
	v.walkChainImports(mpkg, v.rpc.GetMemPackage)
	if v.rpc.transportErr != nil {
		return v.rpc.transportErr
	}
	return nil
}

// seedChainImports loads mpkg's chain-domain imports into st, transitively,
// fetching over RPC whatever the disk store lacks. It returns the first import
// path it could not resolve, or "" if everything resolved.
func (v *verifier) seedChainImports(st gno.Store, mpkg *std.MemPackage) string {
	return v.walkChainImports(mpkg, func(path string) *std.MemPackage {
		if dep := v.prodgs.GetMemPackage(path); dep != nil {
			return dep
		}
		if v.rpc == nil {
			return nil
		}
		dep := v.rpc.GetMemPackage(path)
		if dep != nil {
			st.AddMemPackage(dep, gno.MPUserProd)
		}
		return dep
	})
}

// walkChainImports visits mpkg's chain-domain imports transitively, resolving
// each through resolve, and returns the first path resolve could not supply, or
// "" if everything resolved.
//
// Transitive because a fetched dependency has imports of its own, and both the
// typechecker and the preprocessor need the whole graph present, not just the
// first level.
func (v *verifier) walkChainImports(mpkg *std.MemPackage, resolve func(path string) *std.MemPackage) string {
	seen := map[string]bool{mpkg.Path: true}
	queue := []*std.MemPackage{mpkg}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		imports, err := packages.Imports(cur, token.NewFileSet())
		if err != nil {
			// Unparseable source is the typecheck's verdict to give, not ours.
			return ""
		}
		for _, imp := range imports.Merge(packages.FileKindPackageSource) {
			path := imp.PkgPath
			if seen[path] || !strings.HasPrefix(path, chainDomainPrefix) {
				// Stdlibs and already-queued paths need nothing: the store
				// resolves stdlibs itself.
				continue
			}
			seen[path] = true

			dep := resolve(path)
			if dep == nil {
				return path
			}
			queue = append(queue, dep)
		}
	}
	return ""
}

// chainDomainPrefix is the import prefix that means "on this chain" rather than
// "a standard library".
const chainDomainPrefix = "gno.land/"
