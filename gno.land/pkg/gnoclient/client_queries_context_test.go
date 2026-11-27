package gnoclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func TestQueryContextCancelsInFlightRPC(t *testing.T) {
	t.Parallel()

	const pkgPath = "gno.land/r/demo/counter"
	addr, err := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{"Query", func(ctx context.Context, c *Client) error {
			_, err := c.QueryContext(ctx, QueryCfg{Path: "vm/qfile", Data: []byte(pkgPath)})
			return err
		}},
		{"Render", func(ctx context.Context, c *Client) error {
			_, _, err := c.RenderContext(ctx, pkgPath, "")
			return err
		}},
		{"QEval", func(ctx context.Context, c *Client) error {
			_, _, err := c.QEvalContext(ctx, pkgPath, "Counter()")
			return err
		}},
		{"QueryAccount", func(ctx context.Context, c *Client) error {
			_, _, err := c.QueryAccountContext(ctx, addr)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guard, stop := context.WithTimeout(t.Context(), 5*time.Second)
			defer stop()
			started := make(chan struct{})
			rpcCanceled := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				close(started)
				select {
				case <-r.Context().Done():
					close(rpcCanceled)
				case <-guard.Done():
				}
			}))
			defer server.Close()

			rpc, err := rpcclient.NewHTTPClient(server.URL)
			require.NoError(t, err)
			client := &Client{RPCClient: rpc}
			ctx, cancel := context.WithCancel(guard)
			defer cancel()
			result := make(chan error, 1)
			go func() { result <- tc.call(ctx, client) }()

			select {
			case <-started:
				cancel()
			case <-guard.Done():
				t.Fatal("query did not reach the RPC server")
			}
			select {
			case err := <-result:
				require.ErrorIs(t, err, context.Canceled)
			case <-guard.Done():
				t.Fatal("query did not return after cancellation")
			}
			select {
			case <-rpcCanceled:
			case <-guard.Done():
				t.Fatal("RPC server did not observe cancellation")
			}
		})
	}
}
