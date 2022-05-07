package std

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	stdpb "github.com/gnolang/gno/pkgs/std/pb"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func (goo BaseAccount) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.BaseAccount
	{
		pbo = new(stdpb.BaseAccount)
		{
			goor, err1 := goo.Address.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Address = string(goor)
		}
		{
			goor, err1 := goo.Coins.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Coins = string(goor)
		}
		{
			if goo.PubKey != nil {
				typeUrl := cdc.GetTypeURL(goo.PubKey)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.PubKey)
				if err != nil {
					return
				}
				pbo.PubKey = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			pbo.AccountNumber = uint64(goo.AccountNumber)
		}
		{
			pbo.Sequence = uint64(goo.Sequence)
		}
	}
	msg = pbo
	return
}
func (goo BaseAccount) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.BaseAccount)
	msg = pbo
	return
}
func (goo *BaseAccount) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.BaseAccount = msg.(*stdpb.BaseAccount)
	{
		if pbo != nil {
			{
				var goor string
				goor = string(pbo.Address)
				err = (*goo).Address.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				var goor string
				goor = string(pbo.Coins)
				err = (*goo).Coins.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.PubKey.TypeUrl
				bz := pbo.PubKey.Value
				goorp := &(*goo).PubKey
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				(*goo).AccountNumber = uint64(pbo.AccountNumber)
			}
			{
				(*goo).Sequence = uint64(pbo.Sequence)
			}
		}
	}
	return
}
func (_ BaseAccount) GetTypeURL() (typeURL string) {
	return "/std.BaseAccount"
}
func (goo MemFile) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.MemFile
	{
		pbo = new(stdpb.MemFile)
		{
			pbo.Name = string(goo.Name)
		}
		{
			pbo.Body = string(goo.Body)
		}
	}
	msg = pbo
	return
}
func (goo MemFile) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.MemFile)
	msg = pbo
	return
}
func (goo *MemFile) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.MemFile = msg.(*stdpb.MemFile)
	{
		if pbo != nil {
			{
				(*goo).Name = string(pbo.Name)
			}
			{
				(*goo).Body = string(pbo.Body)
			}
		}
	}
	return
}
func (_ MemFile) GetTypeURL() (typeURL string) {
	return "/std.MemFile"
}
func (goo MemPackage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.MemPackage
	{
		pbo = new(stdpb.MemPackage)
		{
			pbo.Name = string(goo.Name)
		}
		{
			pbo.Path = string(goo.Path)
		}
		{
			goorl := len(goo.Files)
			if goorl == 0 {
				pbo.Files = nil
			} else {
				var pbos = make([]*stdpb.MemFile, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Files[i]
						{
							if goore != nil {
								pbom := proto.Message(nil)
								pbom, err = goore.ToPBMessage(cdc)
								if err != nil {
									return
								}
								pbos[i] = pbom.(*stdpb.MemFile)
								if pbos[i] == nil {
									pbos[i] = new(stdpb.MemFile)
								}
							}
						}
					}
				}
				pbo.Files = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo MemPackage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.MemPackage)
	msg = pbo
	return
}
func (goo *MemPackage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.MemPackage = msg.(*stdpb.MemPackage)
	{
		if pbo != nil {
			{
				(*goo).Name = string(pbo.Name)
			}
			{
				(*goo).Path = string(pbo.Path)
			}
			{
				var pbol int = 0
				if pbo.Files != nil {
					pbol = len(pbo.Files)
				}
				if pbol == 0 {
					(*goo).Files = nil
				} else {
					var goors = make([]*MemFile, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Files[i]
							{
								pboev := pboe
								if pboev != nil {
									goors[i] = new(MemFile)
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*goo).Files = goors
				}
			}
		}
	}
	return
}
func (_ MemPackage) GetTypeURL() (typeURL string) {
	return "/std.MemPackage"
}
func (goo InternalError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InternalError
	{
		pbo = new(stdpb.InternalError)
	}
	msg = pbo
	return
}
func (goo InternalError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.InternalError)
	msg = pbo
	return
}
func (goo *InternalError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.InternalError = msg.(*stdpb.InternalError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InternalError) GetTypeURL() (typeURL string) {
	return "/std.InternalError"
}
func (goo TxDecodeError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.TxDecodeError
	{
		pbo = new(stdpb.TxDecodeError)
	}
	msg = pbo
	return
}
func (goo TxDecodeError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.TxDecodeError)
	msg = pbo
	return
}
func (goo *TxDecodeError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.TxDecodeError = msg.(*stdpb.TxDecodeError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ TxDecodeError) GetTypeURL() (typeURL string) {
	return "/std.TxDecodeError"
}
func (goo InvalidSequenceError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InvalidSequenceError
	{
		pbo = new(stdpb.InvalidSequenceError)
	}
	msg = pbo
	return
}
func (goo InvalidSequenceError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.InvalidSequenceError)
	msg = pbo
	return
}
func (goo *InvalidSequenceError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.InvalidSequenceError = msg.(*stdpb.InvalidSequenceError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InvalidSequenceError) GetTypeURL() (typeURL string) {
	return "/std.InvalidSequenceError"
}
func (goo UnauthorizedError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.UnauthorizedError
	{
		pbo = new(stdpb.UnauthorizedError)
	}
	msg = pbo
	return
}
func (goo UnauthorizedError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.UnauthorizedError)
	msg = pbo
	return
}
func (goo *UnauthorizedError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.UnauthorizedError = msg.(*stdpb.UnauthorizedError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ UnauthorizedError) GetTypeURL() (typeURL string) {
	return "/std.UnauthorizedError"
}
func (goo InsufficientFundsError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InsufficientFundsError
	{
		pbo = new(stdpb.InsufficientFundsError)
	}
	msg = pbo
	return
}
func (goo InsufficientFundsError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.InsufficientFundsError)
	msg = pbo
	return
}
func (goo *InsufficientFundsError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.InsufficientFundsError = msg.(*stdpb.InsufficientFundsError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InsufficientFundsError) GetTypeURL() (typeURL string) {
	return "/std.InsufficientFundsError"
}
func (goo UnknownRequestError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.UnknownRequestError
	{
		pbo = new(stdpb.UnknownRequestError)
	}
	msg = pbo
	return
}
func (goo UnknownRequestError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.UnknownRequestError)
	msg = pbo
	return
}
func (goo *UnknownRequestError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.UnknownRequestError = msg.(*stdpb.UnknownRequestError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ UnknownRequestError) GetTypeURL() (typeURL string) {
	return "/std.UnknownRequestError"
}
func (goo InvalidAddressError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InvalidAddressError
	{
		pbo = new(stdpb.InvalidAddressError)
	}
	msg = pbo
	return
}
func (goo InvalidAddressError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.InvalidAddressError)
	msg = pbo
	return
}
func (goo *InvalidAddressError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.InvalidAddressError = msg.(*stdpb.InvalidAddressError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InvalidAddressError) GetTypeURL() (typeURL string) {
	return "/std.InvalidAddressError"
}
func (goo UnknownAddressError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.UnknownAddressError
	{
		pbo = new(stdpb.UnknownAddressError)
	}
	msg = pbo
	return
}
func (goo UnknownAddressError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.UnknownAddressError)
	msg = pbo
	return
}
func (goo *UnknownAddressError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.UnknownAddressError = msg.(*stdpb.UnknownAddressError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ UnknownAddressError) GetTypeURL() (typeURL string) {
	return "/std.UnknownAddressError"
}
func (goo InvalidPubKeyError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InvalidPubKeyError
	{
		pbo = new(stdpb.InvalidPubKeyError)
	}
	msg = pbo
	return
}
func (goo InvalidPubKeyError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.InvalidPubKeyError)
	msg = pbo
	return
}
func (goo *InvalidPubKeyError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.InvalidPubKeyError = msg.(*stdpb.InvalidPubKeyError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InvalidPubKeyError) GetTypeURL() (typeURL string) {
	return "/std.InvalidPubKeyError"
}
func (goo InsufficientCoinsError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InsufficientCoinsError
	{
		pbo = new(stdpb.InsufficientCoinsError)
	}
	msg = pbo
	return
}
func (goo InsufficientCoinsError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.InsufficientCoinsError)
	msg = pbo
	return
}
func (goo *InsufficientCoinsError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.InsufficientCoinsError = msg.(*stdpb.InsufficientCoinsError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InsufficientCoinsError) GetTypeURL() (typeURL string) {
	return "/std.InsufficientCoinsError"
}
func (goo OutOfGasError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.OutOfGasError
	{
		pbo = new(stdpb.OutOfGasError)
	}
	msg = pbo
	return
}
func (goo OutOfGasError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.OutOfGasError)
	msg = pbo
	return
}
func (goo *OutOfGasError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.OutOfGasError = msg.(*stdpb.OutOfGasError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ OutOfGasError) GetTypeURL() (typeURL string) {
	return "/std.OutOfGasError"
}
func (goo MemoTooLargeError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.MemoTooLargeError
	{
		pbo = new(stdpb.MemoTooLargeError)
	}
	msg = pbo
	return
}
func (goo MemoTooLargeError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.MemoTooLargeError)
	msg = pbo
	return
}
func (goo *MemoTooLargeError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.MemoTooLargeError = msg.(*stdpb.MemoTooLargeError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ MemoTooLargeError) GetTypeURL() (typeURL string) {
	return "/std.MemoTooLargeError"
}
func (goo InsufficientFeeError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InsufficientFeeError
	{
		pbo = new(stdpb.InsufficientFeeError)
	}
	msg = pbo
	return
}
func (goo InsufficientFeeError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.InsufficientFeeError)
	msg = pbo
	return
}
func (goo *InsufficientFeeError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.InsufficientFeeError = msg.(*stdpb.InsufficientFeeError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InsufficientFeeError) GetTypeURL() (typeURL string) {
	return "/std.InsufficientFeeError"
}
func (goo TooManySignaturesError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.TooManySignaturesError
	{
		pbo = new(stdpb.TooManySignaturesError)
	}
	msg = pbo
	return
}
func (goo TooManySignaturesError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.TooManySignaturesError)
	msg = pbo
	return
}
func (goo *TooManySignaturesError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.TooManySignaturesError = msg.(*stdpb.TooManySignaturesError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ TooManySignaturesError) GetTypeURL() (typeURL string) {
	return "/std.TooManySignaturesError"
}
func (goo NoSignaturesError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.NoSignaturesError
	{
		pbo = new(stdpb.NoSignaturesError)
	}
	msg = pbo
	return
}
func (goo NoSignaturesError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.NoSignaturesError)
	msg = pbo
	return
}
func (goo *NoSignaturesError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.NoSignaturesError = msg.(*stdpb.NoSignaturesError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ NoSignaturesError) GetTypeURL() (typeURL string) {
	return "/std.NoSignaturesError"
}
func (goo GasOverflowError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.GasOverflowError
	{
		pbo = new(stdpb.GasOverflowError)
	}
	msg = pbo
	return
}
func (goo GasOverflowError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(stdpb.GasOverflowError)
	msg = pbo
	return
}
func (goo *GasOverflowError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *stdpb.GasOverflowError = msg.(*stdpb.GasOverflowError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ GasOverflowError) GetTypeURL() (typeURL string) {
	return "/std.GasOverflowError"
}
