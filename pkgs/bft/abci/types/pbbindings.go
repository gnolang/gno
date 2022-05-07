package abci

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	abcipb "github.com/gnolang/gno/pkgs/bft/abci/types/pb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	anypb "google.golang.org/protobuf/types/known/anypb"
	merklepb "github.com/gnolang/gno/pkgs/crypto/merkle/pb"
	merkle "github.com/gnolang/gno/pkgs/crypto/merkle"
)

func (goo RequestBase) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestBase
	{
		pbo = new(abcipb.RequestBase)
	}
	msg = pbo
	return
}
func (goo RequestBase) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestBase)
	msg = pbo
	return
}
func (goo *RequestBase) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestBase = msg.(*abcipb.RequestBase)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ RequestBase) GetTypeURL() (typeURL string) {
	return "/abci.RequestBase"
}
func (goo RequestEcho) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestEcho
	{
		pbo = new(abcipb.RequestEcho)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			pbo.Message = string(goo.Message)
		}
	}
	msg = pbo
	return
}
func (goo RequestEcho) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestEcho)
	msg = pbo
	return
}
func (goo *RequestEcho) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestEcho = msg.(*abcipb.RequestEcho)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Message = string(pbo.Message)
			}
		}
	}
	return
}
func (_ RequestEcho) GetTypeURL() (typeURL string) {
	return "/abci.RequestEcho"
}
func (goo RequestFlush) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestFlush
	{
		pbo = new(abcipb.RequestFlush)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
	}
	msg = pbo
	return
}
func (goo RequestFlush) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestFlush)
	msg = pbo
	return
}
func (goo *RequestFlush) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestFlush = msg.(*abcipb.RequestFlush)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ RequestFlush) GetTypeURL() (typeURL string) {
	return "/abci.RequestFlush"
}
func (goo RequestInfo) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestInfo
	{
		pbo = new(abcipb.RequestInfo)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
	}
	msg = pbo
	return
}
func (goo RequestInfo) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestInfo)
	msg = pbo
	return
}
func (goo *RequestInfo) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestInfo = msg.(*abcipb.RequestInfo)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ RequestInfo) GetTypeURL() (typeURL string) {
	return "/abci.RequestInfo"
}
func (goo RequestSetOption) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestSetOption
	{
		pbo = new(abcipb.RequestSetOption)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			pbo.Key = string(goo.Key)
		}
		{
			pbo.Value = string(goo.Value)
		}
	}
	msg = pbo
	return
}
func (goo RequestSetOption) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestSetOption)
	msg = pbo
	return
}
func (goo *RequestSetOption) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestSetOption = msg.(*abcipb.RequestSetOption)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Key = string(pbo.Key)
			}
			{
				(*goo).Value = string(pbo.Value)
			}
		}
	}
	return
}
func (_ RequestSetOption) GetTypeURL() (typeURL string) {
	return "/abci.RequestSetOption"
}
func (goo RequestInitChain) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestInitChain
	{
		pbo = new(abcipb.RequestInitChain)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			if !amino.IsEmptyTime(goo.Time) {
				pbo.Time = timestamppb.New(goo.Time)
			}
		}
		{
			pbo.ChainID = string(goo.ChainID)
		}
		{
			if goo.ConsensusParams != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ConsensusParams.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ConsensusParams = pbom.(*abcipb.ConsensusParams)
				if pbo.ConsensusParams == nil {
					pbo.ConsensusParams = new(abcipb.ConsensusParams)
				}
			}
		}
		{
			goorl := len(goo.Validators)
			if goorl == 0 {
				pbo.Validators = nil
			} else {
				var pbos = make([]*abcipb.ValidatorUpdate, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Validators[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*abcipb.ValidatorUpdate)
						}
					}
				}
				pbo.Validators = pbos
			}
		}
		{
			if goo.AppState != nil {
				typeUrl := cdc.GetTypeURL(goo.AppState)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.AppState)
				if err != nil {
					return
				}
				pbo.AppState = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
	}
	msg = pbo
	return
}
func (goo RequestInitChain) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestInitChain)
	msg = pbo
	return
}
func (goo *RequestInitChain) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestInitChain = msg.(*abcipb.RequestInitChain)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Time = pbo.Time.AsTime()
			}
			{
				(*goo).ChainID = string(pbo.ChainID)
			}
			{
				if pbo.ConsensusParams != nil {
					(*goo).ConsensusParams = new(ConsensusParams)
					err = (*goo).ConsensusParams.FromPBMessage(cdc, pbo.ConsensusParams)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Validators != nil {
					pbol = len(pbo.Validators)
				}
				if pbol == 0 {
					(*goo).Validators = nil
				} else {
					var goors = make([]ValidatorUpdate, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Validators[i]
							{
								pboev := pboe
								if pboev != nil {
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*goo).Validators = goors
				}
			}
			{
				typeUrl := pbo.AppState.TypeUrl
				bz := pbo.AppState.Value
				goorp := &(*goo).AppState
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ RequestInitChain) GetTypeURL() (typeURL string) {
	return "/abci.RequestInitChain"
}
func (goo RequestQuery) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestQuery
	{
		pbo = new(abcipb.RequestQuery)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			goorl := len(goo.Data)
			if goorl == 0 {
				pbo.Data = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Data[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Data = pbos
			}
		}
		{
			pbo.Path = string(goo.Path)
		}
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Prove = bool(goo.Prove)
		}
	}
	msg = pbo
	return
}
func (goo RequestQuery) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestQuery)
	msg = pbo
	return
}
func (goo *RequestQuery) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestQuery = msg.(*abcipb.RequestQuery)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Data != nil {
					pbol = len(pbo.Data)
				}
				if pbol == 0 {
					(*goo).Data = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Data[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Data = goors
				}
			}
			{
				(*goo).Path = string(pbo.Path)
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Prove = bool(pbo.Prove)
			}
		}
	}
	return
}
func (_ RequestQuery) GetTypeURL() (typeURL string) {
	return "/abci.RequestQuery"
}
func (goo RequestBeginBlock) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestBeginBlock
	{
		pbo = new(abcipb.RequestBeginBlock)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			goorl := len(goo.Hash)
			if goorl == 0 {
				pbo.Hash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Hash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Hash = pbos
			}
		}
		{
			if goo.Header != nil {
				typeUrl := cdc.GetTypeURL(goo.Header)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.Header)
				if err != nil {
					return
				}
				pbo.Header = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if goo.LastCommitInfo != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.LastCommitInfo.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.LastCommitInfo = pbom.(*abcipb.LastCommitInfo)
				if pbo.LastCommitInfo == nil {
					pbo.LastCommitInfo = new(abcipb.LastCommitInfo)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo RequestBeginBlock) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestBeginBlock)
	msg = pbo
	return
}
func (goo *RequestBeginBlock) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestBeginBlock = msg.(*abcipb.RequestBeginBlock)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Hash != nil {
					pbol = len(pbo.Hash)
				}
				if pbol == 0 {
					(*goo).Hash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Hash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Hash = goors
				}
			}
			{
				typeUrl := pbo.Header.TypeUrl
				bz := pbo.Header.Value
				goorp := &(*goo).Header
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				if pbo.LastCommitInfo != nil {
					(*goo).LastCommitInfo = new(LastCommitInfo)
					err = (*goo).LastCommitInfo.FromPBMessage(cdc, pbo.LastCommitInfo)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ RequestBeginBlock) GetTypeURL() (typeURL string) {
	return "/abci.RequestBeginBlock"
}
func (goo RequestCheckTx) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestCheckTx
	{
		pbo = new(abcipb.RequestCheckTx)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			goorl := len(goo.Tx)
			if goorl == 0 {
				pbo.Tx = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Tx[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Tx = pbos
			}
		}
		{
			pbo.Type = int64(goo.Type)
		}
	}
	msg = pbo
	return
}
func (goo RequestCheckTx) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestCheckTx)
	msg = pbo
	return
}
func (goo *RequestCheckTx) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestCheckTx = msg.(*abcipb.RequestCheckTx)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Tx != nil {
					pbol = len(pbo.Tx)
				}
				if pbol == 0 {
					(*goo).Tx = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Tx[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Tx = goors
				}
			}
			{
				(*goo).Type = CheckTxType(int(pbo.Type))
			}
		}
	}
	return
}
func (_ RequestCheckTx) GetTypeURL() (typeURL string) {
	return "/abci.RequestCheckTx"
}
func (goo RequestDeliverTx) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestDeliverTx
	{
		pbo = new(abcipb.RequestDeliverTx)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			goorl := len(goo.Tx)
			if goorl == 0 {
				pbo.Tx = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Tx[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Tx = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo RequestDeliverTx) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestDeliverTx)
	msg = pbo
	return
}
func (goo *RequestDeliverTx) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestDeliverTx = msg.(*abcipb.RequestDeliverTx)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Tx != nil {
					pbol = len(pbo.Tx)
				}
				if pbol == 0 {
					(*goo).Tx = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Tx[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Tx = goors
				}
			}
		}
	}
	return
}
func (_ RequestDeliverTx) GetTypeURL() (typeURL string) {
	return "/abci.RequestDeliverTx"
}
func (goo RequestEndBlock) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestEndBlock
	{
		pbo = new(abcipb.RequestEndBlock)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
		{
			pbo.Height = int64(goo.Height)
		}
	}
	msg = pbo
	return
}
func (goo RequestEndBlock) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestEndBlock)
	msg = pbo
	return
}
func (goo *RequestEndBlock) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestEndBlock = msg.(*abcipb.RequestEndBlock)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
		}
	}
	return
}
func (_ RequestEndBlock) GetTypeURL() (typeURL string) {
	return "/abci.RequestEndBlock"
}
func (goo RequestCommit) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.RequestCommit
	{
		pbo = new(abcipb.RequestCommit)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.RequestBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.RequestBase = pbom.(*abcipb.RequestBase)
		}
	}
	msg = pbo
	return
}
func (goo RequestCommit) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.RequestCommit)
	msg = pbo
	return
}
func (goo *RequestCommit) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.RequestCommit = msg.(*abcipb.RequestCommit)
	{
		if pbo != nil {
			{
				if pbo.RequestBase != nil {
					err = (*goo).RequestBase.FromPBMessage(cdc, pbo.RequestBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ RequestCommit) GetTypeURL() (typeURL string) {
	return "/abci.RequestCommit"
}
func (goo ResponseBase) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseBase
	{
		pbo = new(abcipb.ResponseBase)
		{
			if goo.Error != nil {
				typeUrl := cdc.GetTypeURL(goo.Error)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.Error)
				if err != nil {
					return
				}
				pbo.Error = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			goorl := len(goo.Data)
			if goorl == 0 {
				pbo.Data = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Data[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Data = pbos
			}
		}
		{
			goorl := len(goo.Events)
			if goorl == 0 {
				pbo.Events = nil
			} else {
				var pbos = make([]*anypb.Any, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Events[i]
						{
							if goore != nil {
								typeUrl := cdc.GetTypeURL(goore)
								bz := []byte(nil)
								bz, err = cdc.Marshal(goore)
								if err != nil {
									return
								}
								pbos[i] = &anypb.Any{TypeUrl: typeUrl, Value: bz}
							}
						}
					}
				}
				pbo.Events = pbos
			}
		}
		{
			pbo.Log = string(goo.Log)
		}
		{
			pbo.Info = string(goo.Info)
		}
	}
	msg = pbo
	return
}
func (goo ResponseBase) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseBase)
	msg = pbo
	return
}
func (goo *ResponseBase) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseBase = msg.(*abcipb.ResponseBase)
	{
		if pbo != nil {
			{
				typeUrl := pbo.Error.TypeUrl
				bz := pbo.Error.Value
				goorp := &(*goo).Error
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				var pbol int = 0
				if pbo.Data != nil {
					pbol = len(pbo.Data)
				}
				if pbol == 0 {
					(*goo).Data = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Data[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Data = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Events != nil {
					pbol = len(pbo.Events)
				}
				if pbol == 0 {
					(*goo).Events = nil
				} else {
					var goors = make([]Event, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Events[i]
							{
								pboev := pboe
								typeUrl := pboev.TypeUrl
								bz := pboev.Value
								goorp := &goors[i]
								err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
								if err != nil {
									return
								}
							}
						}
					}
					(*goo).Events = goors
				}
			}
			{
				(*goo).Log = string(pbo.Log)
			}
			{
				(*goo).Info = string(pbo.Info)
			}
		}
	}
	return
}
func (_ ResponseBase) GetTypeURL() (typeURL string) {
	return "/abci.ResponseBase"
}
func (goo ResponseException) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseException
	{
		pbo = new(abcipb.ResponseException)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
	}
	msg = pbo
	return
}
func (goo ResponseException) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseException)
	msg = pbo
	return
}
func (goo *ResponseException) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseException = msg.(*abcipb.ResponseException)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ResponseException) GetTypeURL() (typeURL string) {
	return "/abci.ResponseException"
}
func (goo ResponseEcho) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseEcho
	{
		pbo = new(abcipb.ResponseEcho)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
		{
			pbo.Message = string(goo.Message)
		}
	}
	msg = pbo
	return
}
func (goo ResponseEcho) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseEcho)
	msg = pbo
	return
}
func (goo *ResponseEcho) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseEcho = msg.(*abcipb.ResponseEcho)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Message = string(pbo.Message)
			}
		}
	}
	return
}
func (_ ResponseEcho) GetTypeURL() (typeURL string) {
	return "/abci.ResponseEcho"
}
func (goo ResponseFlush) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseFlush
	{
		pbo = new(abcipb.ResponseFlush)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
	}
	msg = pbo
	return
}
func (goo ResponseFlush) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseFlush)
	msg = pbo
	return
}
func (goo *ResponseFlush) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseFlush = msg.(*abcipb.ResponseFlush)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ResponseFlush) GetTypeURL() (typeURL string) {
	return "/abci.ResponseFlush"
}
func (goo ResponseInfo) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseInfo
	{
		pbo = new(abcipb.ResponseInfo)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
		{
			pbo.ABCIVersion = string(goo.ABCIVersion)
		}
		{
			pbo.AppVersion = string(goo.AppVersion)
		}
		{
			pbo.LastBlockHeight = int64(goo.LastBlockHeight)
		}
		{
			goorl := len(goo.LastBlockAppHash)
			if goorl == 0 {
				pbo.LastBlockAppHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.LastBlockAppHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.LastBlockAppHash = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo ResponseInfo) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseInfo)
	msg = pbo
	return
}
func (goo *ResponseInfo) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseInfo = msg.(*abcipb.ResponseInfo)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).ABCIVersion = string(pbo.ABCIVersion)
			}
			{
				(*goo).AppVersion = string(pbo.AppVersion)
			}
			{
				(*goo).LastBlockHeight = int64(pbo.LastBlockHeight)
			}
			{
				var pbol int = 0
				if pbo.LastBlockAppHash != nil {
					pbol = len(pbo.LastBlockAppHash)
				}
				if pbol == 0 {
					(*goo).LastBlockAppHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.LastBlockAppHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).LastBlockAppHash = goors
				}
			}
		}
	}
	return
}
func (_ ResponseInfo) GetTypeURL() (typeURL string) {
	return "/abci.ResponseInfo"
}
func (goo ResponseSetOption) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseSetOption
	{
		pbo = new(abcipb.ResponseSetOption)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
	}
	msg = pbo
	return
}
func (goo ResponseSetOption) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseSetOption)
	msg = pbo
	return
}
func (goo *ResponseSetOption) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseSetOption = msg.(*abcipb.ResponseSetOption)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ResponseSetOption) GetTypeURL() (typeURL string) {
	return "/abci.ResponseSetOption"
}
func (goo ResponseInitChain) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseInitChain
	{
		pbo = new(abcipb.ResponseInitChain)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
		{
			if goo.ConsensusParams != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ConsensusParams.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ConsensusParams = pbom.(*abcipb.ConsensusParams)
				if pbo.ConsensusParams == nil {
					pbo.ConsensusParams = new(abcipb.ConsensusParams)
				}
			}
		}
		{
			goorl := len(goo.Validators)
			if goorl == 0 {
				pbo.Validators = nil
			} else {
				var pbos = make([]*abcipb.ValidatorUpdate, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Validators[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*abcipb.ValidatorUpdate)
						}
					}
				}
				pbo.Validators = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo ResponseInitChain) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseInitChain)
	msg = pbo
	return
}
func (goo *ResponseInitChain) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseInitChain = msg.(*abcipb.ResponseInitChain)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ConsensusParams != nil {
					(*goo).ConsensusParams = new(ConsensusParams)
					err = (*goo).ConsensusParams.FromPBMessage(cdc, pbo.ConsensusParams)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Validators != nil {
					pbol = len(pbo.Validators)
				}
				if pbol == 0 {
					(*goo).Validators = nil
				} else {
					var goors = make([]ValidatorUpdate, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Validators[i]
							{
								pboev := pboe
								if pboev != nil {
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*goo).Validators = goors
				}
			}
		}
	}
	return
}
func (_ ResponseInitChain) GetTypeURL() (typeURL string) {
	return "/abci.ResponseInitChain"
}
func (goo ResponseQuery) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseQuery
	{
		pbo = new(abcipb.ResponseQuery)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
		{
			goorl := len(goo.Key)
			if goorl == 0 {
				pbo.Key = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Key[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Key = pbos
			}
		}
		{
			goorl := len(goo.Value)
			if goorl == 0 {
				pbo.Value = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Value[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Value = pbos
			}
		}
		{
			if goo.Proof != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Proof.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Proof = pbom.(*merklepb.Proof)
				if pbo.Proof == nil {
					pbo.Proof = new(merklepb.Proof)
				}
			}
		}
		{
			pbo.Height = int64(goo.Height)
		}
	}
	msg = pbo
	return
}
func (goo ResponseQuery) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseQuery)
	msg = pbo
	return
}
func (goo *ResponseQuery) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseQuery = msg.(*abcipb.ResponseQuery)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Key != nil {
					pbol = len(pbo.Key)
				}
				if pbol == 0 {
					(*goo).Key = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Key[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Key = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Value != nil {
					pbol = len(pbo.Value)
				}
				if pbol == 0 {
					(*goo).Value = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Value[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Value = goors
				}
			}
			{
				if pbo.Proof != nil {
					(*goo).Proof = new(merkle.Proof)
					err = (*goo).Proof.FromPBMessage(cdc, pbo.Proof)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
		}
	}
	return
}
func (_ ResponseQuery) GetTypeURL() (typeURL string) {
	return "/abci.ResponseQuery"
}
func (goo ResponseBeginBlock) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseBeginBlock
	{
		pbo = new(abcipb.ResponseBeginBlock)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
	}
	msg = pbo
	return
}
func (goo ResponseBeginBlock) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseBeginBlock)
	msg = pbo
	return
}
func (goo *ResponseBeginBlock) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseBeginBlock = msg.(*abcipb.ResponseBeginBlock)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ResponseBeginBlock) GetTypeURL() (typeURL string) {
	return "/abci.ResponseBeginBlock"
}
func (goo ResponseCheckTx) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseCheckTx
	{
		pbo = new(abcipb.ResponseCheckTx)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
		{
			pbo.GasWanted = int64(goo.GasWanted)
		}
		{
			pbo.GasUsed = int64(goo.GasUsed)
		}
	}
	msg = pbo
	return
}
func (goo ResponseCheckTx) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseCheckTx)
	msg = pbo
	return
}
func (goo *ResponseCheckTx) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseCheckTx = msg.(*abcipb.ResponseCheckTx)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).GasWanted = int64(pbo.GasWanted)
			}
			{
				(*goo).GasUsed = int64(pbo.GasUsed)
			}
		}
	}
	return
}
func (_ ResponseCheckTx) GetTypeURL() (typeURL string) {
	return "/abci.ResponseCheckTx"
}
func (goo ResponseDeliverTx) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseDeliverTx
	{
		pbo = new(abcipb.ResponseDeliverTx)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
		{
			pbo.GasWanted = int64(goo.GasWanted)
		}
		{
			pbo.GasUsed = int64(goo.GasUsed)
		}
	}
	msg = pbo
	return
}
func (goo ResponseDeliverTx) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseDeliverTx)
	msg = pbo
	return
}
func (goo *ResponseDeliverTx) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseDeliverTx = msg.(*abcipb.ResponseDeliverTx)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).GasWanted = int64(pbo.GasWanted)
			}
			{
				(*goo).GasUsed = int64(pbo.GasUsed)
			}
		}
	}
	return
}
func (_ ResponseDeliverTx) GetTypeURL() (typeURL string) {
	return "/abci.ResponseDeliverTx"
}
func (goo ResponseEndBlock) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseEndBlock
	{
		pbo = new(abcipb.ResponseEndBlock)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
		{
			goorl := len(goo.ValidatorUpdates)
			if goorl == 0 {
				pbo.ValidatorUpdates = nil
			} else {
				var pbos = make([]*abcipb.ValidatorUpdate, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ValidatorUpdates[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*abcipb.ValidatorUpdate)
						}
					}
				}
				pbo.ValidatorUpdates = pbos
			}
		}
		{
			if goo.ConsensusParams != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ConsensusParams.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ConsensusParams = pbom.(*abcipb.ConsensusParams)
				if pbo.ConsensusParams == nil {
					pbo.ConsensusParams = new(abcipb.ConsensusParams)
				}
			}
		}
		{
			goorl := len(goo.Events)
			if goorl == 0 {
				pbo.Events = nil
			} else {
				var pbos = make([]*anypb.Any, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Events[i]
						{
							if goore != nil {
								typeUrl := cdc.GetTypeURL(goore)
								bz := []byte(nil)
								bz, err = cdc.Marshal(goore)
								if err != nil {
									return
								}
								pbos[i] = &anypb.Any{TypeUrl: typeUrl, Value: bz}
							}
						}
					}
				}
				pbo.Events = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo ResponseEndBlock) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseEndBlock)
	msg = pbo
	return
}
func (goo *ResponseEndBlock) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseEndBlock = msg.(*abcipb.ResponseEndBlock)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.ValidatorUpdates != nil {
					pbol = len(pbo.ValidatorUpdates)
				}
				if pbol == 0 {
					(*goo).ValidatorUpdates = nil
				} else {
					var goors = make([]ValidatorUpdate, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ValidatorUpdates[i]
							{
								pboev := pboe
								if pboev != nil {
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*goo).ValidatorUpdates = goors
				}
			}
			{
				if pbo.ConsensusParams != nil {
					(*goo).ConsensusParams = new(ConsensusParams)
					err = (*goo).ConsensusParams.FromPBMessage(cdc, pbo.ConsensusParams)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Events != nil {
					pbol = len(pbo.Events)
				}
				if pbol == 0 {
					(*goo).Events = nil
				} else {
					var goors = make([]Event, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Events[i]
							{
								pboev := pboe
								typeUrl := pboev.TypeUrl
								bz := pboev.Value
								goorp := &goors[i]
								err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
								if err != nil {
									return
								}
							}
						}
					}
					(*goo).Events = goors
				}
			}
		}
	}
	return
}
func (_ ResponseEndBlock) GetTypeURL() (typeURL string) {
	return "/abci.ResponseEndBlock"
}
func (goo ResponseCommit) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ResponseCommit
	{
		pbo = new(abcipb.ResponseCommit)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResponseBase.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResponseBase = pbom.(*abcipb.ResponseBase)
		}
	}
	msg = pbo
	return
}
func (goo ResponseCommit) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ResponseCommit)
	msg = pbo
	return
}
func (goo *ResponseCommit) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ResponseCommit = msg.(*abcipb.ResponseCommit)
	{
		if pbo != nil {
			{
				if pbo.ResponseBase != nil {
					err = (*goo).ResponseBase.FromPBMessage(cdc, pbo.ResponseBase)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ResponseCommit) GetTypeURL() (typeURL string) {
	return "/abci.ResponseCommit"
}
func (goo StringError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.StringError
	{
		pbo = &abcipb.StringError{Value: string(goo)}
	}
	msg = pbo
	return
}
func (goo StringError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.StringError)
	msg = pbo
	return
}
func (goo *StringError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.StringError = msg.(*abcipb.StringError)
	{
		*goo = StringError(pbo.Value)
	}
	return
}
func (_ StringError) GetTypeURL() (typeURL string) {
	return "/abci.StringError"
}
func (goo ConsensusParams) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ConsensusParams
	{
		pbo = new(abcipb.ConsensusParams)
		{
			if goo.Block != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Block.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Block = pbom.(*abcipb.BlockParams)
				if pbo.Block == nil {
					pbo.Block = new(abcipb.BlockParams)
				}
			}
		}
		{
			if goo.Validator != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Validator.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Validator = pbom.(*abcipb.ValidatorParams)
				if pbo.Validator == nil {
					pbo.Validator = new(abcipb.ValidatorParams)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo ConsensusParams) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ConsensusParams)
	msg = pbo
	return
}
func (goo *ConsensusParams) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ConsensusParams = msg.(*abcipb.ConsensusParams)
	{
		if pbo != nil {
			{
				if pbo.Block != nil {
					(*goo).Block = new(BlockParams)
					err = (*goo).Block.FromPBMessage(cdc, pbo.Block)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Validator != nil {
					(*goo).Validator = new(ValidatorParams)
					err = (*goo).Validator.FromPBMessage(cdc, pbo.Validator)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ConsensusParams) GetTypeURL() (typeURL string) {
	return "/abci.ConsensusParams"
}
func (goo BlockParams) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.BlockParams
	{
		pbo = new(abcipb.BlockParams)
		{
			pbo.MaxTxBytes = int64(goo.MaxTxBytes)
		}
		{
			pbo.MaxDataBytes = int64(goo.MaxDataBytes)
		}
		{
			pbo.MaxBlockBytes = int64(goo.MaxBlockBytes)
		}
		{
			pbo.MaxGas = int64(goo.MaxGas)
		}
		{
			pbo.TimeIotaMS = int64(goo.TimeIotaMS)
		}
	}
	msg = pbo
	return
}
func (goo BlockParams) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.BlockParams)
	msg = pbo
	return
}
func (goo *BlockParams) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.BlockParams = msg.(*abcipb.BlockParams)
	{
		if pbo != nil {
			{
				(*goo).MaxTxBytes = int64(pbo.MaxTxBytes)
			}
			{
				(*goo).MaxDataBytes = int64(pbo.MaxDataBytes)
			}
			{
				(*goo).MaxBlockBytes = int64(pbo.MaxBlockBytes)
			}
			{
				(*goo).MaxGas = int64(pbo.MaxGas)
			}
			{
				(*goo).TimeIotaMS = int64(pbo.TimeIotaMS)
			}
		}
	}
	return
}
func (_ BlockParams) GetTypeURL() (typeURL string) {
	return "/abci.BlockParams"
}
func (goo ValidatorParams) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ValidatorParams
	{
		pbo = new(abcipb.ValidatorParams)
		{
			goorl := len(goo.PubKeyTypeURLs)
			if goorl == 0 {
				pbo.PubKeyTypeURLs = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.PubKeyTypeURLs[i]
						{
							pbos[i] = string(goore)
						}
					}
				}
				pbo.PubKeyTypeURLs = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo ValidatorParams) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ValidatorParams)
	msg = pbo
	return
}
func (goo *ValidatorParams) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ValidatorParams = msg.(*abcipb.ValidatorParams)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.PubKeyTypeURLs != nil {
					pbol = len(pbo.PubKeyTypeURLs)
				}
				if pbol == 0 {
					(*goo).PubKeyTypeURLs = nil
				} else {
					var goors = make([]string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.PubKeyTypeURLs[i]
							{
								pboev := pboe
								goors[i] = string(pboev)
							}
						}
					}
					(*goo).PubKeyTypeURLs = goors
				}
			}
		}
	}
	return
}
func (_ ValidatorParams) GetTypeURL() (typeURL string) {
	return "/abci.ValidatorParams"
}
func (goo ValidatorUpdate) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.ValidatorUpdate
	{
		pbo = new(abcipb.ValidatorUpdate)
		{
			goor, err1 := goo.Address.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Address = string(goor)
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
			pbo.Power = int64(goo.Power)
		}
	}
	msg = pbo
	return
}
func (goo ValidatorUpdate) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.ValidatorUpdate)
	msg = pbo
	return
}
func (goo *ValidatorUpdate) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.ValidatorUpdate = msg.(*abcipb.ValidatorUpdate)
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
				typeUrl := pbo.PubKey.TypeUrl
				bz := pbo.PubKey.Value
				goorp := &(*goo).PubKey
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				(*goo).Power = int64(pbo.Power)
			}
		}
	}
	return
}
func (_ ValidatorUpdate) GetTypeURL() (typeURL string) {
	return "/abci.ValidatorUpdate"
}
func (goo LastCommitInfo) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.LastCommitInfo
	{
		pbo = new(abcipb.LastCommitInfo)
		{
			pbo.Round = int32(goo.Round)
		}
		{
			goorl := len(goo.Votes)
			if goorl == 0 {
				pbo.Votes = nil
			} else {
				var pbos = make([]*abcipb.VoteInfo, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Votes[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*abcipb.VoteInfo)
						}
					}
				}
				pbo.Votes = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo LastCommitInfo) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.LastCommitInfo)
	msg = pbo
	return
}
func (goo *LastCommitInfo) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.LastCommitInfo = msg.(*abcipb.LastCommitInfo)
	{
		if pbo != nil {
			{
				(*goo).Round = int32(pbo.Round)
			}
			{
				var pbol int = 0
				if pbo.Votes != nil {
					pbol = len(pbo.Votes)
				}
				if pbol == 0 {
					(*goo).Votes = nil
				} else {
					var goors = make([]VoteInfo, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Votes[i]
							{
								pboev := pboe
								if pboev != nil {
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*goo).Votes = goors
				}
			}
		}
	}
	return
}
func (_ LastCommitInfo) GetTypeURL() (typeURL string) {
	return "/abci.LastCommitInfo"
}
func (goo VoteInfo) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.VoteInfo
	{
		pbo = new(abcipb.VoteInfo)
		{
			goor, err1 := goo.Address.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Address = string(goor)
		}
		{
			pbo.Power = int64(goo.Power)
		}
		{
			pbo.SignedLastBlock = bool(goo.SignedLastBlock)
		}
	}
	msg = pbo
	return
}
func (goo VoteInfo) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.VoteInfo)
	msg = pbo
	return
}
func (goo *VoteInfo) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.VoteInfo = msg.(*abcipb.VoteInfo)
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
				(*goo).Power = int64(pbo.Power)
			}
			{
				(*goo).SignedLastBlock = bool(pbo.SignedLastBlock)
			}
		}
	}
	return
}
func (_ VoteInfo) GetTypeURL() (typeURL string) {
	return "/abci.VoteInfo"
}
func (goo EventString) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.EventString
	{
		pbo = &abcipb.EventString{Value: string(goo)}
	}
	msg = pbo
	return
}
func (goo EventString) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.EventString)
	msg = pbo
	return
}
func (goo *EventString) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.EventString = msg.(*abcipb.EventString)
	{
		*goo = EventString(pbo.Value)
	}
	return
}
func (_ EventString) GetTypeURL() (typeURL string) {
	return "/abci.EventString"
}
func (goo MockHeader) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *abcipb.MockHeader
	{
		pbo = new(abcipb.MockHeader)
		{
			pbo.Version = string(goo.Version)
		}
		{
			pbo.ChainID = string(goo.ChainID)
		}
		{
			pbo.Height = int64(goo.Height)
		}
		{
			if !amino.IsEmptyTime(goo.Time) {
				pbo.Time = timestamppb.New(goo.Time)
			}
		}
		{
			pbo.NumTxs = int64(goo.NumTxs)
		}
		{
			pbo.TotalTxs = int64(goo.TotalTxs)
		}
	}
	msg = pbo
	return
}
func (goo MockHeader) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(abcipb.MockHeader)
	msg = pbo
	return
}
func (goo *MockHeader) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *abcipb.MockHeader = msg.(*abcipb.MockHeader)
	{
		if pbo != nil {
			{
				(*goo).Version = string(pbo.Version)
			}
			{
				(*goo).ChainID = string(pbo.ChainID)
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Time = pbo.Time.AsTime()
			}
			{
				(*goo).NumTxs = int64(pbo.NumTxs)
			}
			{
				(*goo).TotalTxs = int64(pbo.TotalTxs)
			}
		}
	}
	return
}
func (_ MockHeader) GetTypeURL() (typeURL string) {
	return "/abci.MockHeader"
}
