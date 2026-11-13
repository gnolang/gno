package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestBroadcastHandlerSuggestedGasWanted checks that a failed simulation only
// suggests a gas-wanted when its gas used is a real figure. A simulation run at
// the tx's own GasWanted (a Ledger with -simulate test) stops where the tx runs
// out, so a suggestion built on it would be too low.
func TestBroadcastHandlerSuggestedGasWanted(t *testing.T) {
	t.Parallel()

	const gasWanted = int64(100)

	tests := []struct {
		name           string
		simulateMaxGas int64
		gasUsed        int64
		wantSuggestion bool
	}{
		{"simulated as signed and cut short: no suggestion", 0, gasWanted, false},
		{"simulated as signed, failed for another reason: suggestion", 0, 40, true},
		{"simulated at max gas: suggestion", 1_000, 150, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value, err := amino.Marshal(abci.ResponseDeliverTx{
				ResponseBase: abci.ResponseBase{Error: abci.StringError("fail")},
				GasUsed:      tt.gasUsed,
			})
			require.NoError(t, err)
			server := createTestServer(t, defaultHTTPHandler(t, "abci_query", &ctypes.ResultABCIQuery{
				Response: abci.ResponseQuery{Value: value},
			}))

			tx := std.Tx{Fee: std.NewFee(gasWanted, std.MustParseCoin("1ugnot"))}
			res, err := BroadcastHandler(&BroadcastCfg{
				RootCfg:        &BaseCfg{BaseOptions: BaseOptions{Remote: server.URL}},
				tx:             &tx,
				testSimulate:   true,
				simulateMaxGas: tt.simulateMaxGas,
			})
			require.NoError(t, err)
			require.True(t, res.DeliverTx.IsErr())

			if tt.wantSuggestion {
				assert.Contains(t, res.DeliverTx.Info, "suggested gas-wanted")
			} else {
				assert.NotContains(t, res.DeliverTx.Info, "suggested gas-wanted")
			}
		})
	}
}
