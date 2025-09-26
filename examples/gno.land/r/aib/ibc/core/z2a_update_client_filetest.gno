package main

import (
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// UpdateClient success with adjacent height
func main() {
	// CreateClient and RegisterCounterparty
	var (
		chainID     = "osmosis-1"
		height      = uint64(44873551)
		clientState = testing.NewClientState(chainID, types.NewHeight(1, height))
		apphash     = testing.B64Dec("DAHBO3lYaIFu9XmplThmF7ijXG7YwDZawGSDKKCHh7c=")
		val1        = testing.NewValidator("y1pjuR6PTujbk1lCy+JXJGNkeeA=",
			"6Nz09YGHzwWxjczG0IhK4Iv0qY2IcX0P/5KitvRXTUc=", 2)
		val2 = testing.NewValidator("ZraWZuv3dufry+GXq6RmpxLicHY=",
			"wB25StLxbzmD0uTiFiH6xySZd0H13kyanNUvvlUpa34=", 3)
		valset         = testing.NewValset(val1, val2)
		consensusState = testing.GenConsensusState(time.Now(), apphash, valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	// Update clientID
	var (
		newHeight     = clientState.LatestHeight.RevisionHeight + 1 // adjacent heigth
		newTimestamp  = consensusState.Timestamp.Add(time.Minute)
		blockhash     = testing.B64Dec("NpiImIJoaSaIucwNs5cqpgMsL/8wxEPYC3P0jA5aQSI=")
		parsethash    = testing.B64Dec("QqzwnLzvixIcUz+hPeUQjDV6NaLkFRKXACCxJIrBHzw=")
		consensushash = testing.B64Dec("JIS7qLmeUx4aP2QiNi59PuuZqDJRGZaKNV8FwJ7t1As=")
		trustedValset = testing.NewValset(val1, val2)
		trustedHeight = clientState.LatestHeight
		signatures    = []tendermint.CommitSig{
			{
				BlockIDFlag:      tendermint.BlockIDFlagCommit,
				ValidatorAddress: valset.Validators[0].Address,
				Timestamp:        testing.ToTime("2025-09-25T07:55:57.306746166Z"),
				Signature:        testing.B64Dec("qtv1z4S2Q6T87vGQo0lrjRZqv9PrHIji4pTyviMnVyGx9td6eySdzwQwCthwmihU48ebNlFiMlFJ0CT891UmDg=="),
			},
			{
				BlockIDFlag:      tendermint.BlockIDFlagCommit,
				ValidatorAddress: valset.Validators[1].Address,
				Timestamp:        testing.ToTime("2025-09-25T07:55:57.310583641Z"),
				Signature:        testing.B64Dec("Q5E6Kjma00n/T98rC9qJmoB6JTGFX/IB+mDVs4Wd1h0eJ8fabY/6oI8zdoU6/7W6VR6wjpHyWBsJrpGT6C0LCg=="),
			},
		}

		msgHeader = testing.NewMsgHeader(
			chainID, newTimestamp, apphash, blockhash, parsethash, consensushash,
			newHeight, trustedHeight, valset, trustedValset, signatures,
		)
	)

	core.UpdateClient(cross, clientID, msgHeader)

	println("----------- assert render clients")
	println(core.Render("clients"))
	println("----------- assert render clients/07-tendermint-1/consensus_states")
	println(core.Render("clients/" + clientID + "/consensus_states"))
}

// Output:
// ----------- assert render clients
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - Status: Active
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: osmosis-1
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 1/44873552
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - Root: 0c01c13b795868816ef579a995386617b8a35c6ed8c0365ac0648328a08787b7
//   - NextValidatorsHash: f2a14f0ea269dc5c07f035c3be55d5d01fd1532736d38c722ad17e58b1d7182c
//
// ----------- assert render clients/07-tendermint-1/consensus_states
// ## Client 07-tendermint-1 consensus states
// - ConsensusState height 1/44873551:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - RootHash: 0c01c13b795868816ef579a995386617b8a35c6ed8c0365ac0648328a08787b7
//   - NextValidatorsHash: f2a14f0ea269dc5c07f035c3be55d5d01fd1532736d38c722ad17e58b1d7182c
// - ConsensusState height 1/44873552:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - RootHash: 0c01c13b795868816ef579a995386617b8a35c6ed8c0365ac0648328a08787b7
//   - NextValidatorsHash: f2a14f0ea269dc5c07f035c3be55d5d01fd1532736d38c722ad17e58b1d7182c

// Events:
// [
//   {
//     "type": "create_client",
//     "attrs": [
//       {
//         "key": "client_id",
//         "value": "07-tendermint-1"
//       },
//       {
//         "key": "client_type",
//         "value": "07-tendermint"
//       },
//       {
//         "key": "consensus_heights",
//         "value": "1/44873551"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   },
//   {
//     "type": "update_client",
//     "attrs": [
//       {
//         "key": "client_id",
//         "value": "07-tendermint-1"
//       },
//       {
//         "key": "client_type",
//         "value": "07-tendermint"
//       },
//       {
//         "key": "consensus_heights",
//         "value": "1/44873552"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   }
// ]
