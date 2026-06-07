package gnoweb

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type stubDirectory struct {
	realms, packages []string
	err              error
}

func (s stubDirectory) Paths(context.Context) ([]string, []string, error) {
	return s.realms, s.packages, s.err
}

type stubPathLister struct {
	fn func(prefix string) []string
}

func (s stubPathLister) ListPaths(_ context.Context, prefix string, _ int) ([]string, error) {
	return s.fn(prefix), nil
}

func TestRPCRealmDirectory_Paths(t *testing.T) {
	t.Parallel()
	lister := stubPathLister{fn: func(prefix string) []string {
		if strings.HasSuffix(prefix, "/r") {
			return []string{"/r/demo/boards", "", "/r/demo/users"}
		}
		return []string{"/p/demo/avl", ""}
	}}
	dir := newRPCRealmDirectory(lister, "gno.land", 4)

	realms, packages, err := dir.Paths(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"/r/demo/boards", "/r/demo/users"}, realms) // empty entry filtered
	require.Equal(t, []string{"/p/demo/avl"}, packages)
}

func TestHandlerSearchJSON_OK(t *testing.T) {
	t.Parallel()
	dir := stubDirectory{realms: []string{"/r/demo/boards"}, packages: []string{"/p/demo/avl"}}
	h := handlerSearchJSON(newDiscardLogger(), dir)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/search.json", nil))

	require.Equal(t, http.StatusOK, rr.Code)
	var got struct {
		Realms   []string `json:"realms"`
		Packages []string `json:"packages"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	assert.Equal(t, []string{"/r/demo/boards"}, got.Realms)
	assert.Equal(t, []string{"/p/demo/avl"}, got.Packages)
}

func TestHandlerSearchJSON_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	h := handlerSearchJSON(newDiscardLogger(), stubDirectory{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/search.json", nil))
	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandlerSearchJSON_UpstreamError(t *testing.T) {
	t.Parallel()
	h := handlerSearchJSON(newDiscardLogger(), stubDirectory{err: errors.New("boom")})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/search.json", nil))
	require.Equal(t, http.StatusBadGateway, rr.Code)
}
