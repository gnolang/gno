package connect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	h := New(Deps{Meta: PageMeta{AssetsPath: "/public/", ChainId: "dev", Remote: "127.0.0.1:26657"}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/wallets", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")

	body := rec.Body.String()
	// Full gnoweb chrome, not a bare fragment.
	assert.Contains(t, body, "<!doctype html>")
	assert.Contains(t, body, "Gnokey")
	assert.Contains(t, body, "Adena")
}

func TestHandler_RejectsNonGET(t *testing.T) {
	t.Parallel()

	h := New(Deps{Meta: PageMeta{AssetsPath: "/public/"}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/wallets", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
