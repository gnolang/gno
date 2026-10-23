package main

import (
	"context"
	"fmt"
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
	blockedOn string // the import that held it back at the last look, if any
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

// reofferStranded verifies again each submission this run left pending that
// no import holds back any more. Heights are read once, so nothing else
// reaches it again. It runs on the verifier between candidates and stops as
// soon as one is queued, so new blocks go first. The next pass resumes where
// it stopped, so on a chain that keeps receiving candidates a path late in
// order is still reached.
//
// What it keeps grows with the submissions this run left pending, as the
// status board does: one is dropped once the chain no longer holds it parked
// at its height, or after maxReoffers offers.
func (o *oracle) reofferStranded(ctx context.Context) {
	statuses := make(map[string]string)
	paths := slices.Sorted(maps.Keys(o.stranded))
	start, _ := slices.BinarySearch(paths, o.reofferNext)
	o.reofferNext = ""
	for _, path := range slices.Concat(paths[start:], paths[:start]) {
		if ctx.Err() != nil {
			return
		}
		if len(o.candidates) > 0 {
			o.reofferNext = path
			return
		}
		o.reoffer(ctx, path, statuses)
	}
}

func (o *oracle) reoffer(ctx context.Context, path string, statuses map[string]string) {
	s := o.stranded[path]
	if s.blockedOn != "" && holdsBack(o.importStatus(ctx, s.blockedOn, statuses)) {
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
		reason := fmt.Sprintf("still pending after %d more attempts, not offered again this run", s.offers)
		o.status.record(path, statusGaveUp, reason, s.offers)
		o.errf("gpao: %q is %s", path, reason)
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
	if blocked := o.firstImportHoldingBack(ctx, c.mpkg, statuses); blocked != "" {
		s.blockedOn = blocked
		o.stranded[path] = s
		o.status.record(path, statusPending,
			fmt.Sprintf("waits on %q, which is not live; offered again once it is", blocked), s.offers)
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

// firstImportHoldingBack returns the first chain package mpkg imports that
// would leave it pending again, or "" when a verification can reach a verdict.
// A live package's own imports are live, so only mpkg's are asked about.
func (o *oracle) firstImportHoldingBack(ctx context.Context, mpkg *std.MemPackage, statuses map[string]string) string {
	imports, err := packages.Imports(mpkg, token.NewFileSet())
	if err != nil {
		return "" // unparseable source is the verification's verdict to give
	}
	for _, imp := range imports.Merge(packages.FileKindPackageSource) {
		if gno.IsUserlib(imp.PkgPath) && holdsBack(o.importStatus(ctx, imp.PkgPath, statuses)) {
			return imp.PkgPath
		}
	}
	return ""
}

// holdsBack reports whether an import the chain holds in this state leaves its
// importer pending: parked, or not answered for. A live import resolves, and
// an absent one is a rejection, which a verification gives.
func holdsBack(status string) bool {
	return status != vm.PackageStatusLive && status != vm.PackageStatusAbsent
}

// importStatus is what the chain holds at path, one of the vm.PackageStatus
// constants, or "" when the node did not answer, kept in statuses for the rest
// of the pass. A live package with a submission parked over it is live. A
// genesis package the import cache holds is live without asking.
func (o *oracle) importStatus(ctx context.Context, path string, statuses map[string]string) string {
	if _, genesis := o.genesisImports()[path]; genesis {
		return vm.PackageStatusLive
	}
	if status, asked := statuses[path]; asked {
		return status
	}
	var status string
	if meta, err := queryPkgMeta(ctx, o.client.RPCClient, path); err == nil {
		status = meta.Status
	}
	statuses[path] = status
	return status
}
