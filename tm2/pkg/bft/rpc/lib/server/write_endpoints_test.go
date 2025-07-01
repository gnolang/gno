package rpcserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

func TestWriteListOfEndpoints(t *testing.T) {
	funcMap := map[string]*RPCFunc{
		"c": NewWSRPCFunc(func(ctx *types.Context, s string, i int) (string, error) { return "foo", nil }, "s,i"),
		"d": {},
	}

	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	rec := httptest.NewRecorder()
	writeListOfEndpoints(rec, req, funcMap)
	res := rec.Result()
	assert.Equal(t, res.StatusCode, 200, "Should always return 200")
	blob, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	gotResp := string(blob)
	wantResp := `<html><body><br>Available endpoints:<br><a href="//localhost/d">//localhost/d</a></br><br>Endpoints that require arguments:<br><a href="//localhost/c?s=_&i=_">//localhost/c?s=_&i=_</a></br></body></html>`
	assert.Equal(t, wantResp, gotResp)
}
