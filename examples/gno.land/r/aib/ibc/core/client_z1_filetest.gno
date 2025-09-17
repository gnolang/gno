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
		apphash = sha256.Sum256([]byte("apphash"))
		val1    = &tendermint.Validator{
			Address:     genAddr("val1"),
			PubKey:      genPubkey("val1_pubkey"),
			VotingPower: 2,
		}
		val2 = &tendermint.Validator{
			Address:     genAddr("val2"),
			PubKey:      genPubkey("val2_pubkey"),
			VotingPower: 3,
		}
		valset = &tendermint.ValidatorSet{
			Validators:       []*tendermint.Validator{val1, val2},
			Proposer:         val1,
			TotalVotingPower: 5,
		}
		valsethash     = valset.Hash()
		consensusState = tendermint.ConsensusState{
			Timestamp:          time.Now(),
			Root:               tendermint.MerkleRoot{Hash: apphash[:]},
			NextValidatorsHash: valsethash,
		}
	)
	clientID1 := core.CreateClient(cross, lightclient.Tendermint, clientState, consensusState)
	println("CreateClient", clientID1)
	clientID2 := core.CreateClient(cross, lightclient.Tendermint, clientState, consensusState)
	println("CreateClient", clientID2)

	// RegisterCounterparty on clientID1
	core.RegisterCounterparty(cross, clientID1, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	println("\n----------- assert render clients")
	println(core.Render("clients"))
	println("----------- assert render clients/unknown")
	println(core.Render("clients/unknown"))
	println("\n----------- assert render clients/07-tendermint-1")
	println(core.Render("clients/" + clientID1))

	// Update clientID1
	// TODO move it to an other filetest
	var (
		blockhash      = sha256.Sum256([]byte("block-5"))
		parsethash     = sha256.Sum256([]byte("parset"))
		nextValsethash = valsethash
		consensushash  = sha256.Sum256([]byte("consensus"))

		msgHeader = &tendermint.MsgHeader{
			Header: &tendermint.Header{
				Version: tendermint.Consensus{
					Block: tendermint.BlockProtocol,
					App:   1, //NOTE no idea what to put there, sounds unused
				},
				ChainID: "chain-id-2",
				Height:  5,
				Time:    time.Now().Add(time.Minute),
				LastBlockID: tendermint.BlockID{
					Hash: blockhash[:],
					PartSetHeader: tendermint.PartSetHeader{
						Total: 2,
						Hash:  parsethash[:],
					},
				},
				LastCommitHash:     apphash[:], //FIXME
				DataHash:           apphash[:], //FIXME
				ValidatorsHash:     valsethash,
				NextValidatorsHash: nextValsethash,
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
			ValidatorSet:      valset, // validator set hasn't changed
			TrustedHeight:     types.Height{2, 1},
			TrustedValidators: valset, // hash of this must correspond to the consensusState.NextValidatorHash mapped to TrustedHeight
		}
	)
	core.UpdateClient(cross, clientID1, msgHeader)

	println("----------- assert render clients/07-tendermint-1 after update client")
	println(core.Render("clients/" + clientID1))

}

// TODO move this in a testing package
func genAddr(prefix string) []byte {
	return genBytes(prefix, tendermint.AddressSize)
}

func genSignature(prefix string) []byte {
	return genBytes(prefix, tendermint.MaxSignatureSize)
}

func genPubkey(prefix string) []byte {
	return genBytes(prefix, 32)
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
// ----------- assert render clients
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: chain-id-2
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 2/1
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
// ## Client 07-tendermint-2
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID:
// - CounterpartyMerklePrefix:
// - Client state:
//   - ChainId: chain-id-2
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 2/1
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
//
// ----------- assert render clients/unknown
// client unknown not found
//
// ----------- assert render clients/07-tendermint-1
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: chain-id-2
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 2/1
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
//
// ----------- assert render clients/07-tendermint-1 after update client
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: chain-id-2
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 2/5
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
