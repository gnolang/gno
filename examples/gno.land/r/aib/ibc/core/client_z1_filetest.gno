package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

func main() {
	// CreateClient
	var (
		clientState = tendermint.ClientState{
			ChainId:                      "chain-id-2",
			TrustLevel:                   tendermint.Fraction{2, 3},
			UnbondingPeriod:              time.Hour * 3,
			TrustingPeriod:               time.Hour,
			MaxClockDrift:                time.Hour,
			LatestHeight:                 types.Height{2, 1},
			ProofSpecs:                   types.GetSDKProofSpecs(),
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
		}
		apphash        = sha256.Sum256([]byte("apphash"))
		valhash        = sha256.Sum256([]byte("validators"))
		consensusState = tendermint.ConsensusState{
			Timestamp:          time.Now(),
			Root:               tendermint.MerkleRoot{Hash: apphash[:]},
			NextValidatorsHash: valhash[:],
		}
	)
	clientID1 := core.CreateClient(cross, lightclient.Tendermint, clientState, consensusState)
	println("CreateClient", clientID1)
	clientID2 := core.CreateClient(cross, lightclient.Tendermint, clientState, consensusState)
	println("CreateClient", clientID2)

	// RegisterCounterparty on clientID1
	core.RegisterCounterparty(cross, clientID1, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	println(core.Render("clients"))
	println("-----------")
	println(core.Render("clients/1"))
	println("-----------")
	println(core.Render("clients/" + clientID1))

	// Update clientID1
	// TODO move it to an other filetest
	var (
		blockhash     = sha256.Sum256([]byte("block-5"))
		parsethash    = sha256.Sum256([]byte("parset"))
		nextvalhash   = sha256.Sum256([]byte("nextval"))
		consensushash = sha256.Sum256([]byte("consensus"))

		header = &tendermint.MsgHeader{
			Header: &tendermint.Header{
				Version: tendermint.Consensus{
					Block: tendermint.BlockProtocol,
					App:   1, //NOTE no idea what to put there, sounds unused
				},
				ChainID: "chain-id-2",
				Height:  5,
				Time:    time.Now(),
				LastBlockID: tendermint.BlockID{
					Hash: blockhash[:],
					PartSetHeader: tendermint.PartSetHeader{
						Total: 2,
						Hash:  parsethash[:],
					},
				},
				LastCommitHash:     apphash[:], //FIXME
				DataHash:           apphash[:], //FIXME
				ValidatorsHash:     valhash[:],
				NextValidatorsHash: nextvalhash[:],
				ConsensusHash:      consensushash[:],
				AppHash:            apphash[:],
				LastResultsHash:    nil, //FIXME?
				EvidenceHash:       nil, //FIXME?
				ProposerAddress:    genAddr("proposer"),
			},
			Commit: &tendermint.Commit{
				Height: 5,
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
						ValidatorAddress: genAddr("val1"),
						Signature:        genSignature("valsign1"), // XXX need more than random bytes ?
					},
				},
			},
			ValidatorSet:      &tendermint.ValidatorSet{},
			TrustedHeight:     types.Height{2, 1},
			TrustedValidators: &tendermint.ValidatorSet{}, // hash of this must correspond to consensusState.NextValidatorHash passed in CreateClient
		}
	)
	core.UpdateClient(cross, clientID1, header)
	// TODO assertion of updated client

}

// TODO move this in a testing package
func genAddr(prefix string) []byte {
	return genBytes(prefix, tendermint.AddressSize)
}

func genSignature(prefix string) []byte {
	return genBytes(prefix, tendermint.MaxSignatureSize)
}

func genBytes(prefix string, size int) []byte {
	bz := []byte(prefix)
	// suffix with zeros to make it size bytes long
	for len(bz) < size {
		bz = append(bz, 0)
	}
	return bz
}

// Output:
// CreateClient 07-tendermint-1
// CreateClient 07-tendermint-2
//
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
//
// ## Client 07-tendermint-2
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID:
// - CounterpartyMerklePrefix:
//
// -----------
// client 1 not found
// -----------
//
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
