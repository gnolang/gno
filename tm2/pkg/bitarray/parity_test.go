package bitarray_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
)

func TestCodecParity_BitArray(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(bitarray.Package)
	cdc.Seal()

	// A BitArray with some bits set at non-trivial positions.
	ba := bitarray.NewBitArray(100)
	for _, i := range []int{0, 1, 7, 63, 64, 99} {
		ba.SetIndex(i, true)
	}

	cases := []struct {
		name string
		v    any
	}{
		{"BitArray/empty", bitarray.NewBitArray(8)},
		{"BitArray/populated", ba},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
