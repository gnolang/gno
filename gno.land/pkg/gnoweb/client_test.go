package gnoweb

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/feature/state"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestCheckResponseSizeRejectsOversized — defense in depth against a
// misbehaving or compromised RPC node returning a multi-MB amino blob:
// gnoweb caps every per-query response so the decode pipeline cannot be
// pressured into a memory amplification attack.
func TestCheckResponseSizeRejectsOversized(t *testing.T) {
	for _, tc := range []struct {
		name    string
		size    int
		wantErr error
	}{
		{"empty", 0, nil},
		{"under cap", 1024, nil},
		{"at cap", maxRPCResponseSize, nil},
		{"one byte over cap", maxRPCResponseSize + 1, ErrClientResponseTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkResponseSize(make([]byte, tc.size))
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want errors.Is(%v)", err, tc.wantErr)
			}
		})
	}
}

// TestAcquireRPCSlotBoundsConcurrency pins the semaphore contract:
// (a) cap parallelism, (b) honour ctx cancellation while waiting,
// (c) release frees exactly one slot.
func TestAcquireRPCSlotBoundsConcurrency(t *testing.T) {
	slots := make(chan struct{}, 2)

	rel1, err := acquireRPCSlot(context.Background(), slots)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	rel2, err := acquireRPCSlot(context.Background(), slots)
	if err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}

	// Bucket full — third acquire must block; ctx deadline triggers an
	// orderly cancellation rather than a stuck goroutine.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := acquireRPCSlot(ctx, slots); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded waiting for slot, got %v", err)
	}

	// Release frees a slot — the next acquire succeeds immediately.
	rel1()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	rel3, err := acquireRPCSlot(ctx2, slots)
	if err != nil {
		t.Fatalf("expected acquire after release, got %v", err)
	}
	rel3()
	rel2()

	// Bucket fully drained — len must be 0 (no slot leak from release fn).
	if len(slots) != 0 {
		t.Fatalf("slot leak: len(slots)=%d, want 0", len(slots))
	}
}

// TestStateErrorSentinelPact pins the message contract between gnoweb's
// ErrClient* sentinels and the substrings feature/state matches on in
// mapClientError. The state package cannot import gnoweb (cycle), so a
// silent rename of either side would route real 404s as generic 500s
// without anyone noticing — this test catches that at build time.
func TestStateErrorSentinelPact(t *testing.T) {
	for _, tc := range []struct {
		name     string
		sentinel error
		substr   string
	}{
		{"package not found", ErrClientPackageNotFound, state.ClientErrPackageNotFound},
		{"file not found", ErrClientFileNotFound, state.ClientErrFileNotFound},
		{"object not found", ErrClientObjectNotFound, state.ClientErrObjectNotFound},
		{"timeout", ErrClientTimeout, state.ClientErrTimeout},
		{"bad request", ErrClientBadRequest, state.ClientErrBadRequest},
		{"response too large", ErrClientResponseTooLarge, state.ClientErrResponseTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.sentinel.Error(), tc.substr) {
				t.Fatalf("sentinel %q does not contain feature/state substring %q — mapClientError will not classify this error correctly",
					tc.sentinel.Error(), tc.substr)
			}
		})
	}
}

// pathsCaller is a JSON-RPC caller that records the abci_query path it is
// sent and answers with a fixed listing.
type pathsCaller struct {
	gotPath string
	listing string
	calls   int
}

func (p *pathsCaller) SendRequest(_ context.Context, req rpctypes.RPCRequest) (*rpctypes.RPCResponse, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, err
	}
	p.gotPath = params.Path
	p.calls++

	result, err := amino.MarshalJSON(&ctypes.ResultABCIQuery{
		Response: abci.ResponseQuery{ResponseBase: abci.ResponseBase{Data: []byte(p.listing)}},
	})
	if err != nil {
		return nil, err
	}
	return &rpctypes.RPCResponse{JSONRPC: "2.0", ID: req.ID, Result: result}, nil
}

func (p *pathsCaller) SendBatch(context.Context, rpctypes.RPCRequests) (rpctypes.RPCResponses, error) {
	return nil, errors.New("not implemented")
}

func (p *pathsCaller) Close() error { return nil }

// The node reads qpaths' cap off the query string and silently applies its
// own default (1000) without one, so `limit` must reach it there.
func TestListPathsForwardsLimit(t *testing.T) {
	t.Parallel()

	caller := &pathsCaller{listing: "gno.land/r/demo/a\ngno.land/r/demo/b"}
	c := NewRPCClientAdapter(newDiscardLogger(), client.NewRPCClient(caller), "gno.land", 0)

	paths, err := c.ListPaths(context.Background(), "gno.land/r", 10_000)
	if err != nil {
		t.Fatalf("ListPaths: %v", err)
	}
	if caller.gotPath != "vm/qpaths?limit=10000" {
		t.Fatalf("query path = %q, want %q", caller.gotPath, "vm/qpaths?limit=10000")
	}
	if want := []string{"/r/demo/a", "/r/demo/b"}; strings.Join(paths, ",") != strings.Join(want, ",") {
		t.Fatalf("paths = %v, want %v (domain-relative)", paths, want)
	}

	// No limit leaves the node's default alone rather than sending `limit=0`.
	if _, err := c.ListPaths(context.Background(), "gno.land/r", 0); err != nil {
		t.Fatalf("ListPaths: %v", err)
	}
	if caller.gotPath != "vm/qpaths" {
		t.Fatalf("query path = %q, want %q", caller.gotPath, "vm/qpaths")
	}
}

// Before ListPaths forwarded its limit, /u/<user> got the node's default of
// 1000 paths per namespace. Forwarding MaxUserContributions must not lower
// that: the page counts what it receives as the total.
func TestUserContributionsKeepNodeDefault(t *testing.T) {
	t.Parallel()

	caller := &pathsCaller{listing: "gno.land/r/alice/a"}
	c := NewRPCClientAdapter(newDiscardLogger(), client.NewRPCClient(caller), "gno.land", 0)
	if _, err := c.ListPaths(context.Background(), "@alice", MaxUserContributions); err != nil {
		t.Fatal(err)
	}
	if caller.gotPath != "vm/qpaths?limit=1000" {
		t.Fatalf("query path = %q, want the node default of 1000", caller.gotPath)
	}
}

func TestValidFileName(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"render.gno":          true,
		"gnomod.toml":         true,
		"README.md":           true,
		"README":              true,
		"LICENSE":             true,
		".gitignore":          true,
		"init":                false, // no dot: vm/qfile reads it as a child package
		".":                   false,
		"..":                  false,
		"../other/render.gno": false,
		"sub/render.gno":      false,
		"/render.gno":         false,
		"render.gno/":         false,
	}
	for name, want := range cases {
		if got := validFileName(name); got != want {
			t.Errorf("validFileName(%q) = %v, want %v", name, got, want)
		}
	}
}

// GasPrice reads auth/gasprice as the node writes it, reuses the answer for
// gasPriceTTL, a failure included, and asks again once it expires.
func TestGasPriceCached(t *testing.T) {
	t.Parallel()

	caller := &pathsCaller{listing: "{\n  \"gas\": \"1000\",\n  \"price\": \"1ugnot\"\n}"}
	c := NewRPCClientAdapter(newDiscardLogger(), client.NewRPCClient(caller), "gno.land", 0).(*rpcClient)
	ctx := context.Background()

	gp, err := c.GasPrice(ctx)
	if err != nil {
		t.Fatalf("GasPrice: %v", err)
	}
	if caller.gotPath != "auth/gasprice" {
		t.Fatalf("query path = %q, want auth/gasprice", caller.gotPath)
	}
	if want := std.NewCoin("ugnot", 1); gp.Gas != 1000 || gp.Price != want {
		t.Fatalf("gas price = %v, want 1ugnot/1000gas", gp)
	}

	caller.listing = "{\"gas\":\"1000\",\"price\":\"2ugnot\"}"
	if gp, _ = c.GasPrice(ctx); caller.calls != 1 || gp.Price.Amount != 1 {
		t.Fatalf("within the TTL: %d queries, price %v; want 1 query, the cached price", caller.calls, gp)
	}

	c.gasAt = c.gasAt.Add(-gasPriceTTL)
	if gp, _ = c.GasPrice(ctx); caller.calls != 2 || gp.Price.Amount != 2 {
		t.Fatalf("after the TTL: %d queries, price %v; want 2 queries, the new price", caller.calls, gp)
	}

	// A failure is cached like an answer, so a down node is asked once per TTL.
	caller.listing = "not json"
	c.gasAt = c.gasAt.Add(-gasPriceTTL)
	for range 2 {
		if gp, err = c.GasPrice(ctx); err == nil || gp != (std.GasPrice{}) {
			t.Fatalf("GasPrice = %v, %v; want the zero price and an error", gp, err)
		}
	}
	if caller.calls != 3 {
		t.Fatalf("%d queries, want 3", caller.calls)
	}
}
