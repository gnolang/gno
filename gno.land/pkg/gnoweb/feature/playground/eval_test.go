package playground

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// funcClient answers Eval with fn and counts the calls.
type funcClient struct {
	stubClient
	calls atomic.Int32
	fn    func(ctx context.Context, pkgPath, expr string) ([]byte, error)
}

func (c *funcClient) Eval(ctx context.Context, pkgPath, expr string) ([]byte, error) {
	c.calls.Add(1)
	return c.fn(ctx, pkgPath, expr)
}

func serveEvalJSON(t *testing.T, h *Handler, pkgPath, expr string) (int, evalResponse) {
	t.Helper()

	rr := httptest.NewRecorder()
	h.EvalHandler().ServeHTTP(rr, newEvalRequest(t, pkgPath, expr))
	return decodeEval(t, rr)
}

func newEvalRequest(t *testing.T, pkgPath, expr string) *http.Request {
	t.Helper()

	body, err := json.Marshal(evalRequest{PkgPath: pkgPath, Expression: expr})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/_/api/eval", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeEval(t *testing.T, rr *httptest.ResponseRecorder) (int, evalResponse) {
	t.Helper()

	var resp evalResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	return rr.Code, resp
}

func TestEvalCache(t *testing.T) {
	t.Parallel()

	t.Run("answers are reused, per package and expression", func(t *testing.T) {
		t.Parallel()
		c := &funcClient{fn: func(_ context.Context, p, e string) ([]byte, error) {
			return []byte(p + "." + e), nil
		}}
		h := New(Deps{Packages: &stubPackages{}, Client: c, Logger: discardLogger()})

		for range 3 {
			_, resp := serveEvalJSON(t, h, "r/a", "X()")
			assert.Equal(t, "r/a.X()", resp.Result)
		}
		_, resp := serveEvalJSON(t, h, "r/a", "Y()")
		assert.Equal(t, "r/a.Y()", resp.Result)
		_, resp = serveEvalJSON(t, h, "r/b", "X()")
		assert.Equal(t, "r/b.X()", resp.Result)
		assert.EqualValues(t, 3, c.calls.Load())
	})

	t.Run("chain errors are reused, node failures are not", func(t *testing.T) {
		t.Parallel()
		c := &funcClient{fn: func(_ context.Context, _, e string) ([]byte, error) {
			if e == "Bad()" {
				return nil, fmt.Errorf("name Bad not declared")
			}
			return nil, fmt.Errorf("%w: refused", ErrEvalUnavailable)
		}}
		h := New(Deps{Packages: &stubPackages{}, Client: c, Logger: discardLogger()})

		for range 2 {
			code, _ := serveEvalJSON(t, h, "r/a", "Bad()")
			assert.Equal(t, http.StatusOK, code)
		}
		assert.EqualValues(t, 1, c.calls.Load())
		for range 2 {
			code, _ := serveEvalJSON(t, h, "r/a", "Down()")
			assert.Equal(t, http.StatusBadGateway, code)
		}
		assert.EqualValues(t, 3, c.calls.Load())
	})

	t.Run("concurrent readers share one query", func(t *testing.T) {
		t.Parallel()
		release := make(chan struct{})
		c := &funcClient{fn: func(context.Context, string, string) ([]byte, error) {
			<-release
			return []byte("ok"), nil
		}}
		h := New(Deps{Packages: &stubPackages{}, Client: c, Logger: discardLogger()})

		// Only ServeHTTP runs off the test goroutine: requests are built and
		// answers checked here, as require may only stop this goroutine.
		const readers = 8
		var wg sync.WaitGroup
		recs := make([]*httptest.ResponseRecorder, readers)
		for i := range readers {
			req := newEvalRequest(t, "r/a", "X()")
			recs[i] = httptest.NewRecorder()
			wg.Go(func() { h.EvalHandler().ServeHTTP(recs[i], req) })
		}
		// Let every reader join the flight before it lands.
		require.Eventually(t, func() bool { return c.calls.Load() == 1 }, time.Second, time.Millisecond)
		time.Sleep(20 * time.Millisecond)
		close(release)
		wg.Wait()

		assert.EqualValues(t, 1, c.calls.Load())
		for _, rr := range recs {
			code, resp := decodeEval(t, rr)
			assert.Equal(t, http.StatusOK, code)
			assert.Equal(t, "ok", resp.Result)
		}
	})
}

func TestEvalResultCap(t *testing.T) {
	t.Parallel()

	// A multi-byte rune straddles the cap, so the cut must back off to its start.
	big := strings.Repeat("a", maxEvalResultBytes-1) + "é" + strings.Repeat("b", 1000)
	c := &funcClient{fn: func(context.Context, string, string) ([]byte, error) {
		return []byte(big), nil
	}}
	h := New(Deps{Packages: &stubPackages{}, Client: c, Logger: discardLogger()})

	code, resp := serveEvalJSON(t, h, "r/a", "Big()")
	require.Equal(t, http.StatusOK, code)
	head, marker, ok := strings.Cut(resp.Result, "\n… (truncated")
	require.True(t, ok, "missing truncation marker")
	assert.Equal(t, strings.Repeat("a", maxEvalResultBytes-1), head)
	assert.Contains(t, marker, fmt.Sprintf("of %d bytes shown", len(big)))

	small := &funcClient{fn: func(context.Context, string, string) ([]byte, error) {
		return []byte("(1 int)"), nil
	}}
	_, resp = serveEvalJSON(t, New(Deps{Packages: &stubPackages{}, Client: small, Logger: discardLogger()}), "r/a", "One()")
	assert.Equal(t, "(1 int)", resp.Result)
}

func TestEvalDeadline(t *testing.T) {
	t.Parallel()

	// serve answers one eval, failing fast instead of hanging when no
	// deadline ends it. Only the handler runs off the test goroutine: t
	// must not be used there, and a goroutine left behind by a failure
	// writes to its own recorder only.
	serve := func(t *testing.T, h *Handler) (int, evalResponse) {
		t.Helper()
		req := newEvalRequest(t, "r/a", "Slow()")
		rr := httptest.NewRecorder()
		done := make(chan struct{})
		go func() {
			defer close(done)
			h.EvalHandler().ServeHTTP(rr, req)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("the eval did not end at its deadline")
		}
		return decodeEval(t, rr)
	}

	t.Run("the reader is answered when the node ignores the deadline", func(t *testing.T) {
		t.Parallel()
		release := make(chan struct{})
		t.Cleanup(func() { close(release) })
		c := &funcClient{fn: func(context.Context, string, string) ([]byte, error) {
			<-release
			return nil, nil
		}}
		h := New(Deps{Packages: &stubPackages{}, Client: c, Logger: discardLogger(), Timeout: 20 * time.Millisecond})

		code, resp := serve(t, h)
		assert.Equal(t, http.StatusGatewayTimeout, code)
		assert.Equal(t, ErrEvalTimeout.Error(), resp.Error)
	})

	t.Run("the node query has its own deadline", func(t *testing.T) {
		t.Parallel()
		returned := make(chan struct{})
		c := &funcClient{fn: func(ctx context.Context, _, _ string) ([]byte, error) {
			defer close(returned)
			<-ctx.Done()
			return nil, fmt.Errorf("%w: %w", ErrEvalTimeout, ctx.Err())
		}}
		h := New(Deps{Packages: &stubPackages{}, Client: c, Logger: discardLogger(), Timeout: 20 * time.Millisecond})

		code, resp := serve(t, h)
		assert.Equal(t, http.StatusGatewayTimeout, code)
		assert.Equal(t, ErrEvalTimeout.Error(), resp.Error)
		select {
		case <-returned:
		case <-time.After(5 * time.Second):
			t.Fatal("the detached node query never ended")
		}
	})
}
