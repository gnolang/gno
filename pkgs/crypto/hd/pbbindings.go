package hd

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	hdpb "github.com/gnolang/gno/pkgs/crypto/hd/pb"
)

func (goo BIP44Params) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *hdpb.Bip44Params
	{
		pbo = new(hdpb.Bip44Params)
		{
			pbo.Purpose = uint32(goo.Purpose)
		}
		{
			pbo.CoinType = uint32(goo.CoinType)
		}
		{
			pbo.Account = uint32(goo.Account)
		}
		{
			pbo.Change = bool(goo.Change)
		}
		{
			pbo.AddressIndex = uint32(goo.AddressIndex)
		}
	}
	msg = pbo
	return
}
func (goo BIP44Params) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(hdpb.Bip44Params)
	msg = pbo
	return
}
func (goo *BIP44Params) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *hdpb.Bip44Params = msg.(*hdpb.Bip44Params)
	{
		if pbo != nil {
			{
				(*goo).Purpose = uint32(pbo.Purpose)
			}
			{
				(*goo).CoinType = uint32(pbo.CoinType)
			}
			{
				(*goo).Account = uint32(pbo.Account)
			}
			{
				(*goo).Change = bool(pbo.Change)
			}
			{
				(*goo).AddressIndex = uint32(pbo.AddressIndex)
			}
		}
	}
	return
}
func (_ BIP44Params) GetTypeURL() (typeURL string) {
	return "/tm.Bip44Params"
}
