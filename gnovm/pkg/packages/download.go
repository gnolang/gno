package packages

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnofiles"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func DownloadModule(io commands.IO, pkgPath string, dst string) error {
	modFilePath := filepath.Join(dst, gnofiles.ModfileName)
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		io.ErrPrintfln("gno: downloading %s", pkgPath)

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
		tmClient, err := client.NewHTTPClient(rpcURL)
		if err != nil {
			return fmt.Errorf("failed to instantiate tm2 client with remote %q: %w", rpcURL, err)
		}
		defer tmClient.Close()

		// fetch files
		data, err := qfile(tmClient, pkgPath)
		if err != nil {
			return fmt.Errorf("failed to query files list for pkg %q: %w", pkgPath, err)
		}
		if err := os.MkdirAll(dst, 0o744); err != nil {
			return fmt.Errorf("failed to create cache dir for %q at %q: %w", pkgPath, dst, err)
		}
		files := strings.Split(string(data), "\n")
		for _, file := range files {
			filePath := path.Join(pkgPath, file)
			data, err := qfile(tmClient, filePath)
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
	} else if err != nil {
		return fmt.Errorf("failed to stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	// modfile exists in modcache, do nothing

	return nil
}

// not using gno client due to cyclic dep
func qfile(tmClient client.Client, pkgPath string) ([]byte, error) {
	path := "vm/qfile"
	data := []byte(pkgPath)

	qres, err := tmClient.ABCIQuery(path, data)
	if err != nil {
		return nil, fmt.Errorf("query qfile: %w", err)
	}
	if qres.Response.Error != nil {
		return nil, fmt.Errorf("qfile failed: %w\nlog:\n%s", qres.Response.Error, qres.Response.Log)
	}

	return qres.Response.Data, nil
}
