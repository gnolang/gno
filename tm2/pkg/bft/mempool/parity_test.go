package mempool_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool"
	btypes "github.com/gnolang/gno/tm2/pkg/bft/types"
)

func TestCodecParity_Mempool(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(btypes.Package)
	cdc.RegisterPackage(mempool.Package)
	cdc.Seal()

	cases := []struct {
		name string
		v    any
	}{
		{"TxMessage/empty", &mempool.TxMessage{}},
		{"TxMessage/payload", &mempool.TxMessage{Tx: btypes.Tx([]byte{0xca, 0xfe, 0xba, 0xbe})}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
