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

	// Byte-boundary bit positions — the BitArray's internal []uint64
	// representation packs 64 bits per word, so index 63↔64 and 127↔128
	// cross word boundaries that past bugs have been sensitive to.
	boundary := bitarray.NewBitArray(256)
	for _, i := range []int{63, 64, 127, 128, 255} {
		boundary.SetIndex(i, true)
	}

	// Size exactly 127 and 128 exercise the varint length-prefix
	// boundary at the top-level wire encoding.
	b127 := bitarray.NewBitArray(127)
	b127.SetIndex(0, true)
	b127.SetIndex(126, true)
	b128 := bitarray.NewBitArray(128)
	b128.SetIndex(0, true)
	b128.SetIndex(127, true)

	cases := []struct {
		name string
		v    any
	}{
		{"BitArray/empty", bitarray.NewBitArray(8)},
		{"BitArray/populated", ba},
		{"BitArray/word-boundaries", boundary},
		{"BitArray/size-127", b127},
		{"BitArray/size-128", b128},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
