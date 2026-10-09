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
}

func (p *pathsCaller) SendRequest(_ context.Context, req rpctypes.RPCRequest) (*rpctypes.RPCResponse, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, err
	}
	p.gotPath = params.Path

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

// vm/qstorage answers "storage: <bytes>, deposit: <ugnot>"; anything else is
// an error rather than a zero that would read as an empty realm.
func TestParseStorage(t *testing.T) {
	t.Parallel()

	got, err := parseStorage([]byte("storage: 1292654, deposit: 129265400"))
	if err != nil || got.Bytes != 1292654 || got.Deposit != 129265400 {
		t.Fatalf("parseStorage = %+v, %v; want 1292654 bytes, 129265400 ugnot", got, err)
	}
	for _, bad := range []string{"", "storage: x, deposit: 1", "storage: 1", "deposit: 1, storage: 2", "storage: -1, deposit: 0"} {
		if _, err := parseStorage([]byte(bad)); err == nil {
			t.Errorf("parseStorage(%q) accepted a malformed answer", bad)
		}
	}
}
