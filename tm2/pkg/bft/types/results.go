package types

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
)

//-----------------------------------------------------------------------------

// ABCIResult is the deterministic component of a ResponseDeliverTx.
// TODO: add tags and other fields
// https://github.com/tendermint/tendermint/issues/1007
type ABCIResult struct {
	Error  abci.Error   `json:"error"`
	Data   []byte       `json:"data"`
	Events []abci.Event `json:"events"`
}

// Bytes returns the amino encoded ABCIResult
func (a ABCIResult) Bytes() []byte {
	return bytesOrNil(a)
}

// ABCIResults wraps the deliver tx results to return a proof
type ABCIResults []ABCIResult

// NewResults creates ABCIResults from the list of ResponseDeliverTx.
func NewResults(responses []abci.ResponseDeliverTx) ABCIResults {
	res := make(ABCIResults, len(responses))
	for i, d := range responses {
		res[i] = NewResultFromResponse(d)
	}
	return res
}

// NewResultFromResponse creates ABCIResult from ResponseDeliverTx.
func NewResultFromResponse(response abci.ResponseDeliverTx) ABCIResult {
	return ABCIResult{
		Error:  response.Error,
		Data:   response.Data,
		Events: response.Events,
	}
}

// Bytes serializes the ABCIResponse using amino
func (a ABCIResults) Bytes() []byte {
	bz, err := amino.MarshalSized(a) // XXX: not length-prefixed
	if err != nil {
		panic(err)
	}
	return bz
}

// Hash returns a merkle hash of all results
func (a ABCIResults) Hash() []byte {
	// NOTE: we copy the impl of the merkle tree for txs -
	// we should be consistent and either do it for both or not.
	return merkle.SimpleHashFromByteSlices(a.toByteSlices())
}

// ProveResult returns a merkle proof of one result from the set
func (a ABCIResults) ProveResult(i int) merkle.SimpleProof {
	_, proofs := merkle.SimpleProofsFromByteSlices(a.toByteSlices())
	return *proofs[i]
}

func (a ABCIResults) toByteSlices() [][]byte {
	l := len(a)
	bzs := make([][]byte, l)
	for i := range l {
		bzs[i] = a[i].Bytes()
	}
	return bzs
}
