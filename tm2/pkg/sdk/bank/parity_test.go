package bank_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCodecParity_Bank(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(std.Package)
	cdc.RegisterPackage(bank.Package)
	cdc.Seal()

	from := crypto.AddressFromPreimage([]byte("from"))
	to := crypto.AddressFromPreimage([]byte("to"))

	cases := []struct {
		name string
		v    any
	}{
		{"Params", &bank.Params{RestrictedDenoms: []string{"ugnot", "ugno"}}},
		{"GenesisState", &bank.GenesisState{Params: bank.Params{RestrictedDenoms: nil}}},
		{"MsgSend", &bank.MsgSend{
			FromAddress: from,
			ToAddress:   to,
			// Keep coins in canonical (alphabetical) order — amino sorts on roundtrip.
			Amount: std.Coins{{Denom: "ugno", Amount: 50}, {Denom: "ugnot", Amount: 100}},
		}},
		{"NoInputsError", &bank.NoInputsError{}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
