package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/testutils"
)

func makeTxs(cnt, size int) Txs {
	txs := make(Txs, cnt)
	for i := range cnt {
		txs[i] = random.RandBytes(size)
	}
	return txs
}

func randInt(low, high int) int {
	off := random.RandInt() % (high - low)
	return low + off
}

func TestTxIndex(t *testing.T) {
	t.Parallel()

	for range 20 {
		txs := makeTxs(15, 60)
		for j := range txs {
			tx := txs[j]
			idx := txs.Index(tx)
			assert.Equal(t, j, idx)
		}
		assert.Equal(t, -1, txs.Index(nil))
		assert.Equal(t, -1, txs.Index(Tx("foodnwkf")))
	}
}

func TestTxIndexByHash(t *testing.T) {
	t.Parallel()

	for range 20 {
		txs := makeTxs(15, 60)
		for j := range txs {
			tx := txs[j]
			idx := txs.IndexByHash(tx.Hash())
			assert.Equal(t, j, idx)
		}
		assert.Equal(t, -1, txs.IndexByHash(nil))
		assert.Equal(t, -1, txs.IndexByHash(Tx("foodnwkf").Hash()))
	}
}

func TestValidTxProof(t *testing.T) {
	t.Parallel()

	cases := []struct {
		txs Txs
	}{
		{Txs{{1, 4, 34, 87, 163, 1}}},
		{Txs{{5, 56, 165, 2}, {4, 77}}},
		{Txs{Tx("foo"), Tx("bar"), Tx("baz")}},
		{makeTxs(20, 5)},
		{makeTxs(7, 81)},
		{makeTxs(61, 15)},
	}

	for h, tc := range cases {
		txs := tc.txs
		root := txs.Hash()
		// make sure valid proof for every tx
		for i := range txs {
			tx := []byte(txs[i])
			proof := txs.Proof(i)
			assert.Equal(t, i, proof.Proof.Index, "%d: %d", h, i)
			assert.Equal(t, len(txs), proof.Proof.Total, "%d: %d", h, i)
			assert.EqualValues(t, root, proof.RootHash, "%d: %d", h, i)
			assert.EqualValues(t, tx, proof.Data, "%d: %d", h, i)
			assert.EqualValues(t, txs[i].Hash(), proof.Leaf(), "%d: %d", h, i)
			assert.Nil(t, proof.Validate(root), "%d: %d", h, i)
			assert.NotNil(t, proof.Validate([]byte("foobar")), "%d: %d", h, i)

			// read-write must also work
			var p2 TxProof
			bin, err := amino.MarshalSized(proof)
			assert.Nil(t, err)
			err = amino.UnmarshalSized(bin, &p2)
			if assert.Nil(t, err, "%d: %d: %+v", h, i, err) {
				assert.Nil(t, p2.Validate(root), "%d: %d", h, i)
			}
		}
	}
}

func TestTxProofUnchangable(t *testing.T) {
	t.Parallel()

	// run the other test a bunch...
	for range 40 {
		testTxProofUnchangable(t)
	}
}

func testTxProofUnchangable(t *testing.T) {
	t.Helper()

	// make some proof
	txs := makeTxs(randInt(2, 100), randInt(16, 128))
	root := txs.Hash()
	i := randInt(0, len(txs)-1)
	proof := txs.Proof(i)

	// make sure it is valid to start with
	assert.Nil(t, proof.Validate(root))
	bin, err := amino.MarshalSized(proof)
	assert.Nil(t, err)

	// try mutating the data and make sure nothing breaks
	for range 500 {
		bad := testutils.MutateByteSlice(bin)
		if !bytes.Equal(bad, bin) {
			assertBadProof(t, root, bad, proof)
		}
	}
}

// This makes sure that the proof doesn't deserialize into something valid.
func assertBadProof(t *testing.T, root []byte, bad []byte, good TxProof) {
	t.Helper()

	var proof TxProof
	err := amino.UnmarshalSized(bad, &proof)
	if err == nil {
		err = proof.Validate(root)
		if err == nil {
			// XXX Fix simple merkle proofs so the following is *not* OK.
			// This can happen if we have a slightly different total (where the
			// path ends up the same). If it is something else, we have a real
			// problem.
			assert.NotEqual(t, proof.Proof.Total, good.Proof.Total, "bad: %#v\ngood: %#v", proof, good)
		}
	}
}
