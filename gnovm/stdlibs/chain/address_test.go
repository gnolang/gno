package chain

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
)

// pubKeyAddress decodes its argument into crypto.PubKey through the global
// codec, so a payload can name merkle.SimpleProofNode, a type that contains
// itself. The decoder stops at the nesting limit, so it reads only the first
// few hundred bytes of a chain however long it is. The returned error names
// only the type mismatch, which replaces the depth error, so the test reads
// where the decoder stopped.
func TestPubKeyAddressStopsAtNestingLimit(t *testing.T) {
	t.Parallel()

	stoppedAfter := func(n int) int {
		var node *merkle.SimpleProofNode
		for range n {
			node = &merkle.SimpleProofNode{Parent: node}
		}
		bz, err := amino.MarshalAny(node)
		require.NoError(t, err)
		s, err := bech32.Encode(crypto.Bech32PubKeyPrefix(), bz)
		require.NoError(t, err)

		_, errStr := X_pubKeyAddress(s)
		m := regexp.MustCompile(`failed after (\d+) bytes \(decoded type merkle.SimpleProofNode is not assignable`).FindStringSubmatch(errStr)
		require.NotNil(t, m, errStr)
		off, err := strconv.Atoi(m[1])
		require.NoError(t, err)
		return off
	}

	// 64 levels of a few bytes each, however long the chain.
	for _, n := range []int{1000, 5000, 20000} {
		require.Less(t, stoppedAfter(n), 512, "chain of %d", n)
	}
}
