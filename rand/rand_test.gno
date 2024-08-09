package rand

import (
	"testing"
)

// making sure the generation isn't stuck to
// the same number.
//
// Giving it a few iterations to change, or failure.
//
// As all functions come from crypto/chacha20/rand
// we don't provide other tests here
func TestRand(t *testing.T) {
	var prev int = Intn(1 << 30)
	for i := 0; i < 20; i++ {
		if Intn(1<<30) != prev {
			return
		}
	}
	t.Errorf("no randomness, always getting %v", prev)
	t.Fail()
}
