package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

func main() {
	var (
		clientID      = "07-tendermint-1"                  // TODO update
		chainID       = "chain-id-2"                       // TODO update
		timestamp     = time.Now()                         // TODO update
		blockhash     = sha256.Sum256([]byte("block-5"))   // TODO update
		parsethash    = sha256.Sum256([]byte("parset"))    // TODO update
		consensushash = sha256.Sum256([]byte("consensus")) // TODO update
		apphash       = sha256.Sum256([]byte("apphash"))   // TODO update
		height        = types.Height{2, 5}                 // TODO update
		valset        = &tendermint.ValidatorSet{          // TODO update
			Validators: []*tendermint.Validator{
				{
					Address:     testing.GenAddr("val1"),
					PubKey:      testing.GenPubkey("val1_pubkey"),
					VotingPower: 2,
				},
				{
					Address:     testing.GenAddr("val2"),
					PubKey:      testing.GenPubkey("val2_pubkey"),
					VotingPower: 3,
				},
			},
			Proposer: &tendermint.Validator{
				Address:     testing.GenAddr("val1"),
				PubKey:      testing.GenPubkey("val1_pubkey"),
				VotingPower: 2,
			},
			TotalVotingPower: 5,
		}
		trustedHeight = types.Height{2, 2}        // TODO update
		trustedValset = &tendermint.ValidatorSet{ // TODO update
			Validators: []*tendermint.Validator{
				{
					Address:     testing.GenAddr("val1"),
					PubKey:      testing.GenPubkey("val1_pubkey"),
					VotingPower: 2,
				},
				{
					Address:     testing.GenAddr("val2"),
					PubKey:      testing.GenPubkey("val2_pubkey"),
					VotingPower: 3,
				},
			},
			Proposer: &tendermint.Validator{
				Address:     testing.GenAddr("val1"),
				PubKey:      testing.GenPubkey("val1_pubkey"),
				VotingPower: 2,
			},
			TotalVotingPower: 5,
		}
		msgHeader = &tendermint.MsgHeader{
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
				LastCommitHash:     apphash[:], //FIXME
				DataHash:           apphash[:], //FIXME
				ValidatorsHash:     valset.Hash(),
				NextValidatorsHash: valset.Hash(),
				ConsensusHash:      consensushash[:],
				AppHash:            apphash[:],
				LastResultsHash:    nil,                         //FIXME?
				EvidenceHash:       nil,                         //FIXME?
				ProposerAddress:    testing.GenAddr("proposer"), // TODO update
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
						BlockIDFlag:      tendermint.BlockIDFlagCommit,
						ValidatorAddress: testing.GenAddr("val1"),          // TODO update
						Signature:        testing.GenSignature("valsign1"), // TODO update
					},
					{
						BlockIDFlag:      tendermint.BlockIDFlagCommit,
						ValidatorAddress: testing.GenAddr("val2"),          // TODO update
						Signature:        testing.GenSignature("valsign2"), // TODO update
					},
				},
			},
			ValidatorSet:      valset,
			TrustedHeight:     trustedHeight,
			TrustedValidators: trustedValset,
		}
	)
	core.UpdateClient(cross, clientID, msgHeader)
}

// Error:
// client 07-tendermint-1 not found
