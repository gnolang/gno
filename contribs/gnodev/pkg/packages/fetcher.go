package packages

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// newRemoteFetcher returns the fetcher backing remote lookups. Remote
// fetching is opt-in per chain domain: with no -remote entries every fetch
// is refused and the loader stays filesystem-only; with entries, only the
// listed domains are fetched, each through its configured RPC.
func newRemoteFetcher(remotes map[string]string) pkgdownload.PackageFetcher {
	if len(remotes) == 0 {
		return disabledFetcher{}
	}
	return &domainFetcher{remotes: remotes, inner: rpcpkgfetcher.New(remotes)}
}

type disabledFetcher struct{}

func (disabledFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	return nil, fmt.Errorf("remote fetching is disabled, pass -remote <domain>=<rpc> to fetch %q from a chain", pkgPath)
}

type domainFetcher struct {
	remotes map[string]string
	inner   pkgdownload.PackageFetcher
}

func (f *domainFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	domain, _, _ := strings.Cut(pkgPath, "/")
	if _, ok := f.remotes[domain]; !ok {
		return nil, fmt.Errorf("no -remote configured for domain %q, refusing to fetch %q", domain, pkgPath)
	}
	return f.inner.FetchPackage(pkgPath)
}
