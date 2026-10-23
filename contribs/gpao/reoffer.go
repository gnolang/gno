package main

import (
	"context"
	"go/token"
	"maps"
	"slices"
	"time"

	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// maxReoffers caps how often one submission is verified again in a run,
// whatever left it pending.
const maxReoffers = 3

// reofferInterval is how often the verifier looks again at the submissions it
// left pending.
const reofferInterval = 10 * time.Minute

// strandedSubmission is a submission handleCandidate left pending: where it
// sits on the chain, never its bytes, which are read from its block again.
type strandedSubmission struct {
	height    int64
	offers    int
	blockedOn string // the import found not live by the last look, if any
}

// track records how handleCandidate left path. Pending keeps the submission
// for a later offer; anything else is a verdict and ends its tracking.
func (o *oracle) track(path string, height int64) {
	if o.status.get(path).Status != statusPending {
		delete(o.stranded, path)
		return
	}
	s := o.stranded[path]
	if s.height != height {
		s = strandedSubmission{height: height}
	}
	s.blockedOn = ""
	o.stranded[path] = s
}

// reofferStranded verifies again each submission this run left pending whose
// chain imports are all live now. Heights are read once, so nothing else
// reaches it again. It runs on the verifier between candidates and stops as
// soon as one is queued, so new blocks go first.
//
// What it keeps grows with the submissions this run left pending, as the
// status board does: one is dropped once the chain no longer holds it parked
// at its height, or after maxReoffers offers.
func (o *oracle) reofferStranded(ctx context.Context) {
	live := make(map[string]bool)
	for _, path := range slices.Sorted(maps.Keys(o.stranded)) {
		if ctx.Err() != nil || len(o.candidates) > 0 {
			return
		}
		o.reoffer(ctx, path, live)
	}
}

func (o *oracle) reoffer(ctx context.Context, path string, live map[string]bool) {
	s := o.stranded[path]
	if s.blockedOn != "" && !o.isLive(ctx, s.blockedOn, live) {
		return
	}
	meta, err := queryPkgMeta(ctx, o.client.RPCClient, path)
	if err != nil {
		o.errf("gpao: cannot read the status of %q, not offering it again yet: %v", path, err)
		return
	}
	// Enabled, by anyone, or replaced, and a replacement reaches the verifier
	// from its own block. A redeploy parked over a live private realm reports
	// the live package, so it lands here too.
	if meta.Status != vm.PackageStatusInert || int64(meta.Height) != s.height {
		delete(o.stranded, path)
		return
	}
	if s.offers >= maxReoffers {
		delete(o.stranded, path)
		o.errf("gpao: %q is still pending after %d more attempts, not offering it again this run",
			path, s.offers)
		return
	}
	c, err := o.submissionAt(ctx, s.height, path)
	if err != nil {
		o.errf("gpao: cannot read the submission of %q at height %d: %v", path, s.height, err)
		return
	}
	if c == nil {
		delete(o.stranded, path)
		return
	}
	if blocked := o.firstImportNotLive(ctx, c.mpkg, live); blocked != "" {
		s.blockedOn = blocked
		o.stranded[path] = s
		return
	}
	s.offers++
	o.stranded[path] = s
	o.logf("gpao: offering %q again, submitted at height %d", path, s.height)
	o.handleCandidate(ctx, *c)
}

// submissionAt returns the last package the block at height added at path,
// which is the one the chain holds when the block is the path's submitting
// height.
func (o *oracle) submissionAt(ctx context.Context, height int64, path string) (*candidate, error) {
	added, err := o.addedPackages(ctx, height)
	if err != nil {
		return nil, err
	}
	for _, c := range slices.Backward(added) {
		if c.mpkg.Path == path {
			return &c, nil
		}
	}
	return nil, nil
}

// firstImportNotLive returns the first chain package mpkg imports that is not
// live, or "" when a verification can reach a verdict. A live package's own
// imports are live, so only mpkg's are asked about.
func (o *oracle) firstImportNotLive(ctx context.Context, mpkg *std.MemPackage, live map[string]bool) string {
	imports, err := packages.Imports(mpkg, token.NewFileSet())
	if err != nil {
		return "" // unparseable source is the verification's verdict to give
	}
	for _, imp := range imports.Merge(packages.FileKindPackageSource) {
		if gno.IsUserlib(imp.PkgPath) && !o.isLive(ctx, imp.PkgPath, live) {
			return imp.PkgPath
		}
	}
	return ""
}

// isLive reports whether the chain runs a package at path, a live one with a
// submission parked over it included, and keeps the answer in live for the
// rest of the pass. A genesis package the import cache holds is live without
// asking. A path the node does not answer for counts as not live until the
// next pass.
func (o *oracle) isLive(ctx context.Context, path string, live map[string]bool) bool {
	if _, genesis := o.genesisImports()[path]; genesis {
		return true
	}
	if v, asked := live[path]; asked {
		return v
	}
	meta, err := queryPkgMeta(ctx, o.client.RPCClient, path)
	v := err == nil && meta.Status == vm.PackageStatusLive
	live[path] = v
	return v
}
