package types

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	typespb "github.com/gnolang/gno/pkgs/bft/consensus/types/pb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	typespb1 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb2 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb3 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb4 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb5 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb6 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb7 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb8 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb9 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb10 "github.com/gnolang/gno/pkgs/bft/types/pb"
	types "github.com/gnolang/gno/pkgs/bft/types"
	typespb11 "github.com/gnolang/gno/pkgs/bft/types/pb"
	bitarraypb "github.com/gnolang/gno/pkgs/bitarray/pb"
	bitarray "github.com/gnolang/gno/pkgs/bitarray"
	typespb12 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb13 "github.com/gnolang/gno/pkgs/bft/types/pb"
	typespb14 "github.com/gnolang/gno/pkgs/bft/types/pb"
)

func (goo RoundState) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.RoundState
	{
		if IsRoundStateReprEmpty(goo) {
			var pbov *typespb.RoundState
			msg = pbov
			return
		}
		pbo = new(typespb.RoundState)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbo.Step = uint32(goo.Step)
		}
		{
			if !amino.IsEmptyTime(goo.StartTime) {
				pbo.StartTime = timestamppb.New(goo.StartTime)
			}
		}
		{
			if !amino.IsEmptyTime(goo.CommitTime) {
				pbo.CommitTime = timestamppb.New(goo.CommitTime)
			}
		}
		{
			if goo.Validators != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Validators.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Validators = pbom.(*typespb1.ValidatorSet)
				if pbo.Validators == nil {
					pbo.Validators = new(typespb1.ValidatorSet)
				}
			}
		}
		{
			if goo.Proposal != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Proposal.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Proposal = pbom.(*typespb2.Proposal)
				if pbo.Proposal == nil {
					pbo.Proposal = new(typespb2.Proposal)
				}
			}
		}
		{
			if goo.ProposalBlock != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ProposalBlock.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ProposalBlock = pbom.(*typespb3.Block)
				if pbo.ProposalBlock == nil {
					pbo.ProposalBlock = new(typespb3.Block)
				}
			}
		}
		{
			if goo.ProposalBlockParts != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ProposalBlockParts.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ProposalBlockParts = pbom.(*typespb4.PartSet)
				if pbo.ProposalBlockParts == nil {
					pbo.ProposalBlockParts = new(typespb4.PartSet)
				}
			}
		}
		{
			pbo.LockedRound = int64(goo.LockedRound)
		}
		{
			if goo.LockedBlock != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.LockedBlock.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.LockedBlock = pbom.(*typespb5.Block)
				if pbo.LockedBlock == nil {
					pbo.LockedBlock = new(typespb5.Block)
				}
			}
		}
		{
			if goo.LockedBlockParts != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.LockedBlockParts.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.LockedBlockParts = pbom.(*typespb6.PartSet)
				if pbo.LockedBlockParts == nil {
					pbo.LockedBlockParts = new(typespb6.PartSet)
				}
			}
		}
		{
			pbo.ValidRound = int64(goo.ValidRound)
		}
		{
			if goo.ValidBlock != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ValidBlock.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ValidBlock = pbom.(*typespb7.Block)
				if pbo.ValidBlock == nil {
					pbo.ValidBlock = new(typespb7.Block)
				}
			}
		}
		{
			if goo.ValidBlockParts != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ValidBlockParts.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ValidBlockParts = pbom.(*typespb8.PartSet)
				if pbo.ValidBlockParts == nil {
					pbo.ValidBlockParts = new(typespb8.PartSet)
				}
			}
		}
		{
			if goo.Votes != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Votes.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Votes = pbom.(*typespb.HeightVoteSet)
				if pbo.Votes == nil {
					pbo.Votes = new(typespb.HeightVoteSet)
				}
			}
		}
		{
			pbo.CommitRound = int64(goo.CommitRound)
		}
		{
			if goo.LastCommit != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.LastCommit.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.LastCommit = pbom.(*typespb9.VoteSet)
				if pbo.LastCommit == nil {
					pbo.LastCommit = new(typespb9.VoteSet)
				}
			}
		}
		{
			if goo.LastValidators != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.LastValidators.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.LastValidators = pbom.(*typespb10.ValidatorSet)
				if pbo.LastValidators == nil {
					pbo.LastValidators = new(typespb10.ValidatorSet)
				}
			}
		}
		{
			pbo.TriggeredTimeoutPrecommit = bool(goo.TriggeredTimeoutPrecommit)
		}
	}
	msg = pbo
	return
}
func (goo RoundState) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.RoundState)
	msg = pbo
	return
}
func (goo *RoundState) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.RoundState = msg.(*typespb.RoundState)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Step = RoundStepType(uint8(pbo.Step))
			}
			{
				(*goo).StartTime = pbo.StartTime.AsTime()
			}
			{
				(*goo).CommitTime = pbo.CommitTime.AsTime()
			}
			{
				if pbo.Validators != nil {
					(*goo).Validators = new(types.ValidatorSet)
					err = (*goo).Validators.FromPBMessage(cdc, pbo.Validators)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Proposal != nil {
					(*goo).Proposal = new(types.Proposal)
					err = (*goo).Proposal.FromPBMessage(cdc, pbo.Proposal)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ProposalBlock != nil {
					(*goo).ProposalBlock = new(types.Block)
					err = (*goo).ProposalBlock.FromPBMessage(cdc, pbo.ProposalBlock)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ProposalBlockParts != nil {
					(*goo).ProposalBlockParts = new(types.PartSet)
					err = (*goo).ProposalBlockParts.FromPBMessage(cdc, pbo.ProposalBlockParts)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).LockedRound = int(int(pbo.LockedRound))
			}
			{
				if pbo.LockedBlock != nil {
					(*goo).LockedBlock = new(types.Block)
					err = (*goo).LockedBlock.FromPBMessage(cdc, pbo.LockedBlock)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.LockedBlockParts != nil {
					(*goo).LockedBlockParts = new(types.PartSet)
					err = (*goo).LockedBlockParts.FromPBMessage(cdc, pbo.LockedBlockParts)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).ValidRound = int(int(pbo.ValidRound))
			}
			{
				if pbo.ValidBlock != nil {
					(*goo).ValidBlock = new(types.Block)
					err = (*goo).ValidBlock.FromPBMessage(cdc, pbo.ValidBlock)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ValidBlockParts != nil {
					(*goo).ValidBlockParts = new(types.PartSet)
					err = (*goo).ValidBlockParts.FromPBMessage(cdc, pbo.ValidBlockParts)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Votes != nil {
					(*goo).Votes = new(HeightVoteSet)
					err = (*goo).Votes.FromPBMessage(cdc, pbo.Votes)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).CommitRound = int(int(pbo.CommitRound))
			}
			{
				if pbo.LastCommit != nil {
					(*goo).LastCommit = new(types.VoteSet)
					err = (*goo).LastCommit.FromPBMessage(cdc, pbo.LastCommit)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.LastValidators != nil {
					(*goo).LastValidators = new(types.ValidatorSet)
					err = (*goo).LastValidators.FromPBMessage(cdc, pbo.LastValidators)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).TriggeredTimeoutPrecommit = bool(pbo.TriggeredTimeoutPrecommit)
			}
		}
	}
	return
}
func (_ RoundState) GetTypeURL() (typeURL string) {
	return "/tm.RoundState"
}
func IsRoundStateReprEmpty(goor RoundState) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
				return false
			}
		}
		{
			if goor.Round != 0 {
				return false
			}
		}
		{
			if goor.Step != 0 {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.StartTime) {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.CommitTime) {
				return false
			}
		}
		{
			if goor.Validators != nil {
				return false
			}
		}
		{
			if goor.Proposal != nil {
				return false
			}
		}
		{
			if goor.ProposalBlock != nil {
				return false
			}
		}
		{
			if goor.ProposalBlockParts != nil {
				return false
			}
		}
		{
			if goor.LockedRound != 0 {
				return false
			}
		}
		{
			if goor.LockedBlock != nil {
				return false
			}
		}
		{
			if goor.LockedBlockParts != nil {
				return false
			}
		}
		{
			if goor.ValidRound != 0 {
				return false
			}
		}
		{
			if goor.ValidBlock != nil {
				return false
			}
		}
		{
			if goor.ValidBlockParts != nil {
				return false
			}
		}
		{
			if goor.Votes != nil {
				return false
			}
		}
		{
			if goor.CommitRound != 0 {
				return false
			}
		}
		{
			if goor.LastCommit != nil {
				return false
			}
		}
		{
			if goor.LastValidators != nil {
				return false
			}
		}
		{
			if goor.TriggeredTimeoutPrecommit != false {
				return false
			}
		}
	}
	return
}
func (goo HRS) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.HRS
	{
		if IsHRSReprEmpty(goo) {
			var pbov *typespb.HRS
			msg = pbov
			return
		}
		pbo = new(typespb.HRS)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbo.Step = uint32(goo.Step)
		}
	}
	msg = pbo
	return
}
func (goo HRS) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.HRS)
	msg = pbo
	return
}
func (goo *HRS) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.HRS = msg.(*typespb.HRS)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Step = RoundStepType(uint8(pbo.Step))
			}
		}
	}
	return
}
func (_ HRS) GetTypeURL() (typeURL string) {
	return "/tm.HRS"
}
func IsHRSReprEmpty(goor HRS) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
				return false
			}
		}
		{
			if goor.Round != 0 {
				return false
			}
		}
		{
			if goor.Step != 0 {
				return false
			}
		}
	}
	return
}
func (goo RoundStateSimple) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.RoundStateSimple
	{
		if IsRoundStateSimpleReprEmpty(goo) {
			var pbov *typespb.RoundStateSimple
			msg = pbov
			return
		}
		pbo = new(typespb.RoundStateSimple)
		{
			pbo.HeightRoundStep = string(goo.HeightRoundStep)
		}
		{
			if !amino.IsEmptyTime(goo.StartTime) {
				pbo.StartTime = timestamppb.New(goo.StartTime)
			}
		}
		{
			goorl := len(goo.ProposalBlockHash)
			if goorl == 0 {
				pbo.ProposalBlockHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ProposalBlockHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ProposalBlockHash = pbos
			}
		}
		{
			goorl := len(goo.LockedBlockHash)
			if goorl == 0 {
				pbo.LockedBlockHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.LockedBlockHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.LockedBlockHash = pbos
			}
		}
		{
			goorl := len(goo.ValidBlockHash)
			if goorl == 0 {
				pbo.ValidBlockHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ValidBlockHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ValidBlockHash = pbos
			}
		}
		{
			if goo.Votes != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Votes.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Votes = pbom.(*typespb.HeightVoteSet)
				if pbo.Votes == nil {
					pbo.Votes = new(typespb.HeightVoteSet)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo RoundStateSimple) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.RoundStateSimple)
	msg = pbo
	return
}
func (goo *RoundStateSimple) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.RoundStateSimple = msg.(*typespb.RoundStateSimple)
	{
		if pbo != nil {
			{
				(*goo).HeightRoundStep = string(pbo.HeightRoundStep)
			}
			{
				(*goo).StartTime = pbo.StartTime.AsTime()
			}
			{
				var pbol int = 0
				if pbo.ProposalBlockHash != nil {
					pbol = len(pbo.ProposalBlockHash)
				}
				if pbol == 0 {
					(*goo).ProposalBlockHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ProposalBlockHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).ProposalBlockHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.LockedBlockHash != nil {
					pbol = len(pbo.LockedBlockHash)
				}
				if pbol == 0 {
					(*goo).LockedBlockHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.LockedBlockHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).LockedBlockHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.ValidBlockHash != nil {
					pbol = len(pbo.ValidBlockHash)
				}
				if pbol == 0 {
					(*goo).ValidBlockHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ValidBlockHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).ValidBlockHash = goors
				}
			}
			{
				if pbo.Votes != nil {
					(*goo).Votes = new(HeightVoteSet)
					err = (*goo).Votes.FromPBMessage(cdc, pbo.Votes)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ RoundStateSimple) GetTypeURL() (typeURL string) {
	return "/tm.RoundStateSimple"
}
func IsRoundStateSimpleReprEmpty(goor RoundStateSimple) (empty bool) {
	{
		empty = true
		{
			if goor.HeightRoundStep != "" {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.StartTime) {
				return false
			}
		}
		{
			if len(goor.ProposalBlockHash) != 0 {
				return false
			}
		}
		{
			if len(goor.LockedBlockHash) != 0 {
				return false
			}
		}
		{
			if len(goor.ValidBlockHash) != 0 {
				return false
			}
		}
		{
			if goor.Votes != nil {
				return false
			}
		}
	}
	return
}
func (goo PeerRoundState) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.PeerRoundState
	{
		if IsPeerRoundStateReprEmpty(goo) {
			var pbov *typespb.PeerRoundState
			msg = pbov
			return
		}
		pbo = new(typespb.PeerRoundState)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbo.Step = uint32(goo.Step)
		}
		{
			if !amino.IsEmptyTime(goo.StartTime) {
				pbo.StartTime = timestamppb.New(goo.StartTime)
			}
		}
		{
			pbo.Proposal = bool(goo.Proposal)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ProposalBlockPartsHeader.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ProposalBlockPartsHeader = pbom.(*typespb11.PartSetHeader)
		}
		{
			if goo.ProposalBlockParts != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ProposalBlockParts.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ProposalBlockParts = pbom.(*bitarraypb.BitArray)
				if pbo.ProposalBlockParts == nil {
					pbo.ProposalBlockParts = new(bitarraypb.BitArray)
				}
			}
		}
		{
			pbo.ProposalPOLRound = int64(goo.ProposalPOLRound)
		}
		{
			if goo.ProposalPOL != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ProposalPOL.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ProposalPOL = pbom.(*bitarraypb.BitArray)
				if pbo.ProposalPOL == nil {
					pbo.ProposalPOL = new(bitarraypb.BitArray)
				}
			}
		}
		{
			if goo.Prevotes != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Prevotes.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Prevotes = pbom.(*bitarraypb.BitArray)
				if pbo.Prevotes == nil {
					pbo.Prevotes = new(bitarraypb.BitArray)
				}
			}
		}
		{
			if goo.Precommits != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Precommits.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Precommits = pbom.(*bitarraypb.BitArray)
				if pbo.Precommits == nil {
					pbo.Precommits = new(bitarraypb.BitArray)
				}
			}
		}
		{
			pbo.LastCommitRound = int64(goo.LastCommitRound)
		}
		{
			if goo.LastCommit != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.LastCommit.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.LastCommit = pbom.(*bitarraypb.BitArray)
				if pbo.LastCommit == nil {
					pbo.LastCommit = new(bitarraypb.BitArray)
				}
			}
		}
		{
			pbo.CatchupCommitRound = int64(goo.CatchupCommitRound)
		}
		{
			if goo.CatchupCommit != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.CatchupCommit.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.CatchupCommit = pbom.(*bitarraypb.BitArray)
				if pbo.CatchupCommit == nil {
					pbo.CatchupCommit = new(bitarraypb.BitArray)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo PeerRoundState) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.PeerRoundState)
	msg = pbo
	return
}
func (goo *PeerRoundState) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.PeerRoundState = msg.(*typespb.PeerRoundState)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Step = RoundStepType(uint8(pbo.Step))
			}
			{
				(*goo).StartTime = pbo.StartTime.AsTime()
			}
			{
				(*goo).Proposal = bool(pbo.Proposal)
			}
			{
				if pbo.ProposalBlockPartsHeader != nil {
					err = (*goo).ProposalBlockPartsHeader.FromPBMessage(cdc, pbo.ProposalBlockPartsHeader)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ProposalBlockParts != nil {
					(*goo).ProposalBlockParts = new(bitarray.BitArray)
					err = (*goo).ProposalBlockParts.FromPBMessage(cdc, pbo.ProposalBlockParts)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).ProposalPOLRound = int(int(pbo.ProposalPOLRound))
			}
			{
				if pbo.ProposalPOL != nil {
					(*goo).ProposalPOL = new(bitarray.BitArray)
					err = (*goo).ProposalPOL.FromPBMessage(cdc, pbo.ProposalPOL)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Prevotes != nil {
					(*goo).Prevotes = new(bitarray.BitArray)
					err = (*goo).Prevotes.FromPBMessage(cdc, pbo.Prevotes)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Precommits != nil {
					(*goo).Precommits = new(bitarray.BitArray)
					err = (*goo).Precommits.FromPBMessage(cdc, pbo.Precommits)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).LastCommitRound = int(int(pbo.LastCommitRound))
			}
			{
				if pbo.LastCommit != nil {
					(*goo).LastCommit = new(bitarray.BitArray)
					err = (*goo).LastCommit.FromPBMessage(cdc, pbo.LastCommit)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).CatchupCommitRound = int(int(pbo.CatchupCommitRound))
			}
			{
				if pbo.CatchupCommit != nil {
					(*goo).CatchupCommit = new(bitarray.BitArray)
					err = (*goo).CatchupCommit.FromPBMessage(cdc, pbo.CatchupCommit)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ PeerRoundState) GetTypeURL() (typeURL string) {
	return "/tm.PeerRoundState"
}
func IsPeerRoundStateReprEmpty(goor PeerRoundState) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
				return false
			}
		}
		{
			if goor.Round != 0 {
				return false
			}
		}
		{
			if goor.Step != 0 {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.StartTime) {
				return false
			}
		}
		{
			if goor.Proposal != false {
				return false
			}
		}
		{
			e := types.IsPartSetHeaderReprEmpty(goor.ProposalBlockPartsHeader)
			if e == false {
				return false
			}
		}
		{
			if goor.ProposalBlockParts != nil {
				return false
			}
		}
		{
			if goor.ProposalPOLRound != 0 {
				return false
			}
		}
		{
			if goor.ProposalPOL != nil {
				return false
			}
		}
		{
			if goor.Prevotes != nil {
				return false
			}
		}
		{
			if goor.Precommits != nil {
				return false
			}
		}
		{
			if goor.LastCommitRound != 0 {
				return false
			}
		}
		{
			if goor.LastCommit != nil {
				return false
			}
		}
		{
			if goor.CatchupCommitRound != 0 {
				return false
			}
		}
		{
			if goor.CatchupCommit != nil {
				return false
			}
		}
	}
	return
}
func (goo HeightVoteSet) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.HeightVoteSet
	{
		if IsHeightVoteSetReprEmpty(goo) {
			var pbov *typespb.HeightVoteSet
			msg = pbov
			return
		}
		pbo = new(typespb.HeightVoteSet)
	}
	msg = pbo
	return
}
func (goo HeightVoteSet) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.HeightVoteSet)
	msg = pbo
	return
}
func (goo *HeightVoteSet) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.HeightVoteSet = msg.(*typespb.HeightVoteSet)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ HeightVoteSet) GetTypeURL() (typeURL string) {
	return "/tm.HeightVoteSet"
}
func IsHeightVoteSetReprEmpty(goor HeightVoteSet) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo EventNewRoundStep) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventNewRoundStep
	{
		if IsEventNewRoundStepReprEmpty(goo) {
			var pbov *typespb.EventNewRoundStep
			msg = pbov
			return
		}
		pbo = new(typespb.EventNewRoundStep)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.HRS.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.HRS = pbom.(*typespb.HRS)
		}
		{
			pbo.SecondsSinceStartTime = int64(goo.SecondsSinceStartTime)
		}
		{
			pbo.LastCommitRound = int64(goo.LastCommitRound)
		}
	}
	msg = pbo
	return
}
func (goo EventNewRoundStep) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventNewRoundStep)
	msg = pbo
	return
}
func (goo *EventNewRoundStep) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventNewRoundStep = msg.(*typespb.EventNewRoundStep)
	{
		if pbo != nil {
			{
				if pbo.HRS != nil {
					err = (*goo).HRS.FromPBMessage(cdc, pbo.HRS)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).SecondsSinceStartTime = int(int(pbo.SecondsSinceStartTime))
			}
			{
				(*goo).LastCommitRound = int(int(pbo.LastCommitRound))
			}
		}
	}
	return
}
func (_ EventNewRoundStep) GetTypeURL() (typeURL string) {
	return "/tm.EventNewRoundStep"
}
func IsEventNewRoundStepReprEmpty(goor EventNewRoundStep) (empty bool) {
	{
		empty = true
		{
			e := IsHRSReprEmpty(goor.HRS)
			if e == false {
				return false
			}
		}
		{
			if goor.SecondsSinceStartTime != 0 {
				return false
			}
		}
		{
			if goor.LastCommitRound != 0 {
				return false
			}
		}
	}
	return
}
func (goo EventNewValidBlock) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventNewValidBlock
	{
		if IsEventNewValidBlockReprEmpty(goo) {
			var pbov *typespb.EventNewValidBlock
			msg = pbov
			return
		}
		pbo = new(typespb.EventNewValidBlock)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.HRS.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.HRS = pbom.(*typespb.HRS)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockPartsHeader.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockPartsHeader = pbom.(*typespb12.PartSetHeader)
		}
		{
			if goo.BlockParts != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.BlockParts.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.BlockParts = pbom.(*bitarraypb.BitArray)
				if pbo.BlockParts == nil {
					pbo.BlockParts = new(bitarraypb.BitArray)
				}
			}
		}
		{
			pbo.IsCommit = bool(goo.IsCommit)
		}
	}
	msg = pbo
	return
}
func (goo EventNewValidBlock) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventNewValidBlock)
	msg = pbo
	return
}
func (goo *EventNewValidBlock) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventNewValidBlock = msg.(*typespb.EventNewValidBlock)
	{
		if pbo != nil {
			{
				if pbo.HRS != nil {
					err = (*goo).HRS.FromPBMessage(cdc, pbo.HRS)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.BlockPartsHeader != nil {
					err = (*goo).BlockPartsHeader.FromPBMessage(cdc, pbo.BlockPartsHeader)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.BlockParts != nil {
					(*goo).BlockParts = new(bitarray.BitArray)
					err = (*goo).BlockParts.FromPBMessage(cdc, pbo.BlockParts)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).IsCommit = bool(pbo.IsCommit)
			}
		}
	}
	return
}
func (_ EventNewValidBlock) GetTypeURL() (typeURL string) {
	return "/tm.EventNewValidBlock"
}
func IsEventNewValidBlockReprEmpty(goor EventNewValidBlock) (empty bool) {
	{
		empty = true
		{
			e := IsHRSReprEmpty(goor.HRS)
			if e == false {
				return false
			}
		}
		{
			e := types.IsPartSetHeaderReprEmpty(goor.BlockPartsHeader)
			if e == false {
				return false
			}
		}
		{
			if goor.BlockParts != nil {
				return false
			}
		}
		{
			if goor.IsCommit != false {
				return false
			}
		}
	}
	return
}
func (goo EventNewRound) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventNewRound
	{
		if IsEventNewRoundReprEmpty(goo) {
			var pbov *typespb.EventNewRound
			msg = pbov
			return
		}
		pbo = new(typespb.EventNewRound)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.HRS.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.HRS = pbom.(*typespb.HRS)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Proposer.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Proposer = pbom.(*typespb13.Validator)
		}
		{
			pbo.ProposerIndex = int64(goo.ProposerIndex)
		}
	}
	msg = pbo
	return
}
func (goo EventNewRound) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventNewRound)
	msg = pbo
	return
}
func (goo *EventNewRound) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventNewRound = msg.(*typespb.EventNewRound)
	{
		if pbo != nil {
			{
				if pbo.HRS != nil {
					err = (*goo).HRS.FromPBMessage(cdc, pbo.HRS)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Proposer != nil {
					err = (*goo).Proposer.FromPBMessage(cdc, pbo.Proposer)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).ProposerIndex = int(int(pbo.ProposerIndex))
			}
		}
	}
	return
}
func (_ EventNewRound) GetTypeURL() (typeURL string) {
	return "/tm.EventNewRound"
}
func IsEventNewRoundReprEmpty(goor EventNewRound) (empty bool) {
	{
		empty = true
		{
			e := IsHRSReprEmpty(goor.HRS)
			if e == false {
				return false
			}
		}
		{
			e := types.IsValidatorReprEmpty(goor.Proposer)
			if e == false {
				return false
			}
		}
		{
			if goor.ProposerIndex != 0 {
				return false
			}
		}
	}
	return
}
func (goo EventCompleteProposal) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventCompleteProposal
	{
		if IsEventCompleteProposalReprEmpty(goo) {
			var pbov *typespb.EventCompleteProposal
			msg = pbov
			return
		}
		pbo = new(typespb.EventCompleteProposal)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.HRS.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.HRS = pbom.(*typespb.HRS)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockID.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockID = pbom.(*typespb14.BlockID)
		}
	}
	msg = pbo
	return
}
func (goo EventCompleteProposal) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventCompleteProposal)
	msg = pbo
	return
}
func (goo *EventCompleteProposal) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventCompleteProposal = msg.(*typespb.EventCompleteProposal)
	{
		if pbo != nil {
			{
				if pbo.HRS != nil {
					err = (*goo).HRS.FromPBMessage(cdc, pbo.HRS)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.BlockID != nil {
					err = (*goo).BlockID.FromPBMessage(cdc, pbo.BlockID)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EventCompleteProposal) GetTypeURL() (typeURL string) {
	return "/tm.EventCompleteProposal"
}
func IsEventCompleteProposalReprEmpty(goor EventCompleteProposal) (empty bool) {
	{
		empty = true
		{
			e := IsHRSReprEmpty(goor.HRS)
			if e == false {
				return false
			}
		}
		{
			e := types.IsBlockIDReprEmpty(goor.BlockID)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo EventTimeoutPropose) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventTimeoutPropose
	{
		if IsEventTimeoutProposeReprEmpty(goo) {
			var pbov *typespb.EventTimeoutPropose
			msg = pbov
			return
		}
		pbo = new(typespb.EventTimeoutPropose)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.HRS.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.HRS = pbom.(*typespb.HRS)
		}
	}
	msg = pbo
	return
}
func (goo EventTimeoutPropose) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventTimeoutPropose)
	msg = pbo
	return
}
func (goo *EventTimeoutPropose) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventTimeoutPropose = msg.(*typespb.EventTimeoutPropose)
	{
		if pbo != nil {
			{
				if pbo.HRS != nil {
					err = (*goo).HRS.FromPBMessage(cdc, pbo.HRS)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EventTimeoutPropose) GetTypeURL() (typeURL string) {
	return "/tm.EventTimeoutPropose"
}
func IsEventTimeoutProposeReprEmpty(goor EventTimeoutPropose) (empty bool) {
	{
		empty = true
		{
			e := IsHRSReprEmpty(goor.HRS)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo EventTimeoutWait) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *typespb.EventTimeoutWait
	{
		if IsEventTimeoutWaitReprEmpty(goo) {
			var pbov *typespb.EventTimeoutWait
			msg = pbov
			return
		}
		pbo = new(typespb.EventTimeoutWait)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.HRS.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.HRS = pbom.(*typespb.HRS)
		}
	}
	msg = pbo
	return
}
func (goo EventTimeoutWait) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(typespb.EventTimeoutWait)
	msg = pbo
	return
}
func (goo *EventTimeoutWait) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *typespb.EventTimeoutWait = msg.(*typespb.EventTimeoutWait)
	{
		if pbo != nil {
			{
				if pbo.HRS != nil {
					err = (*goo).HRS.FromPBMessage(cdc, pbo.HRS)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EventTimeoutWait) GetTypeURL() (typeURL string) {
	return "/tm.EventTimeoutWait"
}
func IsEventTimeoutWaitReprEmpty(goor EventTimeoutWait) (empty bool) {
	{
		empty = true
		{
			e := IsHRSReprEmpty(goor.HRS)
			if e == false {
				return false
			}
		}
	}
	return
}
