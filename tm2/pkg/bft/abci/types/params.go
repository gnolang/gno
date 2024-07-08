package abci

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
)

func (params ValidatorParams) IsValidPubKeyTypeURL(pubKeyTypeURL string) bool {
	for i := 0; i < len(params.PubKeyTypeURLs); i++ {
		if params.PubKeyTypeURLs[i] == pubKeyTypeURL {
			return true
		}
	}
	return false
}

func (params ConsensusParams) Hash() []byte {
	hasher := tmhash.New()
	bz := amino.MustMarshal(params)
	hasher.Write(bz)
	return hasher.Sum(nil)
}

func (params ConsensusParams) Update(params2 ConsensusParams) ConsensusParams {
	res := params // explicit copy

	// Nil subparams get left alone.
	// Non-nil subparams get copied whole.
	if params2.Block != nil {
		res.Block = amino.DeepCopy(params2.Block).(*BlockParams)
	}
	if params2.Validator != nil {
		res.Validator = amino.DeepCopy(params2.Validator).(*ValidatorParams)
	}

	return res
}
