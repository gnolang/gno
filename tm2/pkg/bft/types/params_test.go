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
		2: {makeParams(47*1024*1024, 47*1024*1024+1024, 0, 10, valEd25519), true},
		3: {makeParams(10, 1024, 0, 10, valEd25519), true},

		4: {makeParams(100*1024*1024, 100*1024*1024+1024, 0, 10, valEd25519), true},
		5: {makeParams(101*1024*1024, 101*1024*1024+1024, 0, 10, valEd25519), false},
		6: {makeParams(1024*1024*1024, 1024*1024*1024+1024, 0, 10, valEd25519), false},
		7: {makeParams(1024*1024*1024, 1024*1024*1024+1024, 0, 10, valEd25519), false},
		8: {makeParams(1, 1024, 0, -10, valEd25519), false},
		// test no pubkey type provided
		9: {makeParams(1, 1024, 0, 10, []string{}), false},
		// test invalid pubkey type provided
		10: {makeParams(1, 1024, 0, 10, []string{"potatoes make good pubkeys"}), false},
	}
	for i, tc := range testCases {
		if tc.valid {
			assert.NoErrorf(t, ValidateConsensusParams(tc.params), "expected no error for valid params (#%d)", i)
		} else {
			assert.Errorf(t, ValidateConsensusParams(tc.params), "expected error for non valid params (#%d)", i)
		}
	}
}

func makeParams(
	dataBytes, blockBytes, blockGas int64,
	blockTimeIotaMS int64,
	pubkeyTypeURLs []string,
) abci.ConsensusParams {
	return abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:            dataBytes,
			MaxBlockBytes:         blockBytes,
			MaxGas:                blockGas,
			TimeIotaMS:            blockTimeIotaMS,
			PriceChangeCompressor: 1,
			TargetGas:             0,
			InitialGasPriceGas:    1,
			InitialGasPriceDenom:  "token",
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
	for i := 0; i < len(hashes)-1; i++ {
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
					MaxTxBytes:            100,
					MaxBlockBytes:         1024,
					MaxGas:                200,
					TimeIotaMS:            10,
					PriceChangeCompressor: 1,
					TargetGas:             0,
					InitialGasPriceGas:    1,
					InitialGasPriceDenom:  "token",
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
