package gnoland_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCodecParity_Gnoland(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(ed25519.Package)
	cdc.RegisterPackage(std.Package)
	cdc.RegisterPackage(auth.Package)
	cdc.RegisterPackage(bank.Package)
	cdc.RegisterPackage(vm.Package)
	cdc.RegisterPackage(gnoland.Package)
	cdc.Seal()

	addr := crypto.AddressFromPreimage([]byte("gno-account"))
	pk := ed25519.PubKeyEd25519{0x01, 0x02}

	cases := []struct {
		name string
		v    any
	}{
		{"GnoAccount", &gnoland.GnoAccount{
			BaseAccount: std.BaseAccount{
				Address:       addr,
				PubKey:        pk,
				AccountNumber: 1,
				Sequence:      5,
			},
		}},
		{"GnoTxMetadata", &gnoland.GnoTxMetadata{Timestamp: 1700000000}},
		{"TxWithMetadata", &gnoland.TxWithMetadata{
			Tx: std.Tx{
				Fee:  std.Fee{GasWanted: 100000, GasFee: std.Coin{Denom: "ugnot", Amount: 1000}},
				Memo: "genesis tx",
			},
			Metadata: &gnoland.GnoTxMetadata{Timestamp: 1700000000},
		}},
		{"GnoGenesisState/empty", &gnoland.GnoGenesisState{}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
