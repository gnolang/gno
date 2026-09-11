package types

import (
	"bytes"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

var (
	valEd25519   = []string{"/tm.PubKeyEd25519"}
	valSecp256k1 = []string{"/tm.PubKeySecp256k1"}
)

func TestConsensusParamsValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		params abci.ConsensusParams
		valid  bool
	}{
		// test block params
		0: {makeParams(1, 1024, 0, 10, valEd25519), true},
		1: {makeParams(0, 1024, 0, 10, valEd25519), false},
		// MaxTxBytes has to leave MaxBlockOverheadBytes inside MaxDataBytes, and
		// makeParams pins MaxDataBytes at the 2MB default, so these three are
		// rejected for that reason -- a tx larger than a block's data budget
		// could never be included in one.
		2: {makeParams(47*1024*1024, 47*1024*1024+1024, 0, 10, valEd25519), false},
		3: {makeParams(10, 1024, 0, 10, valEd25519), true},
		4: {makeParams(100*1024*1024, 100*1024*1024+1024, 0, 10, valEd25519), false},
		5: {makeParams(101*1024*1024, 101*1024*1024+1024, 0, 10, valEd25519), false},
		6: {makeParams(1024*1024*1024, 1024*1024*1024+1024, 0, 10, valEd25519), false},
		7: {makeParams(1024*1024*1024, 1024*1024*1024+1024, 0, 10, valEd25519), false},
		8: {makeParams(1, 1024, 0, -10, valEd25519), false},
		// test no pubkey type provided
		9: {makeParams(1, 1024, 0, 10, []string{}), false},
		// test invalid pubkey type provided
		10: {makeParams(1, 1024, 0, 10, []string{"potatoes make good pubkeys"}), false},
		// secp256k1 is not supported for validators
		11: {makeParams(1, 1024, 0, 10, valSecp256k1), false},
	}
	for i, tc := range testCases {
		if tc.valid {
			assert.NoErrorf(t, ValidateConsensusParams(tc.params), "expected no error for valid params (#%d)", i)
		} else {
			assert.Errorf(t, ValidateConsensusParams(tc.params), "expected error for non valid params (#%d)", i)
		}
	}
}

func TestConsensusParamsValidationMaxDataBytes(t *testing.T) {
	t.Parallel()

	newParams := func(maxDataBytes int64) abci.ConsensusParams {
		return abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:   1024,
				MaxDataBytes: maxDataBytes,
				MaxGas:       10,
				TimeIotaMS:   10,
			},
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: valEd25519},
		}
	}

	// At or below the limit is accepted; above it is rejected so the chain can
	// never produce blocks larger than the fast-sync message envelope.
	assert.NoError(t, ValidateConsensusParams(newParams(MaxBlockDataBytes)))
	assert.NoError(t, ValidateConsensusParams(newParams(MaxBlockDataBytesLimit)))
	assert.Error(t, ValidateConsensusParams(newParams(MaxBlockDataBytesLimit+1)))

	// Non-positive values must be rejected too. 0 panics the proposer in
	// ReapMaxBytesMaxGas, and a negative value disables the reaping limit
	// altogether (bypassing the ceiling above) while panicking amino when the
	// consensus state decodes a proposal block with it as the max size.
	assert.Error(t, ValidateConsensusParams(newParams(0)))
	assert.Error(t, ValidateConsensusParams(newParams(-1)))
}

func TestConsensusParamsValidationMaxTxBytes(t *testing.T) {
	t.Parallel()

	newParams := func(maxTxBytes, maxDataBytes int64) abci.ConsensusParams {
		return abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:   maxTxBytes,
				MaxDataBytes: maxDataBytes,
				MaxGas:       10,
				TimeIotaMS:   10,
			},
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: valEd25519},
		}
	}

	// MaxDataBytes bounds the whole serialized block, so a MaxTxBytes-sized tx
	// has to leave room for the header and the LastCommit. Right at the boundary
	// is accepted; one byte over is not.
	assert.NoError(t, ValidateConsensusParams(
		newParams(MaxBlockDataBytes-MaxBlockOverheadBytes, MaxBlockDataBytes)))
	assert.Error(t, ValidateConsensusParams(
		newParams(MaxBlockDataBytes-MaxBlockOverheadBytes+1, MaxBlockDataBytes)))

	// The defaults have to satisfy it, and so does the largest pair the ceiling
	// permits -- that pair is what the per-connection recv budget is sized
	// against (see TestDefaultBudgetCoversWorstLegalConfig in bft/node).
	assert.NoError(t, ValidateConsensusParams(newParams(MaxBlockTxBytes, MaxBlockDataBytes)))
	assert.NoError(t, ValidateConsensusParams(
		newParams(MaxBlockDataBytesLimit-MaxBlockOverheadBytes, MaxBlockDataBytesLimit)))

	// A tx budget larger than the block data budget can never be satisfied: the
	// tx is admitted by CheckTx, reaped on its own, then trimmed out of every
	// proposal, starving itself and everything queued behind it.
	assert.Error(t, ValidateConsensusParams(newParams(MaxBlockDataBytes, MaxBlockDataBytes)))
}

func makeParams(
	txBytes, blockBytes, blockGas int64,
	blockTimeIotaMS int64,
	pubkeyTypeURLs []string,
) abci.ConsensusParams {
	return abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes: txBytes,
			// MaxDataBytes is not varied by this helper, but it must be positive
			// for the params to validate at all.
			MaxDataBytes:  MaxBlockDataBytes,
			MaxBlockBytes: blockBytes,
			MaxGas:        blockGas,
			TimeIotaMS:    blockTimeIotaMS,
		},
		Validator: &abci.ValidatorParams{
			PubKeyTypeURLs: pubkeyTypeURLs,
		},
	}
}

func TestConsensusParamsHash(t *testing.T) {
	t.Parallel()

	params := []abci.ConsensusParams{
		makeParams(4, 1024, 2, 10, valEd25519),
		makeParams(1, 1024, 4, 10, valEd25519),
		makeParams(1, 1024, 2, 10, valEd25519),
		makeParams(2, 1024, 5, 10, valEd25519),
		makeParams(1, 1024, 7, 10, valEd25519),
		makeParams(9, 1024, 5, 10, valEd25519),
		makeParams(7, 1024, 8, 10, valEd25519),
		makeParams(4, 1024, 6, 10, valEd25519),
	}

	hashes := make([][]byte, len(params))
	for i := range params {
		hashes[i] = params[i].Hash()
	}

	// make sure there are no duplicates...
	// sort, then check in order for matches
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i], hashes[j]) < 0
	})
	for i := range len(hashes) - 1 {
		assert.NotEqual(t, hashes[i], hashes[i+1])
	}
}

func TestConsensusParamsUpdate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		params        abci.ConsensusParams
		updates       abci.ConsensusParams
		updatedParams abci.ConsensusParams
	}{
		// empty updates
		{
			makeParams(1, 1024, 2, 10, valEd25519),
			abci.ConsensusParams{},
			makeParams(1, 1024, 2, 10, valEd25519),
		},
		// fine updates
		{
			makeParams(1, 1024, 2, 10, valEd25519),
			abci.ConsensusParams{
				Block: &abci.BlockParams{
					MaxTxBytes:    100,
					MaxDataBytes:  MaxBlockDataBytes,
					MaxBlockBytes: 1024,
					MaxGas:        200,
					TimeIotaMS:    10,
				},
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: valSecp256k1,
				},
			},
			makeParams(100, 1024, 200, 10, valSecp256k1),
		},
	}
	for _, tc := range testCases {
		assert.Equal(t, tc.updatedParams, tc.params.Update(tc.updates))
	}
}
