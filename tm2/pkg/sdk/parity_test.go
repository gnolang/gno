package sdk_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func TestCodecParity_SDK(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(abci.Package)
	cdc.RegisterPackage(sdk.Package)
	cdc.Seal()

	cases := []struct {
		name string
		v    any
	}{
		{"Result/empty", &sdk.Result{}},
		{"Result/populated", &sdk.Result{
			ResponseBase: abci.ResponseBase{
				Data: []byte("result-data"),
				Info: "info-string",
			},
			GasWanted: 100000,
			GasUsed:   75000,
		}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
