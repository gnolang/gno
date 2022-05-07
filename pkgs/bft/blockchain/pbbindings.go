package blockchain

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	blockchainpb "github.com/gnolang/gno/pkgs/bft/blockchain/pb"
	typespb "github.com/gnolang/gno/pkgs/bft/types/pb"
	types "github.com/gnolang/gno/pkgs/bft/types"
)

func (goo bcBlockRequestMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.BlockRequest
	{
		pbo = new(blockchainpb.BlockRequest)
		{
			pbo.Height = int64(goo.Height)
		}
	}
	msg = pbo
	return
}
func (goo bcBlockRequestMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(blockchainpb.BlockRequest)
	msg = pbo
	return
}
func (goo *bcBlockRequestMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *blockchainpb.BlockRequest = msg.(*blockchainpb.BlockRequest)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
		}
	}
	return
}
func (_ bcBlockRequestMessage) GetTypeURL() (typeURL string) {
	return "/tm.BlockRequest"
}
func (goo bcBlockResponseMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.BlockResponse
	{
		pbo = new(blockchainpb.BlockResponse)
		{
			if goo.Block != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Block.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Block = pbom.(*typespb.Block)
				if pbo.Block == nil {
					pbo.Block = new(typespb.Block)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo bcBlockResponseMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(blockchainpb.BlockResponse)
	msg = pbo
	return
}
func (goo *bcBlockResponseMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *blockchainpb.BlockResponse = msg.(*blockchainpb.BlockResponse)
	{
		if pbo != nil {
			{
				if pbo.Block != nil {
					(*goo).Block = new(types.Block)
					err = (*goo).Block.FromPBMessage(cdc, pbo.Block)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ bcBlockResponseMessage) GetTypeURL() (typeURL string) {
	return "/tm.BlockResponse"
}
func (goo bcNoBlockResponseMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.NoBlockResponse
	{
		pbo = new(blockchainpb.NoBlockResponse)
		{
			pbo.Height = int64(goo.Height)
		}
	}
	msg = pbo
	return
}
func (goo bcNoBlockResponseMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(blockchainpb.NoBlockResponse)
	msg = pbo
	return
}
func (goo *bcNoBlockResponseMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *blockchainpb.NoBlockResponse = msg.(*blockchainpb.NoBlockResponse)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
		}
	}
	return
}
func (_ bcNoBlockResponseMessage) GetTypeURL() (typeURL string) {
	return "/tm.NoBlockResponse"
}
func (goo bcStatusRequestMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.StatusRequest
	{
		pbo = new(blockchainpb.StatusRequest)
		{
			pbo.Height = int64(goo.Height)
		}
	}
	msg = pbo
	return
}
func (goo bcStatusRequestMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(blockchainpb.StatusRequest)
	msg = pbo
	return
}
func (goo *bcStatusRequestMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *blockchainpb.StatusRequest = msg.(*blockchainpb.StatusRequest)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
		}
	}
	return
}
func (_ bcStatusRequestMessage) GetTypeURL() (typeURL string) {
	return "/tm.StatusRequest"
}
func (goo bcStatusResponseMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.StatusResponse
	{
		pbo = new(blockchainpb.StatusResponse)
		{
			pbo.Height = int64(goo.Height)
		}
	}
	msg = pbo
	return
}
func (goo bcStatusResponseMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(blockchainpb.StatusResponse)
	msg = pbo
	return
}
func (goo *bcStatusResponseMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *blockchainpb.StatusResponse = msg.(*blockchainpb.StatusResponse)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
		}
	}
	return
}
func (_ bcStatusResponseMessage) GetTypeURL() (typeURL string) {
	return "/tm.StatusResponse"
}
