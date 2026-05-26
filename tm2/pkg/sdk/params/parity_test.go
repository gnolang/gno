package params_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
)

func TestCodecParity_SDKParams(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(params.Package)
	cdc.Seal()

	cases := []struct {
		name string
		v    any
	}{
		{"Param/string", &params.Param{Key: "max_tx_bytes", Type: "string", Value: "1048576"}},
		{"Param/int64", &params.Param{Key: "gas_price", Type: "int64", Value: int64(7)}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
