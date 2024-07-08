package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

func TestABCIResults(t *testing.T) {
	t.Parallel()

	a := ABCIResult{Error: nil, Data: nil}
	b := ABCIResult{Error: nil, Data: []byte{}}
	c := ABCIResult{Error: nil, Data: []byte("one")}
	d := ABCIResult{Error: abci.StringError("?"), Data: nil}
	e := ABCIResult{Error: abci.StringError("?"), Data: []byte("foo")}
	f := ABCIResult{Error: abci.StringError("?"), Data: []byte("bar")}

	// Nil and []byte{} should produce the same bytes
	require.Equal(t, a.Bytes(), a.Bytes())
	require.Equal(t, b.Bytes(), b.Bytes())
	require.Equal(t, a.Bytes(), b.Bytes())

	// a and b should be the same, don't go in results.
	results := ABCIResults{a, c, d, e, f}

	// Make sure each result serializes differently
	var last []byte
	assert.Equal(t, last, a.Bytes()) // first one is empty
	for i, res := range results[1:] {
		bz := res.Bytes()
		assert.NotEqual(t, last, bz, "%d", i)
		last = bz
	}

	// Make sure that we can get a root hash from results and verify proofs.
	root := results.Hash()
	assert.NotEmpty(t, root)

	for i, res := range results {
		proof := results.ProveResult(i)
		valid := proof.Verify(root, res.Bytes())
		assert.NoError(t, valid, "%d", i)
	}
}

func TestABCIResultsBytes(t *testing.T) {
	t.Parallel()

	results := NewResults([]abci.ResponseDeliverTx{
		{ResponseBase: abci.ResponseBase{Error: nil, Data: []byte{}}},
		{ResponseBase: abci.ResponseBase{Error: nil, Data: []byte("one")}},
		{ResponseBase: abci.ResponseBase{Error: abci.StringError("?"), Data: nil}},
		{ResponseBase: abci.ResponseBase{Error: abci.StringError("?"), Data: []byte("foo")}},
		{ResponseBase: abci.ResponseBase{Error: abci.StringError("?"), Data: []byte("bar")}},
	})
	assert.NotNil(t, results.Bytes())
}
