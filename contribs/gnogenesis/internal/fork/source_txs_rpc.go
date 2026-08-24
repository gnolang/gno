package fork

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	coretypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// rpcFetcherBackoff is the delay between full-cycle retries when every
// endpoint fails for one height. Kept generous to weather brief 5xx storms.
const rpcFetcherBackoff = 500 * time.Millisecond

// rpcFetcherMaxCycles caps how many full endpoint-cycles a single height
// retries before yielding a terminal error.
const rpcFetcherMaxCycles = 3

// defaultWorkersPerEndpoint is the default number of in-flight block fetches
// per RPC endpoint when streaming the historical tx range.
const defaultWorkersPerEndpoint = 4

// errNoEndpoints is returned by tryEndpoints when given an empty client slice.
var errNoEndpoints = errors.New("no RPC clients configured")

// rpcTxsSource fetches blocks + ABCI responses from one or more live (or
// recently-halted) RPC endpoints. Single-shot calls (status, auth account
// query) try endpoints in order until one succeeds; the block range fetch
// uses a concurrent worker pool with per-endpoint semaphores (see
// pooledFetcher).
//
// rpcURLs and clients are parallel slices and immutable after construction;
// the goroutines spawned by FetchTxs read them concurrently without locking.
type rpcTxsSource struct {
	rpcURLs            []string
	clients            []*rpcclient.RPCClient
	workersPerEndpoint int
}

// newRPCTxsSource opens one RPC client per URL parsed from rpcInput. The
// input may be a single URL or a comma-separated list of URLs for parallel
// fetch + failover. workersPerEndpoint <= 0 falls back to
// defaultWorkersPerEndpoint.
func newRPCTxsSource(rpcInput string, workersPerEndpoint int) (*rpcTxsSource, error) {
	urls, err := parseRPCURLs(rpcInput)
	if err != nil {
		return nil, err
	}
	if workersPerEndpoint <= 0 {
		workersPerEndpoint = defaultWorkersPerEndpoint
	}
	clients, err := openRPCClients(urls)
	if err != nil {
		return nil, err
	}
	return &rpcTxsSource{
		rpcURLs:            urls,
		clients:            clients,
		workersPerEndpoint: workersPerEndpoint,
	}, nil
}

func (s *rpcTxsSource) Description() string {
	if len(s.clients) > 1 {
		return fmt.Sprintf("RPC (%d endpoints)", len(s.clients))
	}
	return "RPC"
}

func (s *rpcTxsSource) Close() error {
	var firstErr error
	for _, c := range s.clients {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *rpcTxsSource) LatestHeight(ctx context.Context) (int64, error) {
	res, err := tryEndpoints(s.clients, func(c *rpcclient.RPCClient) (*coretypes.ResultStatus, error) {
		return c.Status(ctx, nil)
	})
	if err != nil {
		return 0, fmt.Errorf("RPC status call: %w", err)
	}
	return res.SyncInfo.LatestBlockHeight, nil
}

// FetchTxs walks the requested height range via the pooled fetcher (parallel
// over endpoints, retried with backoff) and feeds each block through the
// shared sequence-resolution pipeline. chainID is supplied by the caller
// (sourced from the GenesisSource).
func (s *rpcTxsSource) FetchTxs(ctx context.Context, chainID string, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error) {
	stream := newTxStream(chainID, func(addr crypto.Address) std.Account {
		return s.queryAccountAtHeight(ctx, addr, toHeight, io)
	}, io)

	total := toHeight - fromHeight + 1
	var processed, txCount int64

	blocks := s.newPooledFetcher().FetchRange(ctx, fromHeight, toHeight)
	for r := range blocks {
		if r.err != nil {
			return nil, r.err
		}
		processed++
		if processed%1000 == 0 || processed == total {
			io.Printf("\r  Blocks: %d/%d  Txs: %d", processed, total, txCount)
		}
		block := r.data.block
		if len(block.Block.Data.Txs) == 0 {
			continue
		}
		txCount += int64(stream.processBlock(
			r.height,
			block.Block.Header.Time.Unix(),
			block.Block.Data.Txs,
			r.data.results.Results.DeliverTxs,
		))
	}

	stream.resolveTrailingFailures()
	io.Printf("\r  Blocks: %d/%d  Txs: %d\n", processed, total, txCount)
	return stream.txs, nil
}

// queryAccountAtHeight queries an account's state at a specific block height
// via the auth ABCI module, trying endpoints in order until one succeeds.
func (s *rpcTxsSource) queryAccountAtHeight(
	ctx context.Context, addr crypto.Address, height int64, io commands.IO,
) std.Account {
	path := fmt.Sprintf("auth/accounts/%s", addr)
	res, err := tryEndpoints(s.clients, func(c *rpcclient.RPCClient) (*coretypes.ResultABCIQuery, error) {
		return c.ABCIQueryWithOptions(ctx, path, nil, rpcclient.ABCIQueryOptions{Height: height})
	})
	if err != nil {
		io.Printf("\n  WARNING: account query failed for %s at height %d: %v\n", addr, height, err)
		return nil
	}
	if res.Response.Error != nil {
		io.Printf("\n  WARNING: account query returned error for %s at height %d: %v\n",
			addr, height, res.Response.Error)
		return nil
	}
	if len(res.Response.Data) == 0 {
		io.Printf("\n  WARNING: empty account response for %s at height %d\n", addr, height)
		return nil
	}

	// Decode as gnoland.GnoAccount; the auth handler returns the concrete
	// account type the gno.land app installs.
	var acc gnoland.GnoAccount
	if err := amino.UnmarshalJSON(res.Response.Data, &acc); err != nil {
		io.Printf("\n  WARNING: could not decode account %s at height %d: %v\n",
			addr, height, err)
		return nil
	}
	if acc.Address.IsZero() {
		return nil
	}
	return &acc.BaseAccount
}

// newPooledFetcher wires this source's client pool into a pooledFetcher
// instance configured for block range streaming.
func (s *rpcTxsSource) newPooledFetcher() *pooledFetcher {
	return &pooledFetcher{
		numEndpoints:       len(s.clients),
		workersPerEndpoint: s.workersPerEndpoint,
		maxCycles:          rpcFetcherMaxCycles,
		backoff:            rpcFetcherBackoff,
		fetch: func(ctx context.Context, endpoint int, h int64) (*blockData, error) {
			c := s.clients[endpoint]
			block, err := c.Block(ctx, &h)
			if err != nil {
				return nil, fmt.Errorf("block %d at %s: %w", h, s.rpcURLs[endpoint], err)
			}
			results, err := c.BlockResults(ctx, &h)
			if err != nil {
				return nil, fmt.Errorf("block results %d at %s: %w", h, s.rpcURLs[endpoint], err)
			}
			return &blockData{block: block, results: results}, nil
		},
	}
}

// ---- shared RPC helpers

// parseRPCURLs splits s on commas, trims whitespace, skips empty segments,
// and verifies every remaining segment is an http(s) URL.
func parseRPCURLs(s string) ([]string, error) {
	parts := strings.Split(s, ",")
	urls := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
			return nil, fmt.Errorf("RPC URL segment %q must start with http:// or https://", p)
		}
		if _, err := url.Parse(p); err != nil {
			return nil, fmt.Errorf("invalid RPC URL %q: %w", p, err)
		}
		urls = append(urls, p)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("no usable RPC URLs in %q", s)
	}
	return urls, nil
}

// openRPCClients constructs one HTTP client per URL. On the first failure it
// closes every client it already opened and returns the error — the caller
// gets either a fully-open pool or none.
func openRPCClients(urls []string) ([]*rpcclient.RPCClient, error) {
	clients := make([]*rpcclient.RPCClient, 0, len(urls))
	for _, u := range urls {
		c, err := rpcclient.NewHTTPClient(u)
		if err != nil {
			for _, prev := range clients {
				_ = prev.Close()
			}
			return nil, fmt.Errorf("creating RPC client for %s: %w", u, err)
		}
		clients = append(clients, c)
	}
	return clients, nil
}

// tryEndpoints calls fn against each client in order, returning the first
// successful result. Returns errNoEndpoints if clients is empty, or the last
// error wrapped with endpoint count when every client fails.
func tryEndpoints[T any](
	clients []*rpcclient.RPCClient,
	fn func(*rpcclient.RPCClient) (T, error),
) (T, error) {
	var zero T
	if len(clients) == 0 {
		return zero, errNoEndpoints
	}
	var lastErr error
	for _, c := range clients {
		v, err := fn(c)
		if err == nil {
			return v, nil
		}
		lastErr = err
	}
	return zero, fmt.Errorf("all %d endpoint(s) failed: %w", len(clients), lastErr)
}
