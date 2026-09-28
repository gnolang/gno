package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tm2Client "github.com/gnolang/faucet/client/http"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeFaucet_CleanupShorterThanRateLimit(t *testing.T) {
	t.Parallel()

	cfg := &serveCfg{
		rateLimitInterval:     24 * time.Hour,
		rateLimitCleanTimeout: time.Hour,
	}

	err := serveFaucet(context.Background(), cfg, nil)

	assert.ErrorContains(t, err, "ratelimit-cleanup-timeout must be >= ratelimit-interval")
}

func TestServeFaucet_InvalidGasValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		gasFee      string
		gasWanted   int64
		expectedErr string
	}{
		{
			name:        "malformed gas fee",
			gasFee:      "invalid",
			gasWanted:   100000,
			expectedErr: "invalid gas fee",
		},
		{
			name:        "zero gas wanted",
			gasFee:      "1000000ugnot",
			gasWanted:   0,
			expectedErr: "gas wanted must be greater than zero",
		},
		{
			name:        "negative gas wanted",
			gasFee:      "1000000ugnot",
			gasWanted:   -100,
			expectedErr: "gas wanted must be greater than zero",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := &serveCfg{
				rateLimitInterval:     time.Hour,
				rateLimitCleanTimeout: 24 * time.Hour,
				gasFee:                testCase.gasFee,
				gasWanted:             testCase.gasWanted,
			}

			err := serveFaucet(context.Background(), cfg, nil)

			assert.ErrorContains(t, err, testCase.expectedErr)
		})
	}
}

func TestServeCfg_GasFlags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		args              []string
		expectedGasFee    string
		expectedGasWanted int64
	}{
		{
			name:              "defaults",
			args:              nil,
			expectedGasFee:    "1000000ugnot",
			expectedGasWanted: 100000,
		},
		{
			name: "explicit values",
			args: []string{
				"-gas-fee", "2000000ugnot",
				"-gas-wanted", "2000000",
			},
			expectedGasFee:    "2000000ugnot",
			expectedGasWanted: 2000000,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := &serveCfg{}
			fs := flag.NewFlagSet("serve", flag.ContinueOnError)
			cfg.RegisterFlags(fs)

			require.NoError(t, fs.Parse(testCase.args))

			assert.Equal(t, testCase.expectedGasFee, cfg.gasFee)
			assert.Equal(t, testCase.expectedGasWanted, cfg.gasWanted)
		})
	}
}

// transferCommitResult is the broadcast_tx_commit result of a 100 GNOT faucet
// drip on onyx-1 (height 22949). deliver_tx is the result the node recorded for
// that transaction. check_tx is omitted: CheckTx runs only the ante handler, so
// it carries no events.
const transferCommitResult = `{
	"deliver_tx": {
		"ResponseBase": {
			"Error": null,
			"Data": null,
			"Events": [
				{
					"@type": "/bank.TransferEvent",
					"from": "g1jf3xq3yur9tp9ts0h9qkl9t96kptvzq8cwshyw",
					"to": "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m",
					"coins": "100000000ugnot"
				}
			],
			"Log": "msg:0,success:true,log:,events:[]",
			"Info": ""
		},
		"GasWanted": "5000000",
		"GasUsed": "933276"
	},
	"hash": "vRDuCfcBACb7Zjk/SBQ8BTv2fTqxa9yBBawGdrdG1Wo=",
	"height": "22949"
}`

// TestServeFaucet_NodeClientDecodesBankTransferEvent checks that the node
// client serveFaucet builds decodes a broadcast_tx_commit result carrying a
// bank.TransferEvent. The client decodes with the amino types of the tm2 that
// gnofaucet links, so a drip whose result holds an event type missing from that
// tm2 is reported as failed although the transfer is committed.
func TestServeFaucet_NodeClientDecodesBankTransferEvent(t *testing.T) {
	t.Parallel()

	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rpcRequest struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&rpcRequest); err != nil {
			t.Errorf("decoding JSON-RPC request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if rpcRequest.Method != "broadcast_tx_commit" {
			t.Errorf("JSON-RPC method = %q, want %q", rpcRequest.Method, "broadcast_tx_commit")
			http.Error(w, "unexpected method", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%s}`, rpcRequest.ID, transferCommitResult); err != nil {
			t.Errorf("writing broadcast_tx_commit response: %v", err)
		}
	}))
	t.Cleanup(node.Close)

	cli, err := tm2Client.NewClient(node.URL)
	require.NoError(t, err)

	res, err := cli.SendTransactionCommit(&std.Tx{})
	require.NoError(t, err)
	assert.Len(t, res.DeliverTx.Events, 1)
}
