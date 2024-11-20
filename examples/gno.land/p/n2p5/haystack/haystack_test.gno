package haystack

import (
	"encoding/hex"
	"testing"

	"gno.land/p/n2p5/haystack/needle"
)

func TestHaystack(t *testing.T) {
	t.Parallel()

	t.Run("New", func(t *testing.T) {
		t.Parallel()
		h := New()
		if h == nil {
			t.Error("New returned nil")
		}
	})

	t.Run("Add", func(t *testing.T) {
		t.Parallel()
		h := New()
		n, _ := needle.New(make([]byte, needle.PayloadLength))
		validNeedleHex := hex.EncodeToString(n.Bytes())

		testTable := []struct {
			needleHex string
			err       error
		}{
			{validNeedleHex, nil},
			{validNeedleHex, ErrorDuplicateNeedle},
			{"bad" + validNeedleHex[3:], needle.ErrorInvalidHash},
			{"XXX" + validNeedleHex[3:], hex.InvalidByteError('X')},
			{validNeedleHex[:len(validNeedleHex)-2], ErrorNeedleLength},
			{validNeedleHex + "00", ErrorNeedleLength},
			{"000", ErrorNeedleLength},
		}
		for _, tt := range testTable {
			err := h.Add(tt.needleHex)
			if err != tt.err {
				t.Error(tt.needleHex, err.Error(), "!=", tt.err.Error())
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		h := New()

		// genNeedleHex returns a hex-encoded needle and its hash for a given index.
		genNeedleHex := func(i int) (string, string) {
			b := make([]byte, needle.PayloadLength)
			b[0] = byte(i)
			n, _ := needle.New(b)
			return hex.EncodeToString(n.Bytes()), hex.EncodeToString(n.Hash())
		}

		// Add a valid needle to the haystack.
		validNeedleHex, validHash := genNeedleHex(0)
		h.Add(validNeedleHex)

		// Add a needle and break the value type.
		_, brokenHashValueType := genNeedleHex(1)
		h.internal.Set(brokenHashValueType, 0)

		// Add a needle with invalid hash.
		_, invalidHash := genNeedleHex(2)
		h.internal.Set(invalidHash, make([]byte, needle.PayloadLength))

		testTable := []struct {
			hash     string
			expected string
			err      error
		}{
			{validHash, validNeedleHex, nil},
			{validHash[:len(validHash)-2], "", ErrorHashLength},
			{validHash + "00", "", ErrorHashLength},
			{"XXX" + validHash[3:], "", hex.InvalidByteError('X')},
			{"bad" + validHash[3:], "", ErrorNeedleNotFound},
			{brokenHashValueType, "", ErrorValueInvalidType},
			{invalidHash, "", ErrorHashMismatch},
		}
		for _, tt := range testTable {
			actual, err := h.Get(tt.hash)
			if err != tt.err {
				t.Error(tt.hash, err.Error(), "!=", tt.err.Error())
			}
			if actual != tt.expected {
				t.Error(tt.hash, actual, "!=", tt.expected)
			}
		}
	})
}
