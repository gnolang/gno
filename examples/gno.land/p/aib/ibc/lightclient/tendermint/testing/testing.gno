package testing

import (
	"crypto/sha256"
	"encoding/base64"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
)

func NewClientState(chainID string, height types.Height) tendermint.ClientState {
	return tendermint.ClientState{
		ChainId:                      chainID,
		TrustLevel:                   tendermint.DefaultTrustLevel,
		UnbondingPeriod:              time.Hour * 3,
		TrustingPeriod:               time.Hour,
		MaxClockDrift:                time.Hour,
		LatestHeight:                 height,
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
	return NewValset(val1, val2)
}

func NewValset(vals ...*tendermint.Validator) *tendermint.ValidatorSet {
	var total int64
	for _, v := range vals {
		total += v.VotingPower
	}
	return &tendermint.ValidatorSet{
		Validators:       vals,
		Proposer:         vals[0],
		TotalVotingPower: total,
	}
}

func NewValidator(addr, pubkey string, vp int64) *tendermint.Validator {
	return &tendermint.Validator{
		Address:     B64Dec(addr),
		PubKey:      B64Dec(pubkey),
		VotingPower: vp,
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
	height uint64, trustedHeight types.Height,
	valset, trustedValset *tendermint.ValidatorSet) *tendermint.MsgHeader {
	var (
		blockhash     = GenHash("blockhash")
		parsethash    = GenHash("parsethash")
		consensushash = GenHash("consensushash")
		signatures    = []tendermint.CommitSig{
			{
				BlockIDFlag:      tendermint.BlockIDFlagCommit,
				ValidatorAddress: valset.Validators[0].Address,
				Timestamp:        ToTime("2025-09-25T07:55:57.306746166Z"),
				Signature:        B64Dec("qtv1z4S2Q6T87vGQo0lrjRZqv9PrHIji4pTyviMnVyGx9td6eySdzwQwCthwmihU48ebNlFiMlFJ0CT891UmDg=="),
			},
			{
				BlockIDFlag:      tendermint.BlockIDFlagCommit,
				ValidatorAddress: valset.Validators[1].Address,
				Timestamp:        ToTime("2025-09-25T07:55:57.310583641Z"),
				Signature:        B64Dec("Q5E6Kjma00n/T98rC9qJmoB6JTGFX/IB+mDVs4Wd1h0eJ8fabY/6oI8zdoU6/7W6VR6wjpHyWBsJrpGT6C0LCg=="),
			},
		}
	)
	return NewMsgHeader(chainID, timestamp,
		apphash, blockhash, parsethash, consensushash,
		height, trustedHeight, valset, trustedValset, signatures,
	)
}

func NewMsgHeader(chainID string, timestamp time.Time,
	apphash, blockhash, parsethash, consensushash []byte,
	height uint64, trustedHeight types.Height,
	valset, trustedValset *tendermint.ValidatorSet,
	signatures []tendermint.CommitSig) *tendermint.MsgHeader {
	return &tendermint.MsgHeader{
		Header: &tendermint.Header{
			Version: tendermint.Consensus{
				Block: tendermint.BlockProtocol,
				App:   0,
			},
			ChainID: chainID,
			Height:  height,
			Time:    timestamp,
			LastBlockID: tendermint.BlockID{
				Hash: blockhash,
				PartSetHeader: tendermint.PartSetHeader{
					Total: 1,
					Hash:  parsethash,
				},
			},
			LastCommitHash:     apphash, //FIXME
			DataHash:           apphash, //FIXME
			ValidatorsHash:     valset.Hash(),
			NextValidatorsHash: trustedValset.Hash(),
			ConsensusHash:      consensushash,
			AppHash:            apphash,
			LastResultsHash:    nil, //FIXME?
			EvidenceHash:       nil, //FIXME?
			ProposerAddress:    valset.Proposer.Address,
		},
		Commit: &tendermint.Commit{
			Height: height,
			Round:  0,
			BlockID: tendermint.BlockID{
				Hash: blockhash,
				PartSetHeader: tendermint.PartSetHeader{
					Total: 1,
					Hash:  parsethash,
				},
			},
			Signatures: signatures,
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

func GenHash(prefix string) []byte {
	bz := sha256.Sum256(GenBytes(prefix, 64))
	return bz[:]
}

func B64Dec(s string) []byte {
	bz, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}

func ToTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}
