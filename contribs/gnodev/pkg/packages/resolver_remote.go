package packages

import (
	"bytes"
	"errors"
	"fmt"
	"go/token"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

type remoteResolver struct {
	*client.RPCClient // Root folder
	fset              *token.FileSet
}

func NewRemoteResolver(cl *client.RPCClient) Resolver {
	return &remoteResolver{
		RPCClient: cl,
		fset:      token.NewFileSet(),
	}
}

func (res *remoteResolver) Name() string {
	return fmt.Sprintf("remote")
}

func (res *remoteResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	const qpath = "vm/qfile"

	// First query files
	data := []byte(path)
	qres, err := res.RPCClient.ABCIQuery(qpath, data)
	if err != nil {
		return nil, fmt.Errorf("client unable to query: %w", err)
	}

	if err := qres.Response.Error; err != nil {
		if errors.Is(err, vm.InvalidPkgPathError{}) {
			return nil, ErrResolverPackageNotFound
		}

		return nil, fmt.Errorf("querying %q error: %w", path, err)
	}

	var name string
	memFiles := []*gnovm.MemFile{}
	files := bytes.Split(qres.Response.Data, []byte{'\n'})
	for _, filename := range files {
		fname := string(filename)
		fpath := filepath.Join(path, fname)
		qres, err := res.RPCClient.ABCIQuery(qpath, []byte(fpath))
		if err != nil {
			return nil, fmt.Errorf("unable to query path")
		}

		if err := qres.Response.Error; err != nil {
			return nil, fmt.Errorf("unable to query file %q on path %q: %w", fname, path, err)
		}

		body := qres.Response.Data
		memfile, pkgname, err := parseFile(fset, fname, body)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", fname, err)
		}

		if name != "" && name != pkgname {
			return nil, fmt.Errorf("conflict package name between %q and %q", name, memfile.Name)
		}

		name = pkgname
		memFiles = append(memFiles, memfile)
	}

	return &Package{
		MemPackage: gnovm.MemPackage{
			Name:  name,
			Path:  path,
			Files: memFiles,
		},
		Kind:     PackageKindRemote,
		Location: path,
	}, nil
}
