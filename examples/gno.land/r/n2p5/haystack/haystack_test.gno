package haystack

import (
	"encoding/hex"
	"testing"

	"gno.land/p/n2p5/haystack"
	"gno.land/p/n2p5/haystack/needle"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
func TestHaystack(cur realm, t *testing.T) {
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
		urequire.NotPanics(t, cur, func() { Add(cross(cur), n1) })
		urequire.AbortsWithMessage(t, cur,
			haystack.ErrorDuplicateNeedle.Error(),
			func() {
				Add(cross(cur), n1)
			})
		testing.SetOriginCaller(u2)
		urequire.NotPanics(t, cur, func() { Add(cross(cur), n2) })
		urequire.NotPanics(t, cur, func() { Add(cross(cur), n3) })
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		n1, h1 := genNeedleHex(4)
		_, h2 := genNeedleHex(5)

		testing.SetOriginCaller(u1)
		urequire.NotPanics(t, cur, func() { Add(cross(cur), n1) })
		urequire.NotPanics(t, cur, func() {
			result := Get(h1)
			urequire.Equal(t, n1, result)
		})

		testing.SetOriginCaller(u2)
		urequire.NotPanics(t, cur, func() {
			result := Get(h1)
			urequire.Equal(t, n1, result)
		})
		urequire.PanicsWithMessage(t, cur,
			haystack.ErrorNeedleNotFound.Error(),
			func() {
				Get(h2)
			})
	})
}
