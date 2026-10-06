package state

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// servePageReqB invokes servePage and discards the body; used to bench
// servePage cost without measuring template output bytes.
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

// refsPkgJSON builds a qpkg_json with n top-level ref slots. Drives
// the amplification benchmark: previews hydrate lazily via hx-trigger=
// revealed, never bursted server-side, so rpc/op must stay flat.
func refsPkgJSON(n int) []byte {
	var refs []string
	for i := range n {
		oid := fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1)
		refs = append(refs,
			fmt.Sprintf(`{"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo.T"}, "V": {"@type": "/gno.RefValue", "ObjectID": %q}}`, oid))
	}
	names := make([]string, n)
	for i := range names {
		names[i] = fmt.Sprintf("%q", fmt.Sprintf("R%d", i))
	}
	return fmt.Appendf(nil, `{"names":[%s],"values":[%s]}`,
		strings.Join(names, ","), strings.Join(refs, ","))
}

// BenchmarkAmplification reports rpc/op (per simulated GET). Lazy
// hydration via hx-trigger="revealed" keeps rpc/op flat regardless of
// ref count; a regression past 0 here means the burst is back.
func BenchmarkAmplification(b *testing.B) {
	for _, count := range []int{5, 20, 50} {
		b.Run(fmt.Sprintf("n=%d", count), func(b *testing.B) {
			client := &pageMockClient{pkgBytes: refsPkgJSON(count)}
			h := newPageHandler(client)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = servePageReqB(b, h, "/r/demo")
			}
			b.StopTimer()
			if b.N > 0 {
				b.ReportMetric(float64(atomic.LoadInt32(&client.objCalls))/float64(b.N), "rpc/op")
			}
		})
	}
}

// BenchmarkInitialPagePackage measures end-to-end servePage cost.
// Floor is 2 RPCs (qpkg + qdoc); previews resolve lazily client-side.
func BenchmarkInitialPagePackage(b *testing.B) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = servePageReqB(b, h, "/r/demo")
	}
	b.StopTimer()
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
