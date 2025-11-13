package transfer_test

import (
	"encoding/hex"
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
	"gno.land/r/aib/ibc/apps/transfer"
)

func TestFungibleTokenPacketDataProtoMarshal(t *testing.T) {
	f := transfer.FungibleTokenPacketData{
		Denom:    "ugnot",
		Amount:   "100",
		Sender:   "g1user",
		Receiver: "atone1user",
		Memo:     "memo",
	}

	bz := f.ProtoMarshal()

	expected := "0a0575676e6f7412033130301a06673175736572220a61746f6e6531757365722a046d656d6f"
	got := hex.EncodeToString(bz)
	uassert.Equal(t, expected, got)
}

func TestFungibleTokenPacketDataProtoUnmarshal(t *testing.T) {
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

func TestFungibleTokenPacketDataValidateBasic(t *testing.T) {
	const (
		sender             = "sender"
		receiver           = "receiver"
		denom              = "transfer/gaiachannel/atom"
		amount             = "100"
		largeAmount        = "18446744073709551616"                                                           // one greater than largest uint64 (^uint64(0))
		invalidLargeAmount = "115792089237316195423570985008687907853269984665640564039457584007913129639936" // 2^256
	)
	testCases := []struct {
		name       string
		packetData transfer.FungibleTokenPacketData
		expErr     string
	}{
		{
			name:       "valid packet",
			packetData: transfer.NewFungibleTokenPacketData(denom, amount, sender, receiver, ""),
			expErr:     "",
		},
		{
			name:       "valid packet with memo",
			packetData: transfer.NewFungibleTokenPacketData(denom, amount, sender, receiver, "memo"),
			expErr:     "",
		},
		{
			name:       "valid packet with large amount",
			packetData: transfer.NewFungibleTokenPacketData(denom, largeAmount, sender, receiver, ""),
			expErr:     "",
		},
		{
			name:       "invalid denom",
			packetData: transfer.NewFungibleTokenPacketData("", amount, sender, receiver, ""),
			expErr:     "base denomination cannot be blank",
		},
		{
			name:       "invalid denom, invalid portID",
			packetData: transfer.NewFungibleTokenPacketData("(transfer)/channel-1/uatom", amount, sender, receiver, ""),
			expErr:     "invalid trace",
		},
		{
			name:       "invalid empty amount",
			packetData: transfer.NewFungibleTokenPacketData(denom, "", sender, receiver, ""),
			expErr:     "unable to parse transfer amount \"\"",
		},
		{
			name:       "invalid zero amount",
			packetData: transfer.NewFungibleTokenPacketData(denom, "0", sender, receiver, ""),
			expErr:     "amount must be strictly positive: got 0",
		},
		{
			name:       "invalid negative amount",
			packetData: transfer.NewFungibleTokenPacketData(denom, "-1", sender, receiver, ""),
			expErr:     "amount must be strictly positive: got -1",
		},
		{
			name:       "invalid large amount",
			packetData: transfer.NewFungibleTokenPacketData(denom, invalidLargeAmount, sender, receiver, ""),
			expErr:     "unable to parse transfer amount \"" + invalidLargeAmount + "\"",
		},
		{
			name:       "missing sender address",
			packetData: transfer.NewFungibleTokenPacketData(denom, amount, "", receiver, ""),
			expErr:     "sender address cannot be blank",
		},
		{
			name:       "missing receiver address",
			packetData: transfer.NewFungibleTokenPacketData(denom, amount, sender, "", ""),
			expErr:     "receiver address cannot be blank",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.packetData.ValidateBasic()

			if tc.expErr == "" {
				urequire.NoError(t, err)
				return
			}
			urequire.ErrorContains(t, err, tc.expErr)
		})
	}
}
