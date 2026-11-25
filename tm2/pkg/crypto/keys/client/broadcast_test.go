package client

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestSimulateTxRefused checks that a simulate the node refused, with an error
// and no value, is an error and not an empty result that reads as a pass, so
// gnokey does not go on to broadcast it.
func TestSimulateTxRefused(t *testing.T) {
	t.Parallel()

	server := createTestServer(t, defaultHTTPHandler(t, "abci_query", &ctypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			ResponseBase: abci.ResponseBase{
				Error: std.InternalError{},
				Log:   "node refused the query",
			},
		},
	}))
	cli, err := client.NewHTTPClient(server.URL)
	require.NoError(t, err)

	res, err := SimulateTx(cli, []byte("tx"))
	require.Nil(t, res)
	require.ErrorContains(t, err, "node refused the query")
}
