package gnoland

import (
	"errors"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateGasReplayMode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		mode    string
		wantErr bool
	}{
		{"", false},
		{"strict", false},
		{"source", false},
		{"max", true},    // not implemented yet
		{"skip", true},   // not implemented yet
		{"STRICT", true}, // case-sensitive
		{"garbage", true},
	} {
		tc := tc
		t.Run(tc.mode, func(t *testing.T) {
			t.Parallel()
			err := validateGasReplayMode(tc.mode)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReplayReport_Categorization(t *testing.T) {
	t.Parallel()

	r := newReplayReport("source")

	// Tx 0: success, gas differs from source
	r.recordDeliverResult(0, &GnoTxMetadata{BlockHeight: 10, GasUsed: 50_000}, sdk.Result{
		GasUsed: 75_000,
	})
	// Tx 1: success, gas matches source
	r.recordDeliverResult(1, &GnoTxMetadata{BlockHeight: 11, GasUsed: 30_000}, sdk.Result{
		GasUsed: 30_000,
	})
	// Tx 2: success, no source gas recorded
	r.recordDeliverResult(2, &GnoTxMetadata{BlockHeight: 12}, sdk.Result{
		GasUsed: 10_000,
	})
	// Tx 3: delivery failed
	r.recordDeliverResult(3, &GnoTxMetadata{BlockHeight: 13, GasUsed: 20_000}, sdk.Result{
		ResponseBase: abci.ResponseBase{
			Error: abci.StringError("out of gas"),
		},
		GasUsed: 20_000,
	})
	// Tx 4: skipped (source failed)
	r.record(4, &GnoTxMetadata{BlockHeight: 14, Failed: true}, 0, 0, ReplayCategorySkippedFailed, nil)

	outcomes := r.Outcomes()
	require.Len(t, outcomes, 5)
	assert.Equal(t, ReplayCategoryOKGasDiffers, outcomes[0].Category)
	assert.Equal(t, ReplayCategoryOK, outcomes[1].Category)
	assert.Equal(t, ReplayCategoryOK, outcomes[2].Category)
	assert.Equal(t, ReplayCategoryFailed, outcomes[3].Category)
	assert.Contains(t, outcomes[3].Error, "out of gas")
	assert.Equal(t, ReplayCategorySkippedFailed, outcomes[4].Category)

	// Explicit record with error
	r.record(5, &GnoTxMetadata{BlockHeight: 15}, 0, 0, ReplayCategoryFailed, errors.New("boom"))
	outcomes = r.Outcomes()
	require.Len(t, outcomes, 6)
	assert.Equal(t, "boom", outcomes[5].Error)
}

func TestReplayReport_ModeDefault(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "strict", modeOrDefault(""))
	assert.Equal(t, "source", modeOrDefault("source"))
}
