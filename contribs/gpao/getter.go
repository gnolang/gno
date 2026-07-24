package main

import (
	"context"
	"path"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// hybridGetter resolves imported packages for the typechecker: first from the
// local disk store (stdlibs + examples/), then, on a miss, from the chain over
// RPC. This lets the oracle typecheck packages that import other on-chain-only
// packages, matching the validator's view.
type hybridGetter struct {
	disk gno.MemPackageGetter
	rpc  *rpcGetter
}

func (h hybridGetter) GetMemPackage(pkgPath string) *std.MemPackage {
	if mpkg := h.disk.GetMemPackage(pkgPath); mpkg != nil {
		return mpkg
	}
	return h.rpc.GetMemPackage(pkgPath)
}

// qfileFunc runs a vm/qfile query for a package path or a package file path.
type qfileFunc func(filepath string) ([]byte, error)

// rpcGetter fetches package sources from a node via the vm/qfile ABCI query and
// reconstructs them into MemPackages. On-chain packages are immutable by path
// (a path is write-once — re-adding fails), so any successfully fetched package
// is cached for the lifetime of the oracle and never re-queried.
type rpcGetter struct {
	qfile qfileFunc
	cache map[string]*std.MemPackage
}

func newRPCGetter(client rpcclient.Client) *rpcGetter {
	qfile := func(filepath string) ([]byte, error) {
		qres, err := client.ABCIQuery(context.Background(), "vm/qfile", []byte(filepath))
		if err != nil {
			return nil, err
		}
		if qres.Response.Error != nil {
			return nil, qres.Response.Error
		}
		return qres.Response.Data, nil
	}
	return &rpcGetter{qfile: qfile, cache: make(map[string]*std.MemPackage)}
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
// query fails (the typechecker then reports the import as unresolved).
func (g *rpcGetter) fetch(pkgPath string) *std.MemPackage {
	list, err := g.qfile(pkgPath)
	if err != nil {
		return nil
	}
	names := strings.Split(string(list), "\n")
	files := make([]*std.MemFile, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		body, err := g.qfile(path.Join(pkgPath, name))
		if err != nil {
			return nil
		}
		files = append(files, &std.MemFile{Name: name, Body: string(body)})
	}
	if len(files) == 0 {
		return nil
	}

	return &std.MemPackage{
		Name:  packageName(files),
		Path:  pkgPath,
		Files: files,
		Type:  gno.MPUserProd,
	}
}

// packageName derives the package name from the first .gno file whose package
// clause parses. Returns "" if none do; the typechecker will then error out.
func packageName(files []*std.MemFile) string {
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".gno") {
			continue
		}
		if name, err := gno.PackageNameFromFileBody(f.Name, f.Body); err == nil {
			return string(name)
		}
	}
	return ""
}
