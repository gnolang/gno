package packages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	gopath "path"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type remoteResolver struct {
	*client.RPCClient
	name string
	fset *token.FileSet
}

func NewRemoteResolver(name string, cl *client.RPCClient) Resolver {
	return &remoteResolver{
		RPCClient: cl,
		name:      name,
		fset:      token.NewFileSet(),
	}
}

func (res *remoteResolver) Name() string {
	return fmt.Sprintf("remote<%s>", res.name)
}

func (res *remoteResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	const qpath = "vm/qfile"

	// First query files
	data := []byte(path)
	qres, err := res.RPCClient.ABCIQuery(context.Background(), qpath, data)
	if err != nil {
		return nil, fmt.Errorf("client unable to query: %w", err)
	}

	if err := qres.Response.Error; err != nil {
		if errors.Is(err, vm.InvalidFileError{}) ||
			errors.Is(err, vm.InvalidPkgPathError{}) ||
			errors.Is(err, vm.InvalidPackageError{}) {
			return nil, ErrResolverPackageNotFound
		}

		return nil, fmt.Errorf("querying %q error: %w", path, err)
	}

	var name string
	memFiles := []*std.MemFile{}
	files := bytes.Split(qres.Response.Data, []byte{'\n'})
	for _, filename := range files {
		fname := string(filename)
		fpath := gopath.Join(path, fname)
		qres, err := res.RPCClient.ABCIQuery(context.Background(), qpath, []byte(fpath))
		if err != nil {
			return nil, fmt.Errorf("unable to query path")
		}

		if err := qres.Response.Error; err != nil {
			return nil, fmt.Errorf("unable to query file %q on path %q: %w", fname, path, err)
		}
		body := qres.Response.Data

		// Check package name
		if name == "" && isGnoFile(fname) {
			// Check package name
			f, err := parser.ParseFile(fset, fname, body, parser.PackageClauseOnly)
			if err != nil {
				return nil, fmt.Errorf("unable to parse file %q: %w", fname, err)
			}
			name = f.Name.Name
		}

		memFiles = append(memFiles, &std.MemFile{
			Name: fname, Body: string(body),
		})
	}

	return &Package{
		MemPackage: std.MemPackage{
			Name:  name,
			Path:  path,
			Files: memFiles,
		},
		Kind:     PackageKindRemote,
		Location: path,
	}, nil
}
