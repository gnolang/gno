package blockchain

// Internal test package: the reactor-message types (bcBlockRequestMessage etc.)
// are unexported and can't be reached from outside.

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	btypes "github.com/gnolang/gno/tm2/pkg/bft/types"
)

func TestCodecParity_Blockchain(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(btypes.Package)
	cdc.RegisterPackage(Package)
	cdc.Seal()

	cases := []struct {
		name string
		v    any
	}{
		{"bcBlockRequestMessage", &bcBlockRequestMessage{Height: 42}},
		{"bcStatusRequestMessage", &bcStatusRequestMessage{Height: 101}},
		{"bcNoBlockResponseMessage", &bcNoBlockResponseMessage{Height: 7}},
		{"bcStatusResponseMessage", &bcStatusResponseMessage{Height: 123}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
