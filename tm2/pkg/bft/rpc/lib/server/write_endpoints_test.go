package rpcserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

func TestWriteListOfEndpoints(t *testing.T) {
	funcMap := map[string]*RPCFunc{
		"c": NewWSRPCFunc(func(ctx *types.Context, s string, i int) (string, error) { return "foo", nil }, "s,i"),
		"d": {},
	}

	// The client picks the Host header, up to the server's MaxHeaderBytes,
	// so the page must not echo it.
	req := httptest.NewRequest(http.MethodGet, "http://other.example/", nil)
	rec := httptest.NewRecorder()
	makeJSONRPCHandler(funcMap, log.NewNoopLogger(), 0)(rec, req)
	res := rec.Result()
	assert.Equal(t, res.StatusCode, 200, "Should always return 200")
	blob, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	gotResp := string(blob)
	wantResp := `<html><body><br>Available endpoints:<br><a href="/d">/d</a></br><br>Endpoints that require arguments:<br><a href="/c?s=_&i=_">/c?s=_&i=_</a></br></body></html>`
	assert.Equal(t, wantResp, gotResp)
}
