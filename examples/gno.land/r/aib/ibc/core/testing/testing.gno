package testing

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
)

func GenClientState(chainID string) tendermint.ClientState {
	return tendermint.ClientState{
		ChainId:                      chainID,
		TrustLevel:                   tendermint.DefaultTrustLevel,
		UnbondingPeriod:              time.Hour * 3,
		TrustingPeriod:               time.Hour,
		MaxClockDrift:                time.Hour,
		LatestHeight:                 types.NewHeight(2, 1),
		ProofSpecs:                   types.GetSDKProofSpecs(),
		AllowUpdateAfterExpiry:       false,
		AllowUpdateAfterMisbehaviour: false,
	}
}

func GenValset() *tendermint.ValidatorSet {
	var (
		val1 = &tendermint.Validator{
			Address:     GenAddr("val1"),
			PubKey:      GenPubkey("val1_pubkey"),
			VotingPower: 2,
		}
		val2 = &tendermint.Validator{
			Address:     GenAddr("val2"),
			PubKey:      GenPubkey("val2_pubkey"),
			VotingPower: 3,
		}
	)
	return &tendermint.ValidatorSet{
		Validators:       []*tendermint.Validator{val1, val2},
		Proposer:         val1,
		TotalVotingPower: 5,
	}
}

func GenConsensusState(timestamp time.Time, apphash, valsethash []byte) tendermint.ConsensusState {
	return tendermint.ConsensusState{
		Timestamp:          timestamp,
		Root:               tendermint.MerkleRoot{Hash: apphash},
		NextValidatorsHash: valsethash,
	}
}

func GenMsgHeader(chainID string, timestamp time.Time, apphash []byte,
	height, trustedHeight types.Height,
	valset, trustedValset *tendermint.ValidatorSet) *tendermint.MsgHeader {
	var (
		blockhash     = sha256.Sum256([]byte("block-5"))
		parsethash    = sha256.Sum256([]byte("parset"))
		consensushash = sha256.Sum256([]byte("consensus"))
	)
	return &tendermint.MsgHeader{
		Header: &tendermint.Header{
			Version: tendermint.Consensus{
				Block: tendermint.BlockProtocol,
				App:   1, //NOTE no idea what to put there, sounds unused
			},
			ChainID: chainID,
			Height:  height.RevisionHeight,
			Time:    timestamp,
			LastBlockID: tendermint.BlockID{
				Hash: blockhash[:],
				PartSetHeader: tendermint.PartSetHeader{
					Total: 2,
					Hash:  parsethash[:],
				},
			},
			LastCommitHash:     apphash, //FIXME
			DataHash:           apphash, //FIXME
			ValidatorsHash:     valset.Hash(),
			NextValidatorsHash: valset.Hash(),
			ConsensusHash:      consensushash[:],
			AppHash:            apphash,
			LastResultsHash:    nil, //FIXME?
			EvidenceHash:       nil, //FIXME?
			ProposerAddress:    GenAddr("proposer"),
		},
		Commit: &tendermint.Commit{
			Height: height.RevisionHeight,
			Round:  1,
			BlockID: tendermint.BlockID{
				Hash: blockhash[:],
				PartSetHeader: tendermint.PartSetHeader{
					Total: 5,
					Hash:  parsethash[:], //FIXME
				},
			},
			Signatures: []tendermint.CommitSig{
				{
					BlockIdFlag:      tendermint.BlockIDFlagCommit,
					ValidatorAddress: GenAddr("val1"),
					Signature:        GenSignature("valsign1"), // XXX need more than random bytes ?
				},
			},
		},
		ValidatorSet:      valset,
		TrustedHeight:     trustedHeight,
		TrustedValidators: trustedValset,
	}
}

func GenAddr(prefix string) []byte {
	return GenBytes(prefix, tendermint.AddressSize)
}

func GenSignature(prefix string) []byte {
	return GenBytes(prefix, tendermint.MaxSignatureSize)
}

func GenPubkey(prefix string) []byte {
	return GenBytes(prefix, 32)
}

func GenBytes(prefix string, size int) []byte {
	bz := []byte(prefix)
	// suffix with zeros to make it size bytes long
	for len(bz) < size {
		bz = append(bz, 0)
	}
	return bz
}
