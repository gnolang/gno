package gnomodfetch

import (
	"fmt"
	"go/parser"
	"go/token"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/load"
	tm2client "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"golang.org/x/mod/module"
)

func FetchPackage(io commands.IO, pkgPath string, dst string) error {
	modFilePath := filepath.Join(dst, "gno.mod")

	if _, err := os.Stat(modFilePath); err == nil {
		// modfile exists in modcache, do nothing
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	io.ErrPrintfln("gno: downloading %s", pkgPath)

	client := Client
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

	// write gno.mod
	if err := os.WriteFile(modFilePath, []byte("module "+pkgPath+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write modfile at %q: %w", modFilePath, err)
	}

	return nil
}

func FetchPackagesRecursively(io commands.IO, pkgPath string, gnoMod *gnomod.File) error {
	dst := filepath.Join(gnoenv.HomeDir(), "pkg", "mod", pkgPath)

	modFilePath := filepath.Join(dst, "gno.mod")

	if _, err := os.Stat(modFilePath); err == nil {
		// modfile exists in modcache, do nothing
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	if err := FetchPackage(io, pkgPath, dst); err != nil {
		return err
	}

	gnoFiles, err := load.GnoFilesFromArgs([]string{dst})
	if err != nil {
		return err
	}

	for _, f := range gnoFiles {
		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}

		for _, imp := range parsed.Imports {
			importPkgPath := strings.TrimPrefix(strings.TrimSuffix(imp.Path.Value, "\""), "\"")

			if !strings.ContainsRune(importPkgPath, '.') {
				// std lib, ignore
				continue
			}

			resolved := gnoMod.Resolve(module.Version{Path: importPkgPath})
			resolvedPkgPath := resolved.Path

			// TODO: don't fetch local

			if err := FetchPackagesRecursively(io, resolvedPkgPath, gnoMod); err != nil {
				return fmt.Errorf("fetch: %w", err)
			}
		}
	}

	return nil
}

var Client tm2client.Client

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
