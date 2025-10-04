package transfer_test

import (
	"encoding/hex"
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
	"gno.land/r/aib/ibc/apps/transfer"
)

func TestDecodeFungibleTokenPacketData(t *testing.T) {
	bz, err := hex.DecodeString("0a0575676e6f7412033130301a06673175736572220a61746f6e6531757365722a046d656d6f")
	urequire.NoError(t, err)
	var f transfer.FungibleTokenPacketData

	err = f.ProtoUnmarshal(bz)

	urequire.NoError(t, err)
	uassert.Equal(t, "ugnot", f.Denom)
	uassert.Equal(t, "100", f.Amount)
	uassert.Equal(t, "g1user", f.Sender)
	uassert.Equal(t, "atone1user", f.Receiver)
	uassert.Equal(t, "memo", f.Memo)
}
