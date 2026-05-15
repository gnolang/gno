package state

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// mockJSONClient is a configurable ClientAdapter for JSON-endpoint tests.
// Each method returns a canned (bytes, error) pair so tests can isolate the
// success / not-found / generic-failure paths.
type mockJSONClient struct {
	pkgBytes []byte
	pkgErr   error
	objBytes []byte
	objErr   error
	typBytes []byte
	typErr   error
}

func (m *mockJSONClient) Realm(context.Context, string, string) ([]byte, error) {
	return nil, nil
}

func (m *mockJSONClient) ListPaths(context.Context, string, int) ([]string, error) {
	return nil, nil
}

func (m *mockJSONClient) Doc(context.Context, string, int64) (*doc.JSONDocumentation, error) {
	return nil, nil
}

func (m *mockJSONClient) StatePkg(_ context.Context, _ string, _ int64) ([]byte, error) {
	return m.pkgBytes, m.pkgErr
}

func (m *mockJSONClient) StateObject(_ context.Context, _ string, _ int64) ([]byte, error) {
	return m.objBytes, m.objErr
}

func (m *mockJSONClient) StateType(_ context.Context, _ string, _ int64) ([]byte, error) {
	return m.typBytes, m.typErr
}

func newJSONHandler(client *mockJSONClient) *Handler {
	return New(Deps{
		Client:      client,
		Highlighter: stubHighlighter{},
	})
}

// validOID / validTID satisfy the regex in validate.go (40-hex hash, with
// `:N` suffix for OID).
const (
	validOID = "abcdef0123456789abcdef0123456789abcdef01:1"
	validTID = "abcdef0123456789abcdef0123456789abcdef01"
)

func serveJSONReq(t *testing.T, h *Handler, query url.Values) *httptest.ResponseRecorder {
	t.Helper()
	if !query.Has("state") {
		query.Set("state", "")
	}
	if !query.Has("json") {
		query.Set("json", "")
	}
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: query}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state&json", nil)
	rec := httptest.NewRecorder()
	h.Handle(context.Background(), rec, req, u)
	return rec
}

func TestServeJSONPackageHappyPath(t *testing.T) {
	want := []byte(`{"hello":"world"}`)
	h := newJSONHandler(&mockJSONClient{pkgBytes: want})
	rec := serveJSONReq(t, h, url.Values{})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json...", got)
	}
	if got := rec.Body.Bytes(); string(got) != string(want) {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestServeJSONObjectHappyPath(t *testing.T) {
	want := []byte(`{"oid":"abc","value":42}`)
	h := newJSONHandler(&mockJSONClient{objBytes: want})
	rec := serveJSONReq(t, h, url.Values{"oid": {validOID}})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.Bytes(); string(got) != string(want) {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestServeJSONTypeHappyPath(t *testing.T) {
	want := []byte(`{"tid":"abc","kind":"struct"}`)
	h := newJSONHandler(&mockJSONClient{typBytes: want})
	rec := serveJSONReq(t, h, url.Values{"tid": {validTID}})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.Bytes(); string(got) != string(want) {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestServeJSONInvalidOID400(t *testing.T) {
	h := newJSONHandler(&mockJSONClient{})
	rec := serveJSONReq(t, h, url.Values{"oid": {"garbage"}})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var env map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not JSON: %v (body=%q)", err, rec.Body.String())
	}
	if env["error"] != "invalid object id" {
		t.Fatalf("error = %q, want %q", env["error"], "invalid object id")
	}
}

func TestServeJSONInvalidTID400(t *testing.T) {
	h := newJSONHandler(&mockJSONClient{})
	// A TypeID is a human-readable string, so most inputs are accepted;
	// a control char is what ValidateTID actually rejects.
	rec := serveJSONReq(t, h, url.Values{"tid": {"bad\ttid"}})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var env map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not JSON: %v (body=%q)", err, rec.Body.String())
	}
	if env["error"] != "invalid type id" {
		t.Fatalf("error = %q, want %q", env["error"], "invalid type id")
	}
}

func TestServeJSONNotFound404(t *testing.T) {
	// Sentinel-text match: feature/state can't import gnoweb's ErrClient*
	// (cycle), so it matches on the stable error-message substring.
	h := newJSONHandler(&mockJSONClient{pkgErr: errors.New("package not found")})
	rec := serveJSONReq(t, h, url.Values{})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var env map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not JSON: %v (body=%q)", err, rec.Body.String())
	}
	if env["error"] == "" {
		t.Fatalf("error envelope is empty: %q", rec.Body.String())
	}
}

func TestServeJSONInternalError500(t *testing.T) {
	h := newJSONHandler(&mockJSONClient{pkgErr: errors.New("boom: chain blew up")})
	rec := serveJSONReq(t, h, url.Values{})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var env map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not JSON: %v (body=%q)", err, rec.Body.String())
	}
	if env["error"] != "internal error" {
		t.Fatalf("error = %q, want %q (internals must be hidden)", env["error"], "internal error")
	}
}

func TestServeJSONCacheControlLatest(t *testing.T) {
	h := newJSONHandler(&mockJSONClient{pkgBytes: []byte(`{}`)})
	rec := serveJSONReq(t, h, url.Values{}) // no height → latest

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	got := rec.Header().Get("Cache-Control")
	if !strings.Contains(got, "max-age=1") {
		t.Fatalf("Cache-Control = %q, want max-age=1 for latest height", got)
	}
}

func TestServeJSONCacheControlPinned(t *testing.T) {
	h := newJSONHandler(&mockJSONClient{pkgBytes: []byte(`{}`)})
	rec := serveJSONReq(t, h, url.Values{"height": {"12345"}})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	got := rec.Header().Get("Cache-Control")
	if !strings.Contains(got, "max-age=86400") || !strings.Contains(got, "immutable") {
		t.Fatalf("Cache-Control = %q, want max-age=86400 + immutable for pinned height", got)
	}
}

func TestServeJSONNoIndexHeader(t *testing.T) {
	// X-Robots-Tag on success keeps per-height snapshots out of search engines.
	h := newJSONHandler(&mockJSONClient{pkgBytes: []byte(`{}`)})
	rec := serveJSONReq(t, h, url.Values{})

	if got := rec.Header().Get("X-Robots-Tag"); got != "noindex, nofollow" {
		t.Fatalf("X-Robots-Tag = %q, want %q", got, "noindex, nofollow")
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want %q", got, "nosniff")
	}
}
