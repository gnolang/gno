package client

import (
	"context"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubABCIClient implements client.ABCIClient; only ABCIQuery is exercised, so
// the embedded interface (nil) covers the rest of the method set.
type stubABCIClient struct {
	client.ABCIClient

	resp    *ctypes.ResultABCIQuery
	err     error
	gotPath string
	gotData []byte
}

func (s *stubABCIClient) ABCIQuery(_ context.Context, path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	s.gotPath = path
	s.gotData = data
	return s.resp, s.err
}

func TestProfileTx(t *testing.T) {
	t.Parallel()

	pprof := []byte("\x1f\x8bfake-gzipped-pprof")
	stub := &stubABCIClient{
		resp: &ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				ResponseBase: abci.ResponseBase{Log: "ok"},
				Value:        pprof,
			},
		},
	}

	profile, log, err := ProfileTx(stub, []byte("txbytes"))
	require.NoError(t, err)
	assert.Equal(t, ".app/profiletx", stub.gotPath, "queries the profiletx endpoint")
	assert.Equal(t, []byte("txbytes"), stub.gotData, "forwards the tx bytes")
	assert.Equal(t, pprof, profile, "returns the profile from Response.Value")
	assert.Equal(t, "ok", log, "returns the status log")
}

// A node without the profiler enabled answers with Response.Error; ProfileTx
// must surface it (with the node's log) rather than returning an empty profile.
func TestProfileTx_disabledNode(t *testing.T) {
	t.Parallel()

	const msg = "tx gas profiling is not enabled on this node"
	stub := &stubABCIClient{
		resp: &ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				ResponseBase: abci.ResponseBase{
					Error: abci.ABCIErrorOrStringError(std.ErrUnknownRequest(msg)),
					Log:   msg,
				},
			},
		},
	}

	profile, _, err := ProfileTx(stub, []byte("txbytes"))
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "not enabled")
}
