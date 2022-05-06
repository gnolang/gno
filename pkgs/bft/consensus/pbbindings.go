package consensus

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	consensuspb "github.com/gnolang/gno/pkgs/bft/consensus/pb"
	types "github.com/gnolang/gno/pkgs/bft/consensus/types"
	typespb "github.com/gnolang/gno/pkgs/bft/types/pb"
	bitarraypb "github.com/gnolang/gno/pkgs/bitarray/pb"
	bitarray "github.com/gnolang/gno/pkgs/bitarray"
	types1 "github.com/gnolang/gno/pkgs/bft/types"
	types2 "github.com/gnolang/gno/pkgs/bft/types"
	types3 "github.com/gnolang/gno/pkgs/bft/types"
	types4 "github.com/gnolang/gno/pkgs/bft/types"
	types5 "github.com/gnolang/gno/pkgs/bft/types"
	types6 "github.com/gnolang/gno/pkgs/bft/types"
	types7 "github.com/gnolang/gno/pkgs/bft/types"
	types8 "github.com/gnolang/gno/pkgs/bft/types"
	types9 "github.com/gnolang/gno/pkgs/bft/types"
	typespb1 "github.com/gnolang/gno/pkgs/bft/consensus/types/pb"
	anypb "google.golang.org/protobuf/types/known/anypb"
	crypto "github.com/gnolang/gno/pkgs/crypto"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
)

func (goo NewRoundStepMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.NewRoundStepMessage
	{
		if IsNewRoundStepMessageReprEmpty(goo) {
			var pbov *consensuspb.NewRoundStepMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.NewRoundStepMessage)
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
			pbo.SecondsSinceStartTime = int64(goo.SecondsSinceStartTime)
		}
		{
			pbo.LastCommitRound = int64(goo.LastCommitRound)
		}
	}
	msg = pbo
	return
}
func (goo NewRoundStepMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.NewRoundStepMessage)
	msg = pbo
	return
}
func (goo *NewRoundStepMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.NewRoundStepMessage = msg.(*consensuspb.NewRoundStepMessage)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Step = types.RoundStepType(uint8(pbo.Step))
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
func (_ NewRoundStepMessage) GetTypeURL() (typeURL string) {
	return "/tm.NewRoundStepMessage"
}
func IsNewRoundStepMessageReprEmpty(goor NewRoundStepMessage) (empty bool) {
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
func (goo NewValidBlockMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.NewValidBlockMessage
	{
		if IsNewValidBlockMessageReprEmpty(goo) {
			var pbov *consensuspb.NewValidBlockMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.NewValidBlockMessage)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockPartsHeader.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockPartsHeader = pbom.(*typespb.PartSetHeader)
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
func (goo NewValidBlockMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.NewValidBlockMessage)
	msg = pbo
	return
}
func (goo *NewValidBlockMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.NewValidBlockMessage = msg.(*consensuspb.NewValidBlockMessage)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
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
func (_ NewValidBlockMessage) GetTypeURL() (typeURL string) {
	return "/tm.NewValidBlockMessage"
}
func IsNewValidBlockMessageReprEmpty(goor NewValidBlockMessage) (empty bool) {
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
			e := types1.IsPartSetHeaderReprEmpty(goor.BlockPartsHeader)
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
func (goo ProposalMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.ProposalMessage
	{
		if IsProposalMessageReprEmpty(goo) {
			var pbov *consensuspb.ProposalMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.ProposalMessage)
		{
			if goo.Proposal != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Proposal.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Proposal = pbom.(*typespb.Proposal)
				if pbo.Proposal == nil {
					pbo.Proposal = new(typespb.Proposal)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo ProposalMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.ProposalMessage)
	msg = pbo
	return
}
func (goo *ProposalMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.ProposalMessage = msg.(*consensuspb.ProposalMessage)
	{
		if pbo != nil {
			{
				if pbo.Proposal != nil {
					(*goo).Proposal = new(types2.Proposal)
					err = (*goo).Proposal.FromPBMessage(cdc, pbo.Proposal)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ProposalMessage) GetTypeURL() (typeURL string) {
	return "/tm.ProposalMessage"
}
func IsProposalMessageReprEmpty(goor ProposalMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Proposal != nil {
				return false
			}
		}
	}
	return
}
func (goo ProposalPOLMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.ProposalPOLMessage
	{
		if IsProposalPOLMessageReprEmpty(goo) {
			var pbov *consensuspb.ProposalPOLMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.ProposalPOLMessage)
		{
			pbo.Height = int64(goo.Height)
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
	}
	msg = pbo
	return
}
func (goo ProposalPOLMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.ProposalPOLMessage)
	msg = pbo
	return
}
func (goo *ProposalPOLMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.ProposalPOLMessage = msg.(*consensuspb.ProposalPOLMessage)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
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
		}
	}
	return
}
func (_ ProposalPOLMessage) GetTypeURL() (typeURL string) {
	return "/tm.ProposalPOLMessage"
}
func IsProposalPOLMessageReprEmpty(goor ProposalPOLMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Height != 0 {
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
	}
	return
}
func (goo BlockPartMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.BlockPartMessage
	{
		if IsBlockPartMessageReprEmpty(goo) {
			var pbov *consensuspb.BlockPartMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.BlockPartMessage)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			if goo.Part != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Part.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Part = pbom.(*typespb.Part)
				if pbo.Part == nil {
					pbo.Part = new(typespb.Part)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo BlockPartMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.BlockPartMessage)
	msg = pbo
	return
}
func (goo *BlockPartMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.BlockPartMessage = msg.(*consensuspb.BlockPartMessage)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				if pbo.Part != nil {
					(*goo).Part = new(types3.Part)
					err = (*goo).Part.FromPBMessage(cdc, pbo.Part)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ BlockPartMessage) GetTypeURL() (typeURL string) {
	return "/tm.BlockPartMessage"
}
func IsBlockPartMessageReprEmpty(goor BlockPartMessage) (empty bool) {
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
			if goor.Part != nil {
				return false
			}
		}
	}
	return
}
func (goo VoteMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.VoteMessage
	{
		if IsVoteMessageReprEmpty(goo) {
			var pbov *consensuspb.VoteMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.VoteMessage)
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
func (goo VoteMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.VoteMessage)
	msg = pbo
	return
}
func (goo *VoteMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.VoteMessage = msg.(*consensuspb.VoteMessage)
	{
		if pbo != nil {
			{
				if pbo.Vote != nil {
					(*goo).Vote = new(types4.Vote)
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
func (_ VoteMessage) GetTypeURL() (typeURL string) {
	return "/tm.VoteMessage"
}
func IsVoteMessageReprEmpty(goor VoteMessage) (empty bool) {
	{
		empty = true
		{
			if goor.Vote != nil {
				return false
			}
		}
	}
	return
}
func (goo HasVoteMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.HasVoteMessage
	{
		if IsHasVoteMessageReprEmpty(goo) {
			var pbov *consensuspb.HasVoteMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.HasVoteMessage)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbo.Type = uint32(goo.Type)
		}
		{
			pbo.Index = int64(goo.Index)
		}
	}
	msg = pbo
	return
}
func (goo HasVoteMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.HasVoteMessage)
	msg = pbo
	return
}
func (goo *HasVoteMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.HasVoteMessage = msg.(*consensuspb.HasVoteMessage)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Type = types5.SignedMsgType(uint8(pbo.Type))
			}
			{
				(*goo).Index = int(int(pbo.Index))
			}
		}
	}
	return
}
func (_ HasVoteMessage) GetTypeURL() (typeURL string) {
	return "/tm.HasVoteMessage"
}
func IsHasVoteMessageReprEmpty(goor HasVoteMessage) (empty bool) {
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
			if goor.Type != 0 {
				return false
			}
		}
		{
			if goor.Index != 0 {
				return false
			}
		}
	}
	return
}
func (goo VoteSetMaj23Message) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.VoteSetMaj23Message
	{
		if IsVoteSetMaj23MessageReprEmpty(goo) {
			var pbov *consensuspb.VoteSetMaj23Message
			msg = pbov
			return
		}
		pbo = new(consensuspb.VoteSetMaj23Message)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbo.Type = uint32(goo.Type)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.BlockID.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.BlockID = pbom.(*typespb.BlockID)
		}
	}
	msg = pbo
	return
}
func (goo VoteSetMaj23Message) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.VoteSetMaj23Message)
	msg = pbo
	return
}
func (goo *VoteSetMaj23Message) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.VoteSetMaj23Message = msg.(*consensuspb.VoteSetMaj23Message)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Type = types6.SignedMsgType(uint8(pbo.Type))
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
func (_ VoteSetMaj23Message) GetTypeURL() (typeURL string) {
	return "/tm.VoteSetMaj23Message"
}
func IsVoteSetMaj23MessageReprEmpty(goor VoteSetMaj23Message) (empty bool) {
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
			if goor.Type != 0 {
				return false
			}
		}
		{
			e := types7.IsBlockIDReprEmpty(goor.BlockID)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo VoteSetBitsMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.VoteSetBitsMessage
	{
		if IsVoteSetBitsMessageReprEmpty(goo) {
			var pbov *consensuspb.VoteSetBitsMessage
			msg = pbov
			return
		}
		pbo = new(consensuspb.VoteSetBitsMessage)
		{
			pbo.Height = int64(goo.Height)
		}
		{
			pbo.Round = int64(goo.Round)
		}
		{
			pbo.Type = uint32(goo.Type)
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
			if goo.Votes != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Votes.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Votes = pbom.(*bitarraypb.BitArray)
				if pbo.Votes == nil {
					pbo.Votes = new(bitarraypb.BitArray)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo VoteSetBitsMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.VoteSetBitsMessage)
	msg = pbo
	return
}
func (goo *VoteSetBitsMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.VoteSetBitsMessage = msg.(*consensuspb.VoteSetBitsMessage)
	{
		if pbo != nil {
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Type = types8.SignedMsgType(uint8(pbo.Type))
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
				if pbo.Votes != nil {
					(*goo).Votes = new(bitarray.BitArray)
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
func (_ VoteSetBitsMessage) GetTypeURL() (typeURL string) {
	return "/tm.VoteSetBitsMessage"
}
func IsVoteSetBitsMessageReprEmpty(goor VoteSetBitsMessage) (empty bool) {
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
			if goor.Type != 0 {
				return false
			}
		}
		{
			e := types9.IsBlockIDReprEmpty(goor.BlockID)
			if e == false {
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
func (goo newRoundStepInfo) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.NewRoundStepInfo
	{
		if IsnewRoundStepInfoReprEmpty(goo) {
			var pbov *consensuspb.NewRoundStepInfo
			msg = pbov
			return
		}
		pbo = new(consensuspb.NewRoundStepInfo)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.HRS.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.HRS = pbom.(*typespb1.HRS)
		}
	}
	msg = pbo
	return
}
func (goo newRoundStepInfo) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.NewRoundStepInfo)
	msg = pbo
	return
}
func (goo *newRoundStepInfo) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.NewRoundStepInfo = msg.(*consensuspb.NewRoundStepInfo)
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
func (_ newRoundStepInfo) GetTypeURL() (typeURL string) {
	return "/tm.newRoundStepInfo"
}
func IsnewRoundStepInfoReprEmpty(goor newRoundStepInfo) (empty bool) {
	{
		empty = true
		{
			e := types.IsHRSReprEmpty(goor.HRS)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo msgInfo) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.MsgInfo
	{
		if IsmsgInfoReprEmpty(goo) {
			var pbov *consensuspb.MsgInfo
			msg = pbov
			return
		}
		pbo = new(consensuspb.MsgInfo)
		{
			if goo.Msg != nil {
				typeUrl := cdc.GetTypeURL(goo.Msg)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.Msg)
				if err != nil {
					return
				}
				pbo.Msg = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			pbo.PeerID = string(goo.PeerID)
		}
	}
	msg = pbo
	return
}
func (goo msgInfo) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.MsgInfo)
	msg = pbo
	return
}
func (goo *msgInfo) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.MsgInfo = msg.(*consensuspb.MsgInfo)
	{
		if pbo != nil {
			{
				typeUrl := pbo.Msg.TypeUrl
				bz := pbo.Msg.Value
				goorp := &(*goo).Msg
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				(*goo).PeerID = crypto.ID(pbo.PeerID)
			}
		}
	}
	return
}
func (_ msgInfo) GetTypeURL() (typeURL string) {
	return "/tm.msgInfo"
}
func IsmsgInfoReprEmpty(goor msgInfo) (empty bool) {
	{
		empty = true
		{
			if goor.Msg != nil {
				return false
			}
		}
		{
			if goor.PeerID != "" {
				return false
			}
		}
	}
	return
}
func (goo timeoutInfo) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *consensuspb.TimeoutInfo
	{
		if IstimeoutInfoReprEmpty(goo) {
			var pbov *consensuspb.TimeoutInfo
			msg = pbov
			return
		}
		pbo = new(consensuspb.TimeoutInfo)
		{
			if goo.Duration.Nanoseconds() != 0 {
				pbo.Duration = durationpb.New(goo.Duration)
			}
		}
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
func (goo timeoutInfo) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(consensuspb.TimeoutInfo)
	msg = pbo
	return
}
func (goo *timeoutInfo) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *consensuspb.TimeoutInfo = msg.(*consensuspb.TimeoutInfo)
	{
		if pbo != nil {
			{
				(*goo).Duration = pbo.Duration.AsDuration()
			}
			{
				(*goo).Height = int64(pbo.Height)
			}
			{
				(*goo).Round = int(int(pbo.Round))
			}
			{
				(*goo).Step = types.RoundStepType(uint8(pbo.Step))
			}
		}
	}
	return
}
func (_ timeoutInfo) GetTypeURL() (typeURL string) {
	return "/tm.timeoutInfo"
}
func IstimeoutInfoReprEmpty(goor timeoutInfo) (empty bool) {
	{
		empty = true
		{
			if goor.Duration != 0 {
				return false
			}
		}
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
