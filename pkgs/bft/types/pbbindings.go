package types

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	typespb "github.com/gnolang/gno/pkgs/bft/types/pb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	merklepb "github.com/gnolang/gno/pkgs/crypto/merkle/pb"
	anypb "google.golang.org/protobuf/types/known/anypb"
	abcipb "github.com/gnolang/gno/pkgs/bft/abci/types/pb"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
)

func (goo Proposal) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Proposal
	{
		pbo = new(typespb.Proposal)
		{
			pbo.Type = uint32(goo.Type)
		}
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbo.POLRound = int64(goo.POLRound)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockID.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockID = pbom.(*typespb.BlockID)
		}
		{
			if !amino.IsEmptyTime(goo.Timestamp) {
				pbo.Timestamp = timestamppb.New(goo.Timestamp)
			}
		}
		{
			goorl := len(goo.Signature)
			if goorl == 0 {
				pbo.Signature = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Signature[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Signature = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo Proposal) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Proposal)
	msg = pbo
	return
}
func (goo *Proposal) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Proposal = msg.(*typespb.Proposal)
	{
		if pbo != nil {
			{
				(*goo).Type = SignedMsgType(uint8(pbo.Type))
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).POLRound = int(int(pbo.POLRound))
			}
			{
				if pbo.BlockID != nil {
					err = (*goo).BlockID.FromPBMessage(cdc, pbo.BlockID)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Timestamp = pbo.Timestamp.AsTime()
			}
			{
				var pbol int = 0
				if pbo.Signature != nil {
					pbol = len(pbo.Signature)
				}
				if pbol == 0 {
					(*goo).Signature = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Signature[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Signature = goors
				}
			}
		}
	}
	return
}
func (_ Proposal) GetTypeURL() (typeURL string) {
	return "/tm.Proposal"
}
func (goo Block) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Block
	{
		pbo = new(typespb.Block)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Header.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Header = pbom.(*typespb.Header)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Data.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Data = pbom.(*typespb.Data)
		}
		{
			if goo.LastCommit != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.LastCommit.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.LastCommit = pbom.(*typespb.Commit)
				if pbo.LastCommit == nil {
					pbo.LastCommit = new(typespb.Commit)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo Block) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Block)
	msg = pbo
	return
}
func (goo *Block) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Block = msg.(*typespb.Block)
	{
		if pbo != nil {
			{
				if pbo.Header != nil {
					err = (*goo).Header.FromPBMessage(cdc, pbo.Header)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Data != nil {
					err = (*goo).Data.FromPBMessage(cdc, pbo.Data)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.LastCommit != nil {
					(*goo).LastCommit = new(Commit)
					err = (*goo).LastCommit.FromPBMessage(cdc, pbo.LastCommit)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ Block) GetTypeURL() (typeURL string) {
	return "/tm.Block"
}
func (goo Header) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Header
	{
		pbo = new(typespb.Header)
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
		{
			pbo.AppVersion = string(goo.AppVersion)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.LastBlockID.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.LastBlockID = pbom.(*typespb.BlockID)
		}
		{
			goorl := len(goo.LastCommitHash)
			if goorl == 0 {
				pbo.LastCommitHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.LastCommitHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.LastCommitHash = pbos
			}
		}
		{
			goorl := len(goo.DataHash)
			if goorl == 0 {
				pbo.DataHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.DataHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.DataHash = pbos
			}
		}
		{
			goorl := len(goo.ValidatorsHash)
			if goorl == 0 {
				pbo.ValidatorsHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ValidatorsHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ValidatorsHash = pbos
			}
		}
		{
			goorl := len(goo.NextValidatorsHash)
			if goorl == 0 {
				pbo.NextValidatorsHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.NextValidatorsHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.NextValidatorsHash = pbos
			}
		}
		{
			goorl := len(goo.ConsensusHash)
			if goorl == 0 {
				pbo.ConsensusHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ConsensusHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ConsensusHash = pbos
			}
		}
		{
			goorl := len(goo.AppHash)
			if goorl == 0 {
				pbo.AppHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.AppHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.AppHash = pbos
			}
		}
		{
			goorl := len(goo.LastResultsHash)
			if goorl == 0 {
				pbo.LastResultsHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.LastResultsHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.LastResultsHash = pbos
			}
		}
		{
			goor, err1 := goo.ProposerAddress.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.ProposerAddress = string(goor)
		}
	}
	msg = pbo
	return
}
func (goo Header) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Header)
	msg = pbo
	return
}
func (goo *Header) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Header = msg.(*typespb.Header)
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
			{
				(*goo).AppVersion = string(pbo.AppVersion)
			}
			{
				if pbo.LastBlockID != nil {
					err = (*goo).LastBlockID.FromPBMessage(cdc, pbo.LastBlockID)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.LastCommitHash != nil {
					pbol = len(pbo.LastCommitHash)
				}
				if pbol == 0 {
					(*goo).LastCommitHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.LastCommitHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).LastCommitHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.DataHash != nil {
					pbol = len(pbo.DataHash)
				}
				if pbol == 0 {
					(*goo).DataHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.DataHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).DataHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.ValidatorsHash != nil {
					pbol = len(pbo.ValidatorsHash)
				}
				if pbol == 0 {
					(*goo).ValidatorsHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ValidatorsHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).ValidatorsHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.NextValidatorsHash != nil {
					pbol = len(pbo.NextValidatorsHash)
				}
				if pbol == 0 {
					(*goo).NextValidatorsHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.NextValidatorsHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).NextValidatorsHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.ConsensusHash != nil {
					pbol = len(pbo.ConsensusHash)
				}
				if pbol == 0 {
					(*goo).ConsensusHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ConsensusHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).ConsensusHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.AppHash != nil {
					pbol = len(pbo.AppHash)
				}
				if pbol == 0 {
					(*goo).AppHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.AppHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).AppHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.LastResultsHash != nil {
					pbol = len(pbo.LastResultsHash)
				}
				if pbol == 0 {
					(*goo).LastResultsHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.LastResultsHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).LastResultsHash = goors
				}
			}
			{
				var goor string
				goor = string(pbo.ProposerAddress)
				err = (*goo).ProposerAddress.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ Header) GetTypeURL() (typeURL string) {
	return "/tm.Header"
}
func (goo Data) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Data
	{
		pbo = new(typespb.Data)
		{
			goorl := len(goo.Txs)
			if goorl == 0 {
				pbo.Txs = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Txs[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.Txs = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo Data) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Data)
	msg = pbo
	return
}
func (goo *Data) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Data = msg.(*typespb.Data)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Txs != nil {
					pbol = len(pbo.Txs)
				}
				if pbol == 0 {
					(*goo).Txs = nil
				} else {
					var goors = make([]Tx, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Txs[i]
							{
								pboev := pboe
								var pbol1 int = 0
								if pboev != nil {
									pbol1 = len(pboev)
								}
								if pbol1 == 0 {
									goors[i] = nil
								} else {
									var goors1 = make([]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboev[i]
											{
												pboev := pboe
												goors1[i] = uint8(uint8(pboev))
											}
										}
									}
									goors[i] = goors1
								}
							}
						}
					}
					(*goo).Txs = goors
				}
			}
		}
	}
	return
}
func (_ Data) GetTypeURL() (typeURL string) {
	return "/tm.Data"
}
func (goo Commit) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Commit
	{
		pbo = new(typespb.Commit)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockID.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockID = pbom.(*typespb.BlockID)
		}
		{
			goorl := len(goo.Precommits)
			if goorl == 0 {
				pbo.Precommits = nil
			} else {
				var pbos = make([]*typespb.CommitSig, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Precommits[i]
						{
							if goore != nil {
								pbom := proto.Message(nil)
								pbom, err = goore.ToPBMessage(cdc)
								if err != nil {
									return
								}
								pbos[i] = pbom.(*typespb.CommitSig)
								if pbos[i] == nil {
									pbos[i] = new(typespb.CommitSig)
								}
							}
						}
					}
				}
				pbo.Precommits = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo Commit) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Commit)
	msg = pbo
	return
}
func (goo *Commit) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Commit = msg.(*typespb.Commit)
	{
		if pbo != nil {
			{
				if pbo.BlockID != nil {
					err = (*goo).BlockID.FromPBMessage(cdc, pbo.BlockID)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Precommits != nil {
					pbol = len(pbo.Precommits)
				}
				if pbol == 0 {
					(*goo).Precommits = nil
				} else {
					var goors = make([]*CommitSig, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Precommits[i]
							{
								pboev := pboe
								if pboev != nil {
									goors[i] = new(CommitSig)
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*goo).Precommits = goors
				}
			}
		}
	}
	return
}
func (_ Commit) GetTypeURL() (typeURL string) {
	return "/tm.Commit"
}
func (goo BlockID) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.BlockID
	{
		pbo = new(typespb.BlockID)
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
			pbom := proto.Message(nil)
			pbom, err = goo.PartsHeader.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PartsHeader = pbom.(*typespb.PartSetHeader)
		}
	}
	msg = pbo
	return
}
func (goo BlockID) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.BlockID)
	msg = pbo
	return
}
func (goo *BlockID) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.BlockID = msg.(*typespb.BlockID)
	{
		if pbo != nil {
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
				if pbo.PartsHeader != nil {
					err = (*goo).PartsHeader.FromPBMessage(cdc, pbo.PartsHeader)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ BlockID) GetTypeURL() (typeURL string) {
	return "/tm.BlockID"
}
func (goo CommitSig) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.CommitSig
	{
		pbo = new(typespb.CommitSig)
		{
			pbo.Type = uint32(goo.Type)
		}
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockID.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockID = pbom.(*typespb.BlockID)
		}
		{
			if !amino.IsEmptyTime(goo.Timestamp) {
				pbo.Timestamp = timestamppb.New(goo.Timestamp)
			}
		}
		{
			goor, err1 := goo.ValidatorAddress.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.ValidatorAddress = string(goor)
		}
		{
			pbo.ValidatorIndex = int64(goo.ValidatorIndex)
		}
		{
			goorl := len(goo.Signature)
			if goorl == 0 {
				pbo.Signature = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Signature[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Signature = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo CommitSig) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.CommitSig)
	msg = pbo
	return
}
func (goo *CommitSig) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.CommitSig = msg.(*typespb.CommitSig)
	{
		if pbo != nil {
			{
				(*goo).Type = SignedMsgType(uint8(pbo.Type))
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				if pbo.BlockID != nil {
					err = (*goo).BlockID.FromPBMessage(cdc, pbo.BlockID)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Timestamp = pbo.Timestamp.AsTime()
			}
			{
				var goor string
				goor = string(pbo.ValidatorAddress)
				err = (*goo).ValidatorAddress.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				(*goo).ValidatorIndex = int(int(pbo.ValidatorIndex))
			}
			{
				var pbol int = 0
				if pbo.Signature != nil {
					pbol = len(pbo.Signature)
				}
				if pbol == 0 {
					(*goo).Signature = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Signature[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Signature = goors
				}
			}
		}
	}
	return
}
func (_ CommitSig) GetTypeURL() (typeURL string) {
	return "/tm.CommitSig"
}
func (goo Vote) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Vote
	{
		pbo = new(typespb.Vote)
		{
			pbo.Type = uint32(goo.Type)
		}
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockID.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockID = pbom.(*typespb.BlockID)
		}
		{
			if !amino.IsEmptyTime(goo.Timestamp) {
				pbo.Timestamp = timestamppb.New(goo.Timestamp)
			}
		}
		{
			goor, err1 := goo.ValidatorAddress.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.ValidatorAddress = string(goor)
		}
		{
			pbo.ValidatorIndex = int64(goo.ValidatorIndex)
		}
		{
			goorl := len(goo.Signature)
			if goorl == 0 {
				pbo.Signature = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Signature[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Signature = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo Vote) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Vote)
	msg = pbo
	return
}
func (goo *Vote) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Vote = msg.(*typespb.Vote)
	{
		if pbo != nil {
			{
				(*goo).Type = SignedMsgType(uint8(pbo.Type))
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				if pbo.BlockID != nil {
					err = (*goo).BlockID.FromPBMessage(cdc, pbo.BlockID)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Timestamp = pbo.Timestamp.AsTime()
			}
			{
				var goor string
				goor = string(pbo.ValidatorAddress)
				err = (*goo).ValidatorAddress.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				(*goo).ValidatorIndex = int(int(pbo.ValidatorIndex))
			}
			{
				var pbol int = 0
				if pbo.Signature != nil {
					pbol = len(pbo.Signature)
				}
				if pbol == 0 {
					(*goo).Signature = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Signature[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Signature = goors
				}
			}
		}
	}
	return
}
func (_ Vote) GetTypeURL() (typeURL string) {
	return "/tm.Vote"
}
func (goo Part) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Part
	{
		pbo = new(typespb.Part)
		{
			pbo.Index = int64(goo.Index)
		}
		{
			goorl := len(goo.Bytes)
			if goorl == 0 {
				pbo.Bytes = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Bytes[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Bytes = pbos
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Proof.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Proof = pbom.(*merklepb.SimpleProof)
		}
	}
	msg = pbo
	return
}
func (goo Part) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Part)
	msg = pbo
	return
}
func (goo *Part) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Part = msg.(*typespb.Part)
	{
		if pbo != nil {
			{
				(*goo).Index = int(int(pbo.Index))
			}
			{
				var pbol int = 0
				if pbo.Bytes != nil {
					pbol = len(pbo.Bytes)
				}
				if pbol == 0 {
					(*goo).Bytes = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Bytes[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Bytes = goors
				}
			}
			{
				if pbo.Proof != nil {
					err = (*goo).Proof.FromPBMessage(cdc, pbo.Proof)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ Part) GetTypeURL() (typeURL string) {
	return "/tm.Part"
}
func (goo PartSet) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.PartSet
	{
		pbo = new(typespb.PartSet)
	}
	msg = pbo
	return
}
func (goo PartSet) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.PartSet)
	msg = pbo
	return
}
func (goo *PartSet) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.PartSet = msg.(*typespb.PartSet)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ PartSet) GetTypeURL() (typeURL string) {
	return "/tm.PartSet"
}
func (goo PartSetHeader) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.PartSetHeader
	{
		pbo = new(typespb.PartSetHeader)
		{
			pbo.Total = int64(goo.Total)
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
	}
	msg = pbo
	return
}
func (goo PartSetHeader) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.PartSetHeader)
	msg = pbo
	return
}
func (goo *PartSetHeader) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.PartSetHeader = msg.(*typespb.PartSetHeader)
	{
		if pbo != nil {
			{
				(*goo).Total = int(int(pbo.Total))
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
		}
	}
	return
}
func (_ PartSetHeader) GetTypeURL() (typeURL string) {
	return "/tm.PartSetHeader"
}
func (goo Validator) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.Validator
	{
		pbo = new(typespb.Validator)
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
			pbo.VotingPower = int64(goo.VotingPower)
		}
		{
			pbo.ProposerPriority = int64(goo.ProposerPriority)
		}
	}
	msg = pbo
	return
}
func (goo Validator) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.Validator)
	msg = pbo
	return
}
func (goo *Validator) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.Validator = msg.(*typespb.Validator)
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
				(*goo).VotingPower = int64(pbo.VotingPower)
			}
			{
				(*goo).ProposerPriority = int64(pbo.ProposerPriority)
			}
		}
	}
	return
}
func (_ Validator) GetTypeURL() (typeURL string) {
	return "/tm.Validator"
}
func (goo ValidatorSet) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.ValidatorSet
	{
		pbo = new(typespb.ValidatorSet)
		{
			goorl := len(goo.Validators)
			if goorl == 0 {
				pbo.Validators = nil
			} else {
				var pbos = make([]*typespb.Validator, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Validators[i]
						{
							if goore != nil {
								pbom := proto.Message(nil)
								pbom, err = goore.ToPBMessage(cdc)
								if err != nil {
									return
								}
								pbos[i] = pbom.(*typespb.Validator)
								if pbos[i] == nil {
									pbos[i] = new(typespb.Validator)
								}
							}
						}
					}
				}
				pbo.Validators = pbos
			}
		}
		{
			if goo.Proposer != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Proposer.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Proposer = pbom.(*typespb.Validator)
				if pbo.Proposer == nil {
					pbo.Proposer = new(typespb.Validator)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo ValidatorSet) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.ValidatorSet)
	msg = pbo
	return
}
func (goo *ValidatorSet) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.ValidatorSet = msg.(*typespb.ValidatorSet)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Validators != nil {
					pbol = len(pbo.Validators)
				}
				if pbol == 0 {
					(*goo).Validators = nil
				} else {
					var goors = make([]*Validator, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Validators[i]
							{
								pboev := pboe
								if pboev != nil {
									goors[i] = new(Validator)
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
				if pbo.Proposer != nil {
					(*goo).Proposer = new(Validator)
					err = (*goo).Proposer.FromPBMessage(cdc, pbo.Proposer)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ValidatorSet) GetTypeURL() (typeURL string) {
	return "/tm.ValidatorSet"
}
func (goo EventNewBlock) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventNewBlock
	{
		pbo = new(typespb.EventNewBlock)
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
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResultBeginBlock.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResultBeginBlock = pbom.(*abcipb.ResponseBeginBlock)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResultEndBlock.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResultEndBlock = pbom.(*abcipb.ResponseEndBlock)
		}
	}
	msg = pbo
	return
}
func (goo EventNewBlock) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventNewBlock)
	msg = pbo
	return
}
func (goo *EventNewBlock) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventNewBlock = msg.(*typespb.EventNewBlock)
	{
		if pbo != nil {
			{
				if pbo.Block != nil {
					(*goo).Block = new(Block)
					err = (*goo).Block.FromPBMessage(cdc, pbo.Block)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ResultBeginBlock != nil {
					err = (*goo).ResultBeginBlock.FromPBMessage(cdc, pbo.ResultBeginBlock)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ResultEndBlock != nil {
					err = (*goo).ResultEndBlock.FromPBMessage(cdc, pbo.ResultEndBlock)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EventNewBlock) GetTypeURL() (typeURL string) {
	return "/tm.EventNewBlock"
}
func (goo EventNewBlockHeader) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventNewBlockHeader
	{
		pbo = new(typespb.EventNewBlockHeader)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Header.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Header = pbom.(*typespb.Header)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResultBeginBlock.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResultBeginBlock = pbom.(*abcipb.ResponseBeginBlock)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ResultEndBlock.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ResultEndBlock = pbom.(*abcipb.ResponseEndBlock)
		}
	}
	msg = pbo
	return
}
func (goo EventNewBlockHeader) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventNewBlockHeader)
	msg = pbo
	return
}
func (goo *EventNewBlockHeader) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventNewBlockHeader = msg.(*typespb.EventNewBlockHeader)
	{
		if pbo != nil {
			{
				if pbo.Header != nil {
					err = (*goo).Header.FromPBMessage(cdc, pbo.Header)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ResultBeginBlock != nil {
					err = (*goo).ResultBeginBlock.FromPBMessage(cdc, pbo.ResultBeginBlock)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ResultEndBlock != nil {
					err = (*goo).ResultEndBlock.FromPBMessage(cdc, pbo.ResultEndBlock)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EventNewBlockHeader) GetTypeURL() (typeURL string) {
	return "/tm.EventNewBlockHeader"
}
func (goo EventTx) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventTx
	{
		pbo = new(typespb.EventTx)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Result.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Result = pbom.(*typespb.TxResult)
		}
	}
	msg = pbo
	return
}
func (goo EventTx) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventTx)
	msg = pbo
	return
}
func (goo *EventTx) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventTx = msg.(*typespb.EventTx)
	{
		if pbo != nil {
			{
				if pbo.Result != nil {
					err = (*goo).Result.FromPBMessage(cdc, pbo.Result)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EventTx) GetTypeURL() (typeURL string) {
	return "/tm.EventTx"
}
func (goo EventVote) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventVote
	{
		pbo = new(typespb.EventVote)
		{
			if goo.Vote != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Vote.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Vote = pbom.(*typespb.Vote)
				if pbo.Vote == nil {
					pbo.Vote = new(typespb.Vote)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo EventVote) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventVote)
	msg = pbo
	return
}
func (goo *EventVote) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventVote = msg.(*typespb.EventVote)
	{
		if pbo != nil {
			{
				if pbo.Vote != nil {
					(*goo).Vote = new(Vote)
					err = (*goo).Vote.FromPBMessage(cdc, pbo.Vote)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EventVote) GetTypeURL() (typeURL string) {
	return "/tm.EventVote"
}
func (goo EventString) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventString
	{
		pbo = &typespb.EventString{Value: string(goo)}
	}
	msg = pbo
	return
}
func (goo EventString) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventString)
	msg = pbo
	return
}
func (goo *EventString) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventString = msg.(*typespb.EventString)
	{
		*goo = EventString(pbo.Value)
	}
	return
}
func (_ EventString) GetTypeURL() (typeURL string) {
	return "/tm.EventString"
}
func (goo EventValidatorSetUpdates) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventValidatorSetUpdates
	{
		pbo = new(typespb.EventValidatorSetUpdates)
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
	}
	msg = pbo
	return
}
func (goo EventValidatorSetUpdates) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventValidatorSetUpdates)
	msg = pbo
	return
}
func (goo *EventValidatorSetUpdates) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventValidatorSetUpdates = msg.(*typespb.EventValidatorSetUpdates)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.ValidatorUpdates != nil {
					pbol = len(pbo.ValidatorUpdates)
				}
				if pbol == 0 {
					(*goo).ValidatorUpdates = nil
				} else {
					var goors = make([]abci.ValidatorUpdate, pbol)
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
		}
	}
	return
}
func (_ EventValidatorSetUpdates) GetTypeURL() (typeURL string) {
	return "/tm.EventValidatorSetUpdates"
}
func (goo DuplicateVoteEvidence) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.DuplicateVoteEvidence
	{
		pbo = new(typespb.DuplicateVoteEvidence)
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
			if goo.VoteA != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.VoteA.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.VoteA = pbom.(*typespb.Vote)
				if pbo.VoteA == nil {
					pbo.VoteA = new(typespb.Vote)
				}
			}
		}
		{
			if goo.VoteB != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.VoteB.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.VoteB = pbom.(*typespb.Vote)
				if pbo.VoteB == nil {
					pbo.VoteB = new(typespb.Vote)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo DuplicateVoteEvidence) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.DuplicateVoteEvidence)
	msg = pbo
	return
}
func (goo *DuplicateVoteEvidence) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.DuplicateVoteEvidence = msg.(*typespb.DuplicateVoteEvidence)
	{
		if pbo != nil {
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
				if pbo.VoteA != nil {
					(*goo).VoteA = new(Vote)
					err = (*goo).VoteA.FromPBMessage(cdc, pbo.VoteA)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.VoteB != nil {
					(*goo).VoteB = new(Vote)
					err = (*goo).VoteB.FromPBMessage(cdc, pbo.VoteB)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ DuplicateVoteEvidence) GetTypeURL() (typeURL string) {
	return "/tm.DuplicateVoteEvidence"
}
func (goo MockGoodEvidence) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.MockGoodEvidence
	{
		pbo = new(typespb.MockGoodEvidence)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			goor, err1 := goo.Address.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Address = string(goor)
		}
	}
	msg = pbo
	return
}
func (goo MockGoodEvidence) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.MockGoodEvidence)
	msg = pbo
	return
}
func (goo *MockGoodEvidence) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.MockGoodEvidence = msg.(*typespb.MockGoodEvidence)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				var goor string
				goor = string(pbo.Address)
				err = (*goo).Address.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ MockGoodEvidence) GetTypeURL() (typeURL string) {
	return "/tm.MockGoodEvidence"
}
func (goo MockRandomGoodEvidence) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.MockRandomGoodEvidence
	{
		pbo = new(typespb.MockRandomGoodEvidence)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.MockGoodEvidence.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.MockGoodEvidence = pbom.(*typespb.MockGoodEvidence)
		}
	}
	msg = pbo
	return
}
func (goo MockRandomGoodEvidence) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.MockRandomGoodEvidence)
	msg = pbo
	return
}
func (goo *MockRandomGoodEvidence) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.MockRandomGoodEvidence = msg.(*typespb.MockRandomGoodEvidence)
	{
		if pbo != nil {
			{
				if pbo.MockGoodEvidence != nil {
					err = (*goo).MockGoodEvidence.FromPBMessage(cdc, pbo.MockGoodEvidence)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ MockRandomGoodEvidence) GetTypeURL() (typeURL string) {
	return "/tm.MockRandomGoodEvidence"
}
func (goo MockBadEvidence) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.MockBadEvidence
	{
		pbo = new(typespb.MockBadEvidence)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.MockGoodEvidence.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.MockGoodEvidence = pbom.(*typespb.MockGoodEvidence)
		}
	}
	msg = pbo
	return
}
func (goo MockBadEvidence) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.MockBadEvidence)
	msg = pbo
	return
}
func (goo *MockBadEvidence) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.MockBadEvidence = msg.(*typespb.MockBadEvidence)
	{
		if pbo != nil {
			{
				if pbo.MockGoodEvidence != nil {
					err = (*goo).MockGoodEvidence.FromPBMessage(cdc, pbo.MockGoodEvidence)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ MockBadEvidence) GetTypeURL() (typeURL string) {
	return "/tm.MockBadEvidence"
}
func (goo TxResult) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.TxResult
	{
		pbo = new(typespb.TxResult)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Index = uint32(goo.Index)
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
			pbom := proto.Message(nil)
			pbom, err = goo.Response.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Response = pbom.(*abcipb.ResponseDeliverTx)
		}
	}
	msg = pbo
	return
}
func (goo TxResult) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.TxResult)
	msg = pbo
	return
}
func (goo *TxResult) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.TxResult = msg.(*typespb.TxResult)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Index = uint32(pbo.Index)
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
				if pbo.Response != nil {
					err = (*goo).Response.FromPBMessage(cdc, pbo.Response)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ TxResult) GetTypeURL() (typeURL string) {
	return "/tm.TxResult"
}
func (goo MockAppState) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.MockAppState
	{
		pbo = new(typespb.MockAppState)
		{
			pbo.AccountOwner = string(goo.AccountOwner)
		}
	}
	msg = pbo
	return
}
func (goo MockAppState) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.MockAppState)
	msg = pbo
	return
}
func (goo *MockAppState) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.MockAppState = msg.(*typespb.MockAppState)
	{
		if pbo != nil {
			{
				(*goo).AccountOwner = string(pbo.AccountOwner)
			}
		}
	}
	return
}
func (_ MockAppState) GetTypeURL() (typeURL string) {
	return "/tm.MockAppState"
}
func (goo VoteSet) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.VoteSet
	{
		pbo = new(typespb.VoteSet)
	}
	msg = pbo
	return
}
func (goo VoteSet) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.VoteSet)
	msg = pbo
	return
}
func (goo *VoteSet) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.VoteSet = msg.(*typespb.VoteSet)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ VoteSet) GetTypeURL() (typeURL string) {
	return "/tm.VoteSet"
}
