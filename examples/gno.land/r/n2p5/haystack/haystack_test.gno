package haystack_test

import (
	"encoding/hex"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/n2p5/haystack/needle"
	h "gno.land/r/n2p5/haystack"
)

func TestHaystackBasic(t *testing.T) {
	// Reset storage before running tests
	h.ResetStorage()

	// needleHex returns a hex-encoded needle and its hash for a given index.
	genNeedleHex := func(i int) (string, string) {
		b := make([]byte, needle.PayloadLength)
		b[0] = byte(i)
		n, _ := needle.New(b)
		return hex.EncodeToString(n.Bytes()), hex.EncodeToString(n.Hash())
	}

	u1 := testutils.TestAddress("u1")

	// Test Add and Get functionality
	n1, h1 := genNeedleHex(1)
	n2, h2 := genNeedleHex(2)

	// Add first needle
	testing.SetOriginCaller(u1)
	h.Add(n1)

	// Get the needle back
	result := h.Get(h1)
	uassert.Equal(t, n1, result)

	// Add second needle
	h.Add(n2)

	// Get both needles
	result = h.Get(h1)
	uassert.Equal(t, n1, result)

	result = h.Get(h2)
	uassert.Equal(t, n2, result)
}
