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
		if IsBlockRequestReprEmpty(goo) {
			var pbov *blockchainpb.BlockRequest
			msg = pbov
			return
		}
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
func IsBlockRequestReprEmpty(goor bcBlockRequestMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
				return false
			}
		}
	}
	return
}
func (goo bcBlockResponseMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.BlockResponse
	{
		if IsBlockResponseReprEmpty(goo) {
			var pbov *blockchainpb.BlockResponse
			msg = pbov
			return
		}
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
func IsBlockResponseReprEmpty(goor bcBlockResponseMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Block != nil {
				return false
			}
		}
	}
	return
}
func (goo bcNoBlockResponseMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.NoBlockResponse
	{
		if IsNoBlockResponseReprEmpty(goo) {
			var pbov *blockchainpb.NoBlockResponse
			msg = pbov
			return
		}
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
func IsNoBlockResponseReprEmpty(goor bcNoBlockResponseMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
				return false
			}
		}
	}
	return
}
func (goo bcStatusRequestMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.StatusRequest
	{
		if IsStatusRequestReprEmpty(goo) {
			var pbov *blockchainpb.StatusRequest
			msg = pbov
			return
		}
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
func IsStatusRequestReprEmpty(goor bcStatusRequestMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
				return false
			}
		}
	}
	return
}
func (goo bcStatusResponseMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *blockchainpb.StatusResponse
	{
		if IsStatusResponseReprEmpty(goo) {
			var pbov *blockchainpb.StatusResponse
			msg = pbov
			return
		}
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
func IsStatusResponseReprEmpty(goor bcStatusResponseMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
				return false
			}
		}
	}
	return
}
