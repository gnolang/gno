package packages

import (
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
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
	qres, err := res.RPCClient.ABCIQuery(qpath, data)
	if err != nil {
		return nil, fmt.Errorf("client unable to query: %w", err)
	}

	if err := qres.Response.Error; err != nil {
		if errors.Is(err, vm.InvalidPkgPathError{}) ||
			strings.HasSuffix(err.Error(), "is not available") { // XXX: find a better to check this
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

		// Check package name
		if name == "" && isGnoFile(fname) && !isTestFile(fname) {
			// Check package name
			f, err := parser.ParseFile(fset, fname, body, parser.PackageClauseOnly)
			if err != nil {
				return nil, fmt.Errorf("unable to parse file %q: %w", fname, err)
			}
			name = f.Name.Name
		}

		memFiles = append(memFiles, &gnovm.MemFile{
			Name: fname, Body: string(body),
		})
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
