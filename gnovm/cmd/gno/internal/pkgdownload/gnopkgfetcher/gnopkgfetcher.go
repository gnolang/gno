// Package gnopkgfetcher provides an implementation of [pkgdownload.PackageFetcher]
// to fetches packages from tm2 rpc endpoints
package gnopkgfetcher

import (
	"fmt"
	"path"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

type gnoPackageFetcher struct {
	remoteOverrides map[string]string
}

var _ pkgdownload.PackageFetcher = (*gnoPackageFetcher)(nil)

func New(remoteOverrides map[string]string) pkgdownload.PackageFetcher {
	return &gnoPackageFetcher{
		remoteOverrides: remoteOverrides,
	}
}

// FetchPackage implements [pkgdownload.PackageFetcher].
func (gpf *gnoPackageFetcher) FetchPackage(pkgPath string) ([]pkgdownload.PackageFile, error) {
	rpcURL, err := rpcURLFromPkgPath(pkgPath, gpf.remoteOverrides)
	if err != nil {
		return nil, fmt.Errorf("get rpc url for pkg path %q: %w", pkgPath, err)
	}

	client, err := client.NewHTTPClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate tm2 client with remote %q: %w", rpcURL, err)
	}
	defer client.Close()

	data, err := qfile(client, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("query files list for pkg %q: %w", pkgPath, err)
	}

	files := strings.Split(string(data), "\n")
	res := make([]pkgdownload.PackageFile, len(files))
	for i, file := range files {
		filePath := path.Join(pkgPath, file)
		data, err := qfile(client, filePath)
		if err != nil {
			return nil, fmt.Errorf("query package file %q: %w", filePath, err)
		}

		res[i] = pkgdownload.PackageFile{Name: file, Body: data}
	}
	return res, nil
}

func rpcURLFromPkgPath(pkgPath string, remoteOverrides map[string]string) (string, error) {
	parts := strings.Split(pkgPath, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("bad pkg path %q", pkgPath)
	}
	domain := parts[0]

	var rpcURL string
	if override, ok := remoteOverrides[domain]; ok {
		rpcURL = override
	} else {
		// XXX: retrieve host/port from r/sys/zones.
		rpcURL = fmt.Sprintf("https://rpc.%s:443", domain)
	}

	return rpcURL, nil
}

func qfile(c client.Client, pkgPath string) ([]byte, error) {
	path := "vm/qfile"
	data := []byte(pkgPath)

	qres, err := c.ABCIQuery(path, data)
	if err != nil {
		return nil, errors.Wrap(err, "query qfile")
	}
	if qres.Response.Error != nil {
		return nil, errors.Wrap(qres.Response.Error, "QFile failed: log:%s", qres.Response.Log)
	}

	return qres.Response.Data, nil
}