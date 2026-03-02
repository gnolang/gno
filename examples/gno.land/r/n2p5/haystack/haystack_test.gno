package haystack

import (
	"encoding/hex"
	"testing"

	"gno.land/p/n2p5/haystack"
	"gno.land/p/n2p5/haystack/needle"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestHaystack(t *testing.T) {
	t.Parallel()
	// needleHex returns a hex-encoded needle and its hash for a given index.
	genNeedleHex := func(i int) (string, string) {
		b := make([]byte, needle.PayloadLength)
		b[0] = byte(i)
		n, _ := needle.New(b)
		return hex.EncodeToString(n.Bytes()), hex.EncodeToString(n.Hash())
	}

	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	t.Run("Add", func(t *testing.T) {
		t.Parallel()

		n1, _ := genNeedleHex(1)
		n2, _ := genNeedleHex(2)
		n3, _ := genNeedleHex(3)

		testing.SetOriginCaller(u1)
		urequire.NotPanics(t, func() { Add(cross, n1) })
		urequire.AbortsWithMessage(t,
			haystack.ErrorDuplicateNeedle.Error(),
			func() {
				Add(cross, n1)
			})
		testing.SetOriginCaller(u2)
		urequire.NotPanics(t, func() { Add(cross, n2) })
		urequire.NotPanics(t, func() { Add(cross, n3) })
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		n1, h1 := genNeedleHex(4)
		_, h2 := genNeedleHex(5)

		testing.SetOriginCaller(u1)
		urequire.NotPanics(t, func() { Add(cross, n1) })
		urequire.NotPanics(t, func() {
			result := Get(h1)
			urequire.Equal(t, n1, result)
		})

		testing.SetOriginCaller(u2)
		urequire.NotPanics(t, func() {
			result := Get(h1)
			urequire.Equal(t, n1, result)
		})
		urequire.PanicsWithMessage(t,
			haystack.ErrorNeedleNotFound.Error(),
			func() {
				Get(h2)
			})
	})
}
