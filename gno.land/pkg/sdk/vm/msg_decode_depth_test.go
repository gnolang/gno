package vm

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// A tx can carry a merkle.SimpleProofNode in MemPackage.Info, a type that
// contains itself; its nesting is refused at decode, before any fee is taken.
func TestTxDecodeRefusesDeepProofNodeInInfo(t *testing.T) {
	t.Parallel()

	txWithChain := func(n int) []byte {
		var node *merkle.SimpleProofNode
		for range n {
			node = &merkle.SimpleProofNode{Parent: node}
		}
		msg := NewMsgAddPackage(crypto.AddressFromPreimage([]byte("depth")), "gno.land/r/test/depth", nil)
		msg.Package.Info = node
		return amino.MustMarshal(std.Tx{Msgs: []std.Msg{msg}})
	}

	// The Msgs entry and the Info field are two Any levels of the 64 allowed.
	var tx std.Tx
	require.NoError(t, amino.Unmarshal(txWithChain(62), &tx))
	require.ErrorContains(t, amino.Unmarshal(txWithChain(63), &tx), "exceeded max nesting depth 64")
}
