package bank

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	bankpb "github.com/gnolang/gno/pkgs/sdk/bank/pb"
)

func (goo NoInputsError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *bankpb.NoInputsError
	{
		pbo = new(bankpb.NoInputsError)
	}
	msg = pbo
	return
}
func (goo NoInputsError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(bankpb.NoInputsError)
	msg = pbo
	return
}
func (goo *NoInputsError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *bankpb.NoInputsError = msg.(*bankpb.NoInputsError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ NoInputsError) GetTypeURL() (typeURL string) {
	return "/bank.NoInputsError"
}
func (goo NoOutputsError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *bankpb.NoOutputsError
	{
		pbo = new(bankpb.NoOutputsError)
	}
	msg = pbo
	return
}
func (goo NoOutputsError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(bankpb.NoOutputsError)
	msg = pbo
	return
}
func (goo *NoOutputsError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *bankpb.NoOutputsError = msg.(*bankpb.NoOutputsError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ NoOutputsError) GetTypeURL() (typeURL string) {
	return "/bank.NoOutputsError"
}
func (goo InputOutputMismatchError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *bankpb.InputOutputMismatchError
	{
		pbo = new(bankpb.InputOutputMismatchError)
	}
	msg = pbo
	return
}
func (goo InputOutputMismatchError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(bankpb.InputOutputMismatchError)
	msg = pbo
	return
}
func (goo *InputOutputMismatchError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *bankpb.InputOutputMismatchError = msg.(*bankpb.InputOutputMismatchError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InputOutputMismatchError) GetTypeURL() (typeURL string) {
	return "/bank.InputOutputMismatchError"
}
func (goo MsgSend) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *bankpb.MsgSend
	{
		pbo = new(bankpb.MsgSend)
		{
			goor, err1 := goo.FromAddress.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.FromAddress = string(goor)
		}
		{
			goor, err1 := goo.ToAddress.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.ToAddress = string(goor)
		}
		{
			goor, err1 := goo.Amount.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Amount = string(goor)
		}
	}
	msg = pbo
	return
}
func (goo MsgSend) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(bankpb.MsgSend)
	msg = pbo
	return
}
func (goo *MsgSend) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *bankpb.MsgSend = msg.(*bankpb.MsgSend)
	{
		if pbo != nil {
			{
				var goor string
				goor = string(pbo.FromAddress)
				err = (*goo).FromAddress.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				var goor string
				goor = string(pbo.ToAddress)
				err = (*goo).ToAddress.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				var goor string
				goor = string(pbo.Amount)
				err = (*goo).Amount.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ MsgSend) GetTypeURL() (typeURL string) {
	return "/bank.MsgSend"
}
