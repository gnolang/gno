package auth_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCodecParity_Auth(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(ed25519.Package)
	cdc.RegisterPackage(std.Package)
	cdc.RegisterPackage(auth.Package)
	cdc.Seal()

	pk := ed25519.PubKeyEd25519{0x11, 0x22, 0x33}
	creator := crypto.AddressFromPreimage([]byte("creator"))

	cases := []struct {
		name string
		v    any
	}{
		{"Params", &auth.Params{
			MaxMemoBytes:              256,
			TxSigLimit:                7,
			TxSizeCostPerByte:         10,
			SigVerifyCostED25519:      590,
			SigVerifyCostSecp256k1:    1000,
			GasPricesChangeCompressor: 2,
			TargetGasRatio:            70,
		}},
		{"GenesisState", &auth.GenesisState{Params: auth.DefaultParams()}},
		{"MsgCreateSession/full", &auth.MsgCreateSession{
			Creator:     creator,
			SessionKey:  pk,
			ExpiresAt:   1700000000,
			AllowPaths:  []string{"gno.land/r/foo", "gno.land/r/bar"},
			SpendLimit:  std.Coins{{Denom: "ugnot", Amount: 100}},
			SpendPeriod: 3600,
		}},
		{"MsgRevokeSession", &auth.MsgRevokeSession{
			Creator:    creator,
			SessionKey: pk,
		}},
		{"MsgRevokeAllSessions", &auth.MsgRevokeAllSessions{
			Creator: creator,
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
