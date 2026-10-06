package fork

import (
	"context"
	"fmt"

	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	coretypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
)

// rpcGenesisSource fetches the source chain's genesis via the /genesis RPC
// endpoint on one or more nodes. Endpoints are tried in order until one
// succeeds. Suitable when the source's /genesis is small enough to fit in
// a single response — for ~200 MB genesis files (e.g. gnoland1) /genesis
// is typically unavailable, prefer fileGenesisSource against a local copy.
type rpcGenesisSource struct {
	rpcURLs []string
	clients []*rpcclient.RPCClient
}

// newRPCGenesisSource opens one RPC client per URL parsed from rpcInput.
// The input may be a single URL or a comma-separated list of URLs for
// failover.
func newRPCGenesisSource(rpcInput string) (*rpcGenesisSource, error) {
	urls, err := parseRPCURLs(rpcInput)
	if err != nil {
		return nil, err
	}
	clients, err := openRPCClients(urls)
	if err != nil {
		return nil, err
	}
	return &rpcGenesisSource{rpcURLs: urls, clients: clients}, nil
}

func (s *rpcGenesisSource) Description() string {
	if len(s.clients) > 1 {
		return fmt.Sprintf("RPC (%d endpoints)", len(s.clients))
	}
	return "RPC"
}

func (s *rpcGenesisSource) Close() error {
	var firstErr error
	for _, c := range s.clients {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *rpcGenesisSource) FetchGenesis(ctx context.Context) (*bft.GenesisDoc, error) {
	res, err := tryEndpoints(s.clients, func(c *rpcclient.RPCClient) (*coretypes.ResultGenesis, error) {
		return c.Genesis(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("RPC genesis call: %w", err)
	}
	return res.Genesis, nil
}
