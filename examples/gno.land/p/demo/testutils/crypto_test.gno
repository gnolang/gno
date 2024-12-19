package testutils
import (
	"fmt"
	"testing"
	"gno.land/p/demo/uassert"
)

func TestTestAddress(t *testing.T) {
	testAddr := TestAddress("author1")
	uassert.Equal(t, "g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6", string(testAddr))
}