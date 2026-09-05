package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// hybridGetter resolves imported packages for the typechecker, taking each kind
// of import from wherever the chain would take it.
//
// Stdlibs come from local disk. They ship with the binary and are not user
// state, so disk is the only place to get them and the only place that is right.
//
// Everything else -- /p/ and /r/ packages -- comes from the chain. This is the
// point of the whole daemon: the verdict has to describe what the validator will
// see when it runs the enable, and the validator resolves imports from chain
// state.
//
// Disk is NOT consulted for those, even as a fallback, when a remote is
// configured. It used to be tried first, which meant a package importing
// something present in the operator's examples/ but absent from the chain
// verified clean and got approved -- and then failed its own type-check at enable
// time, burning a fee and marking the path rejected for a fault that was the
// operator's local tree, not the code. Where the two agree the answer is the
// same; where they disagree the chain is the one that matters.
//
// With no remote there is nothing to ask, so disk is used for everything. That
// is a development mode, and the verdict then describes the operator's tree.
type hybridGetter struct {
	disk gno.MemPackageGetter
	rpc  *rpcGetter
}

func (h hybridGetter) GetMemPackage(pkgPath string) *std.MemPackage {
	// No remote: nothing to ask, so disk answers everything. Also guards the
	// nil receiver below, which would panic -- a crash rather than a verdict.
	if h.rpc == nil {
		return h.disk.GetMemPackage(pkgPath)
	}
	if gno.IsUserlib(pkgPath) {
		return h.rpc.GetMemPackage(pkgPath)
	}
	return h.disk.GetMemPackage(pkgPath)
}

// errResolverUnavailable reports that an import could not be resolved because
// the node could not be ASKED -- as opposed to the node answering that nothing
// is stored at the path. The distinction is the whole triage: the second is
// evidence about the package, the first is evidence about the operator's
// network, and only evidence about the package may become a verdict.
var errResolverUnavailable = errors.New("import resolver unavailable")

// qfileFunc runs a vm/qfile query for a package path or a package file path.
type qfileFunc func(filepath string) ([]byte, error)

// rpcGetter fetches package sources from a node via the vm/qfile ABCI query and
// reconstructs them into MemPackages. On-chain packages are immutable by path
// (a path is write-once — re-adding fails), so any successfully fetched package
// is cached for the lifetime of the oracle and never re-queried.
type rpcGetter struct {
	qfile qfileFunc
	cache map[string]*std.MemPackage

	// transportErr is the first transport fault seen this verification; the
	// child makes exactly one.
	transportErr error
}

func newRPCGetter(client rpcclient.Client) *rpcGetter {
	qfile := func(filepath string) ([]byte, error) {
		qres, err := client.ABCIQuery(context.Background(), "vm/qfile", []byte(filepath))
		if err != nil {
			// The node could not be reached at all.
			return nil, fmt.Errorf("%w: %w", errResolverUnavailable, err)
		}
		if qerr := qres.Response.Error; qerr != nil {
			if absence(qerr) {
				return nil, qerr
			}
			// The node was reached and answered about itself, not the path.
			return nil, fmt.Errorf("%w: %w", errResolverUnavailable, qerr)
		}
		return qres.Response.Data, nil
	}
	return &rpcGetter{qfile: qfile, cache: make(map[string]*std.MemPackage)}
}

// absence reports whether an answered query said "nothing is stored at this
// path", as opposed to saying something about the node.
//
// vm/qfile reports a genuine miss as exactly two types (VMKeeper.QueryFile).
// Anything else the node answers with -- an internal error from a node that is
// pruned, restarting or replaying and cannot load state at the height; an
// unknown request from a build without the route -- is the node describing
// itself, and is no more evidence about the import than an unreachable node is.
//
// Keyed on the type, not the text: the ABCI layer unwraps to the cause before
// the answer leaves the node, so what arrives carries only the static "file is
// not available" / "invalid package" and no path.
func absence(err abci.Error) bool {
	switch err.(type) {
	case vm.InvalidFileError, *vm.InvalidFileError,
		vm.InvalidPackageError, *vm.InvalidPackageError:
		return true
	}
	return false
}

func (g *rpcGetter) GetMemPackage(pkgPath string) *std.MemPackage {
	if mpkg, ok := g.cache[pkgPath]; ok {
		return mpkg
	}
	mpkg := g.fetch(pkgPath)
	// Cache only what the chain actually returned. Misses are NOT cached: a
	// package that is absent now (e.g. still inert, or enabled later in this
	// run) must resolve on a later query rather than being pinned to nil.
	if mpkg != nil {
		g.cache[pkgPath] = mpkg
	}
	return mpkg
}

// fetch queries vm/qfile for the package's file list, then each file's body,
// and assembles a MemPackage. Returns nil if the package is not on-chain or any
// query fails (the typechecker then reports the import as unresolved); a
// transport failure is additionally recorded in transportErr, because it is not
// evidence about the import.
func (g *rpcGetter) fetch(pkgPath string) *std.MemPackage {
	list, err := g.query(pkgPath)
	if err != nil {
		return nil
	}
	names := strings.Split(string(list), "\n")
	files := make([]*std.MemFile, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		body, err := g.query(path.Join(pkgPath, name))
		if err != nil {
			return nil
		}
		files = append(files, &std.MemFile{Name: name, Body: string(body)})
	}
	if len(files) == 0 {
		return nil
	}

	// MPUserAll, matching what AddPackage stamps on the stored package and
	// therefore what a node serves here. Not MPUserProd: qfile lists every
	// stored file, _test.gno and flattened _filetest.gno included (see
	// TestRPCGetterNamesReconstructedPackage), and MPUserProd's validation
	// rejects those outright -- so stamping it made AddMemPackage panic on any
	// dependency that ships tests, which is most of them. Restricting to
	// production files is the fileset's job at each point of use.
	return &std.MemPackage{
		Name:  packageName(files),
		Path:  pkgPath,
		Files: files,
		Type:  gno.MPUserAll,
	}
}

// query wraps qfile and remembers a transport fault, which is evidence
// about the operator's network rather than about the import.
func (g *rpcGetter) query(filepath string) ([]byte, error) {
	body, err := g.qfile(filepath)
	if err != nil && errors.Is(err, errResolverUnavailable) && g.transportErr == nil {
		g.transportErr = err
	}
	return body, err
}

// packageName derives the package name the way the chain derives it in
// ReadMemPackageFromList: the first production .gno file whose package clause
// parses names the package, with a _test suffix trimmed. Filetests carry
// arbitrary clauses, so they name a package only when it holds nothing else --
// their own clause where they agree on one, and the literal "filetests" where
// they do not.
//
// The order matters here because the file list comes from vm/qfile sorted, and
// ReadMemPackage flattens a package's filetests/ directory into it, so a
// filetest carrying a package clause of its own routinely sorts ahead of the
// production files. The fallback matters because the name is what
// typeCheckMemPackage writes .gnobuiltins.gno under, and an empty one makes
// that file unparseable.
func packageName(files []*std.MemFile) string {
	var filetestName string
	filetestsDiffer := false
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".gno") {
			continue
		}
		name, err := gno.PackageNameFromFileBody(f.Name, f.Body)
		if err != nil {
			continue
		}
		if !strings.HasSuffix(f.Name, "_filetest.gno") {
			return strings.TrimSuffix(string(name), "_test")
		}
		if filetestName == "" {
			filetestName = string(name)
		} else if filetestName != string(name) {
			filetestsDiffer = true
		}
	}
	if filetestsDiffer {
		return "filetests"
	}
	return filetestName
}
