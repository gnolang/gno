package main

import (
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// UpdateClient success with height older than clientState.LatestHeight
func main() {
	// CreateClient and RegisterCounterparty
	var (
		chainID     = "atomone-1"
		height      = uint64(2)
		clientState = testing.NewClientState(chainID, types.NewHeight(1, height))
		apphash     = testing.GenHash("apphash-2")
		// priv=8a6cAbQSpDbebmcTEhCMPhhr/SkL/2pizo60yzHRkN9Uyk7RHOZm7g4xW+yeJh147/Z4/6HXF6gBwcFNkLsZ/A==
		val1 = tendermint.NewValidator("9DIBYr64rywKO3Kk6+743xDHcEU=",
			"VMpO0RzmZu4OMVvsniYdeO/2eP+h1xeoAcHBTZC7Gfw=", 2)
		// priv=nWg6ETc62tyxd94lh8fFaQnZKaAW6vlS0L/4lfseJuI14ZXUKp7AZROkflLFVF+SBg4wJVfzgzIKyWq3D066+g==
		val2 = tendermint.NewValidator("y+naL3ubs9q1bXrY9+uRxY9c+J8=",
			"NeGV1CqewGUTpH5SxVRfkgYOMCVX84MyCslqtw9Ouvo=", 3)
		valset         = tendermint.NewValset(val1, val2)
		consensusState = testing.GenConsensusState(time.Now(), apphash, valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "07-tendermint-2")

	// Update clientID with height=5
	apphash = testing.GenHash("apphash-5")
	var (
		newHeight     = uint64(5)
		newTimestamp  = consensusState.Timestamp.Add(time.Minute)
		blockhash     = testing.B64Dec("NpiImIJoaSaIucwNs5cqpgMsL/8wxEPYC3P0jA5aQSI=")
		parsethash    = testing.B64Dec("QqzwnLzvixIcUz+hPeUQjDV6NaLkFRKXACCxJIrBHzw=")
		consensushash = testing.B64Dec("JIS7qLmeUx4aP2QiNi59PuuZqDJRGZaKNV8FwJ7t1As=")
		trustedValset = tendermint.NewValset(val1, val2)
		trustedHeight = clientState.LatestHeight
		signatures    = []tendermint.CommitSig{
			{
				BlockIDFlag:      tendermint.BlockIDFlagCommit,
				ValidatorAddress: valset.Validators[0].Address,
				Timestamp:        testing.ToTime("2025-09-25T07:55:57.306746166Z"),
				Signature:        testing.B64Dec("dXO4Y8p5cKwubGLT0vkSVu4UpJRfrocMVeD02r9FP6w+WzOT2WIars/BJH1xvIM0x9MutOF6utFbqwi9YFQ+CA=="),
			},
			{
				BlockIDFlag:      tendermint.BlockIDFlagCommit,
				ValidatorAddress: valset.Validators[1].Address,
				Timestamp:        testing.ToTime("2025-09-25T07:55:57.306746166Z"),
				Signature:        testing.B64Dec("TwSjd4gWQExpna4o7WmfgjuhVr3gWD14dcxEez7kxt9Fd3eUChDOj/lGFNkllLJvX3lcSuHtsRK2KPsIfRaKBQ=="),
			},
		}

		msgHeader = tendermint.NewMsgHeader(
			chainID, newTimestamp, apphash, blockhash, parsethash, consensushash,
			newHeight, trustedHeight, valset, trustedValset, signatures,
		)
	)

	core.UpdateClient(cross, clientID, msgHeader)

	println("----------- assert render clients")
	println(core.Render("clients"))
	println("----------- assert render clients/07-tendermint-1/consensus_states")
	println(core.Render("clients/" + clientID + "/consensus_states"))

	// Second Update with height older than clientState.LatestHeight
	newHeight = 3
	newTimestamp = newTimestamp.Add(time.Minute)
	apphash = testing.GenHash("apphash-3")
	// update signatures since height have change
	signatures[0].Signature = testing.B64Dec("iY9QJIgtUz6AKym75cfycieAOLiO8Y7z12ndOYoEkCAXTba+1r4pt7lTuk+RyhOex9q8wye7kMN1DAF25wL1DQ==")
	signatures[1].Signature = testing.B64Dec("Dl1baBptDCLku0Ra4IpkA7u0dV1EW5Jh9Ev/iSLcGbIWfn/rABZ3JLGy5sMNu3Xt1aR0StSvzZ/1HQS8IUWbDg==")
	msgHeader = tendermint.NewMsgHeader(
		chainID, newTimestamp, apphash, blockhash, parsethash, consensushash,
		newHeight, trustedHeight, valset, trustedValset, signatures,
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
// - CounterpartyClientID: 07-tendermint-2
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: atomone-1
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 1/5
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - Root: 774bfe80005952b99fe0765099c1281ab1e7d4db80013b2a840fd7ff506f898a
//   - NextValidatorsHash: df96ae5c7a00ce0fcfbf2bd210602dc347c1756fd0d1e0bad607b585200df602
//
// ----------- assert render clients/07-tendermint-1/consensus_states
// ## Client 07-tendermint-1 consensus states
// - ConsensusState height 1/2:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - RootHash: fda37a1203142d75e6b9f70159edfb8280674ddc448c7e63c767cf770bfe9b14
//   - NextValidatorsHash: df96ae5c7a00ce0fcfbf2bd210602dc347c1756fd0d1e0bad607b585200df602
// - ConsensusState height 1/5:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - RootHash: 774bfe80005952b99fe0765099c1281ab1e7d4db80013b2a840fd7ff506f898a
//   - NextValidatorsHash: df96ae5c7a00ce0fcfbf2bd210602dc347c1756fd0d1e0bad607b585200df602
//
// ----------- assert render clients
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - Status: Active
// - CounterpartyClientID: 07-tendermint-2
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: atomone-1
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 1/5
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - Root: 774bfe80005952b99fe0765099c1281ab1e7d4db80013b2a840fd7ff506f898a
//   - NextValidatorsHash: df96ae5c7a00ce0fcfbf2bd210602dc347c1756fd0d1e0bad607b585200df602
//
// ----------- assert render clients/07-tendermint-1/consensus_states
// ## Client 07-tendermint-1 consensus states
// - ConsensusState height 1/2:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - RootHash: fda37a1203142d75e6b9f70159edfb8280674ddc448c7e63c767cf770bfe9b14
//   - NextValidatorsHash: df96ae5c7a00ce0fcfbf2bd210602dc347c1756fd0d1e0bad607b585200df602
// - ConsensusState height 1/3:
//   - Timestamp: 2009-02-13 23:33:30 +0000 UTC
//   - RootHash: 2c970bdc7aa83d24aa3f38e90537d3bf44d81b93c53902cbb0633ce654b58c0d
//   - NextValidatorsHash: df96ae5c7a00ce0fcfbf2bd210602dc347c1756fd0d1e0bad607b585200df602
// - ConsensusState height 1/5:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - RootHash: 774bfe80005952b99fe0765099c1281ab1e7d4db80013b2a840fd7ff506f898a
//   - NextValidatorsHash: df96ae5c7a00ce0fcfbf2bd210602dc347c1756fd0d1e0bad607b585200df602
