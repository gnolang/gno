package state

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// servePageReqB invokes servePage and discards the body; used by
// BenchmarkInitialPagePackage where we measure work, not output bytes.
func servePageReqB(b *testing.B, h *Handler, path string) *httptest.ResponseRecorder {
	b.Helper()
	u := &weburl.GnoURL{Path: path, WebQuery: url.Values{"state": {""}}}
	req := httptest.NewRequest(http.MethodGet, path+"$state", nil)
	rec := httptest.NewRecorder()
	status, view := h.servePage(context.Background(), rec, req, u)
	rec.Code = status
	if view != nil {
		_ = view.Render(rec.Body)
	}
	return rec
}

// Benchmark + amplification measurement against ADR-004 §Amplification.
//
// Goal: verify that the ResolvePreviews path keeps the amplification
// factor near 1× per HTTP request on the typical realm shape (top-level
// declarations + ≤15 inline previews). Each benchmark records:
//
//   - b.N iterations
//   - total RPCs spent (cumulative across iterations via fetcher.calls)
//   - RPCs per iteration (the actual amplification per simulated GET)
//
// All client fetches are stubbed; latency is in-memory so this measures
// resolve overhead + RPC count, not wall clock. Reported via b.ReportMetric
// so `go test -bench` surfaces the amplification number.

// countingFetcher implements ObjectFetcher and records call counts.
type countingFetcher struct {
	bodies   map[string][]byte
	objCalls int64
}

func (f *countingFetcher) FetchObject(_ context.Context, oid string) ([]byte, error) {
	atomic.AddInt64(&f.objCalls, 1)
	if b, ok := f.bodies[oid]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("not found: %s", oid)
}

// buildRefNodes returns N ref-shaped top-level nodes paired with canned
// struct bodies. Each ref points at a unique OID with a 2-field struct
// body — the shape ResolvePreviews actually decodes and inlines.
func buildRefNodes(n int) ([]StateNode, map[string][]byte) {
	bodies := make(map[string][]byte, n)
	nodes := make([]StateNode, n)
	for i := 0; i < n; i++ {
		oid := fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1)
		bodies[oid] = previewStructBody(oid, i, i+1)
		nodes[i] = StateNode{
			Name: fmt.Sprintf("R%d", i), Kind: KindRef,
			ObjectID: oid, Expandable: true,
		}
	}
	return nodes, bodies
}

// BenchmarkResolvePreviews measures the bounded resolve path on 20 top-
// level refs (typical realm shape) — the cap-15 budget should produce
// ≤15 RPCs per iteration.
func BenchmarkResolvePreviews(b *testing.B) {
	nodes0, bodies := buildRefNodes(20)
	fetcher := &countingFetcher{bodies: bodies}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fresh node copy each iter — ResolvePreviews mutates in place.
		nodes := make([]StateNode, len(nodes0))
		copy(nodes, nodes0)
		if _, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if b.N > 0 {
		b.ReportMetric(float64(atomic.LoadInt64(&fetcher.objCalls))/float64(b.N), "rpc/op")
	}
}

// BenchmarkAmplification reports the per-HTTP-request amplification
// factor: 1 simulated GET → N upstream RPCs. Reported as rpc/op so a
// regression past the ADR-004 §Amplification cap surfaces in `go test
// -bench` output.
func BenchmarkAmplification(b *testing.B) {
	for _, count := range []int{5, 20, 50} {
		nodes0, bodies := buildRefNodes(count)
		b.Run(fmt.Sprintf("n=%d", count), func(b *testing.B) {
			fetcher := &countingFetcher{bodies: bodies}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nodes := make([]StateNode, len(nodes0))
				copy(nodes, nodes0)
				_, _ = ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
			}
			b.StopTimer()
			if b.N > 0 {
				b.ReportMetric(float64(atomic.LoadInt64(&fetcher.objCalls))/float64(b.N), "rpc/op")
			}
		})
	}
}

// BenchmarkInitialPagePackage measures end-to-end servePage cost on a
// small realm fixture — drives the full DecodePackage + ResolvePreviews
// + render pipeline. RPC budget per ADR-004 §URL contract: 1 qpkg + 1
// qdoc + ≤15 previews per HTTP request. The fixture in pageFixturePkg
// has no top-level refs so previews resolve to 0 RPCs.
func BenchmarkInitialPagePackage(b *testing.B) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = servePageReqB(b, h, "/r/demo")
	}
	b.StopTimer()
	// Report cumulative RPC count / iteration. With no refs, previews
	// adds zero — the floor is 1 qpkg + 1 qdoc per page = 2 RPC/op.
	if b.N > 0 {
		b.ReportMetric(float64(atomic.LoadInt32(&client.objCalls))/float64(b.N), "rpc/op")
	}
}

// BenchmarkRateLimitAllow measures token-bucket throughput under
// contention. 1000 distinct IPs share one limiter — the cost per Allow
// call should stay sub-microsecond even with LRU eviction churn.
func BenchmarkRateLimitAllow(b *testing.B) {
	l := NewIPLimiter(RateLimitConfig{
		PerMinute: 100,
		Burst:     100,
		MaxIPs:    1000,
	})
	if l == nil {
		b.Fatal("limiter not constructed")
	}

	ips := make([]string, 1000)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.%d.%d.%d", (i>>16)&0xff, (i>>8)&0xff, i&0xff)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Allow(ips[i%len(ips)])
	}
}
