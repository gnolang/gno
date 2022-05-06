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
		if IsBaseAccountReprEmpty(goo) {
			var pbov *stdpb.BaseAccount
			msg = pbov
			return
		}
		pbo = new(stdpb.BaseAccount)
		{
			pbo.Address = string(goo.Address)
		}
		{
			pbo.Coins = string(goo.Coins)
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
func IsBaseAccountReprEmpty(goor BaseAccount) (empty bool) {
	{
		empty = true
		{
			if goor.Address != "" {
				return false
			}
		}
		{
			if goor.Coins != "" {
				return false
			}
		}
		{
			if goor.PubKey != nil {
				return false
			}
		}
		{
			if goor.AccountNumber != 0 {
				return false
			}
		}
		{
			if goor.Sequence != 0 {
				return false
			}
		}
	}
	return
}
func (goo MemFile) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.MemFile
	{
		if IsMemFileReprEmpty(goo) {
			var pbov *stdpb.MemFile
			msg = pbov
			return
		}
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
func IsMemFileReprEmpty(goor MemFile) (empty bool) {
	{
		empty = true
		{
			if goor.Name != "" {
				return false
			}
		}
		{
			if goor.Body != "" {
				return false
			}
		}
	}
	return
}
func (goo MemPackage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.MemPackage
	{
		if IsMemPackageReprEmpty(goo) {
			var pbov *stdpb.MemPackage
			msg = pbov
			return
		}
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
func IsMemPackageReprEmpty(goor MemPackage) (empty bool) {
	{
		empty = true
		{
			if goor.Name != "" {
				return false
			}
		}
		{
			if goor.Path != "" {
				return false
			}
		}
		{
			if len(goor.Files) != 0 {
				return false
			}
		}
	}
	return
}
func (goo InternalError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InternalError
	{
		if IsInternalErrorReprEmpty(goo) {
			var pbov *stdpb.InternalError
			msg = pbov
			return
		}
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
func IsInternalErrorReprEmpty(goor InternalError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo TxDecodeError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.TxDecodeError
	{
		if IsTxDecodeErrorReprEmpty(goo) {
			var pbov *stdpb.TxDecodeError
			msg = pbov
			return
		}
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
func IsTxDecodeErrorReprEmpty(goor TxDecodeError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo InvalidSequenceError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InvalidSequenceError
	{
		if IsInvalidSequenceErrorReprEmpty(goo) {
			var pbov *stdpb.InvalidSequenceError
			msg = pbov
			return
		}
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
func IsInvalidSequenceErrorReprEmpty(goor InvalidSequenceError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo UnauthorizedError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.UnauthorizedError
	{
		if IsUnauthorizedErrorReprEmpty(goo) {
			var pbov *stdpb.UnauthorizedError
			msg = pbov
			return
		}
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
func IsUnauthorizedErrorReprEmpty(goor UnauthorizedError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo InsufficientFundsError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InsufficientFundsError
	{
		if IsInsufficientFundsErrorReprEmpty(goo) {
			var pbov *stdpb.InsufficientFundsError
			msg = pbov
			return
		}
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
func IsInsufficientFundsErrorReprEmpty(goor InsufficientFundsError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo UnknownRequestError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.UnknownRequestError
	{
		if IsUnknownRequestErrorReprEmpty(goo) {
			var pbov *stdpb.UnknownRequestError
			msg = pbov
			return
		}
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
func IsUnknownRequestErrorReprEmpty(goor UnknownRequestError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo InvalidAddressError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InvalidAddressError
	{
		if IsInvalidAddressErrorReprEmpty(goo) {
			var pbov *stdpb.InvalidAddressError
			msg = pbov
			return
		}
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
func IsInvalidAddressErrorReprEmpty(goor InvalidAddressError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo UnknownAddressError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.UnknownAddressError
	{
		if IsUnknownAddressErrorReprEmpty(goo) {
			var pbov *stdpb.UnknownAddressError
			msg = pbov
			return
		}
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
func IsUnknownAddressErrorReprEmpty(goor UnknownAddressError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo InvalidPubKeyError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InvalidPubKeyError
	{
		if IsInvalidPubKeyErrorReprEmpty(goo) {
			var pbov *stdpb.InvalidPubKeyError
			msg = pbov
			return
		}
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
func IsInvalidPubKeyErrorReprEmpty(goor InvalidPubKeyError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo InsufficientCoinsError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InsufficientCoinsError
	{
		if IsInsufficientCoinsErrorReprEmpty(goo) {
			var pbov *stdpb.InsufficientCoinsError
			msg = pbov
			return
		}
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
func IsInsufficientCoinsErrorReprEmpty(goor InsufficientCoinsError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo OutOfGasError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.OutOfGasError
	{
		if IsOutOfGasErrorReprEmpty(goo) {
			var pbov *stdpb.OutOfGasError
			msg = pbov
			return
		}
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
func IsOutOfGasErrorReprEmpty(goor OutOfGasError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo MemoTooLargeError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.MemoTooLargeError
	{
		if IsMemoTooLargeErrorReprEmpty(goo) {
			var pbov *stdpb.MemoTooLargeError
			msg = pbov
			return
		}
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
func IsMemoTooLargeErrorReprEmpty(goor MemoTooLargeError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo InsufficientFeeError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.InsufficientFeeError
	{
		if IsInsufficientFeeErrorReprEmpty(goo) {
			var pbov *stdpb.InsufficientFeeError
			msg = pbov
			return
		}
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
func IsInsufficientFeeErrorReprEmpty(goor InsufficientFeeError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo TooManySignaturesError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.TooManySignaturesError
	{
		if IsTooManySignaturesErrorReprEmpty(goo) {
			var pbov *stdpb.TooManySignaturesError
			msg = pbov
			return
		}
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
func IsTooManySignaturesErrorReprEmpty(goor TooManySignaturesError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo NoSignaturesError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.NoSignaturesError
	{
		if IsNoSignaturesErrorReprEmpty(goo) {
			var pbov *stdpb.NoSignaturesError
			msg = pbov
			return
		}
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
func IsNoSignaturesErrorReprEmpty(goor NoSignaturesError) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo GasOverflowError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *stdpb.GasOverflowError
	{
		if IsGasOverflowErrorReprEmpty(goo) {
			var pbov *stdpb.GasOverflowError
			msg = pbov
			return
		}
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
func IsGasOverflowErrorReprEmpty(goor GasOverflowError) (empty bool) {
	{
		empty = true
	}
	return
}
