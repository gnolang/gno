package testutils

import (
	"testing"
)

func TestTestAddress(t *testing.T) {
	testAddr := TestAddress("author1")
	if string(testAddr) != "g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6" {
		panic("not equal")
	}
}
