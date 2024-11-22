package gnopkgfetch

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoimports"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	tm2client "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"golang.org/x/mod/module"
)

// FetchPackageImportsRecursively recursively fetches the imports of a local package while following a given gno.mod replace directives
func FetchPackageImportsRecursively(io commands.IO, pkgDir string, gnoMod *gnomod.File) error {
	imports, err := gnoimports.PackageImports(pkgDir)
	if err != nil {
		return fmt.Errorf("read imports at %q: %w", pkgDir, err)
	}

	for _, pkgPath := range imports {
		resolved := gnoMod.Resolve(module.Version{Path: pkgPath})
		resolvedPkgPath := resolved.Path

		if !gnolang.IsRemotePkgPath(resolvedPkgPath) {
			continue
		}

		depDir := gnomod.PackageDir("", module.Version{Path: resolvedPkgPath})

		if err := fetchPackage(io, resolvedPkgPath, depDir); err != nil {
			return fmt.Errorf("fetch import %q of %q: %w", resolvedPkgPath, pkgDir, err)
		}

		if err := FetchPackageImportsRecursively(io, depDir, gnoMod); err != nil {
			return err
		}
	}

	return nil
}

// fetchPackage downloads a remote gno package by pkg path and store it at dst
func fetchPackage(io commands.IO, pkgPath string, dst string) error {
	modFilePath := filepath.Join(dst, "gno.mod")

	if _, err := os.Stat(modFilePath); err == nil {
		// modfile exists in modcache, do nothing
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	io.ErrPrintfln("gno: downloading %s", pkgPath)

	client := fetchClient
	if client == nil {
		// create client from pkgpath
		parts := strings.Split(pkgPath, "/")
		if len(parts) < 1 {
			return fmt.Errorf("bad pkg path %q", pkgPath)
		}
		// XXX: retrieve host/port from r/sys/zones.
		rpcURL := (&url.URL{
			Scheme: "https",
			Host:   "rpc." + parts[0] + ":443",
		}).String()
		tmClient, err := tm2client.NewHTTPClient(rpcURL)
		if err != nil {
			return fmt.Errorf("failed to instantiate tm2 client with remote %q: %w", rpcURL, err)
		}
		defer tmClient.Close()
		client = tmClient
	}

	// fetch files
	data, err := qfile(client, pkgPath)
	if err != nil {
		return fmt.Errorf("failed to query files list for pkg %q: %w", pkgPath, err)
	}
	if err := os.MkdirAll(dst, 0o744); err != nil {
		return fmt.Errorf("failed to create cache dir for %q at %q: %w", pkgPath, dst, err)
	}
	files := strings.Split(string(data), "\n")
	for _, file := range files {
		filePath := path.Join(pkgPath, file)
		data, err := qfile(client, filePath)
		if err != nil {
			return fmt.Errorf("failed to query package file %q: %w", filePath, err)
		}
		dst := filepath.Join(dst, file)
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("failed to write file at %q: %w", dst, err)
		}
	}

	// We need to write a marker file for each downloaded package.
	// For example: if you first download gno.land/r/foo/bar then download gno.land/r/foo,
	// we need to know that gno.land/r/foo is not downloaded.
	// We do this by checking for the presence of gno.land/r/foo/gno.mod
	if err := os.WriteFile(modFilePath, []byte("module "+pkgPath+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write modfile at %q: %w", modFilePath, err)
	}

	return nil
}

// not using gno client due to cyclic dep
func qfile(tmClient tm2client.Client, pkgPath string) ([]byte, error) {
	path := "vm/qfile"
	data := []byte(pkgPath)

	qres, err := tmClient.ABCIQuery(path, data)
	if err != nil {
		return nil, errors.Wrap(err, "query qfile")
	}
	if qres.Response.Error != nil {
		return nil, errors.Wrap(qres.Response.Error, "QFile failed: log:%s", qres.Response.Log)
	}

	return qres.Response.Data, nil
}

var fetchClient tm2client.Client
