package gnoweb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/networks"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerNetworksJSON(t *testing.T) {
	h := handlerNetworksJSON(log.NewTestingLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/api/networks", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.Equal(t, "public, max-age=3600", res.Header.Get("Cache-Control"))
	assert.Equal(t, "*", res.Header.Get("Access-Control-Allow-Origin"))

	etag := res.Header.Get("ETag")
	require.NotEmpty(t, etag, "expected ETag header")

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var reg networks.Registry
	require.NoError(t, json.Unmarshal(body, &reg))
	require.NotEmpty(t, reg.Networks, "expected at least one network")

	// Body should be byte-identical to the embedded registry payload.
	assert.Equal(t, networks.Raw(), body)

	// Conditional GET with the same ETag should yield 304 with empty body.
	req2 := httptest.NewRequest(http.MethodGet, "/api/networks", nil)
	req2.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	res2 := rec2.Result()
	defer res2.Body.Close()
	assert.Equal(t, http.StatusNotModified, res2.StatusCode)
	body2, err := io.ReadAll(res2.Body)
	require.NoError(t, err)
	assert.Empty(t, body2)
}
