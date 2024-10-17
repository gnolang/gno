package blake3

import (
	"crypto/blake3"
	"encoding/hex"
	"testing"
)

func TestSha256Sum(t *testing.T) {
	got := blake3.Sum256([]byte("sha256 this string"))[:]
	expected := "8bf6a3079eba11e7d5479e01531d9a08f2791bdd94e3e1d12d483e5bb6d6f570"

	if hex.EncodeToString(got) != expected {
		t.Errorf("got %v(%T), expected %v(%T)", hex.EncodeToString(got), got, expected, expected)
	}
}
