package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// CreateClient success
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = testing.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)

	clientID1 := core.CreateClient(cross, clientState, consensusState)
	println("CreateClient", clientID1)
	clientID2 := core.CreateClient(cross, clientState, consensusState)
	println("CreateClient", clientID2)

	// RegisterCounterparty on clientID1
	core.RegisterCounterparty(cross, clientID1, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "07-tendermint-2")

	println("\n----------- assert render clients")
	println(core.Render("clients"))
	println("\n----------- assert render clients/xxx")
	println(core.Render("clients/xxx"))
	println("\n----------- assert render clients/07-tendermint-1")
	println(core.Render("clients/" + clientID1))
	println("\n----------- assert render clients/07-tendermint-1/consensus_states")
	println(core.Render("clients/" + clientID1 + "/consensus_states"))
	println("\n----------- assert render clients/07-tendermint-1/consensus_states/2/2")
	println(core.Render("clients/" + clientID1 + "/consensus_states/2/2"))
	println("\n----------- assert render clients/07-tendermint-1/consensus_states/2/3")
	println(core.Render("clients/" + clientID1 + "/consensus_states/2/3"))
	println("\n----------- assert render clients/07-tendermint-1/consensus_states/xxx/yyy")
	println(core.Render("clients/" + clientID1 + "/consensus_states/xxx/yyy"))
	println("\n----------- assert render clients/xxx/consensus_states/2/3")
	println(core.Render("clients/xxx/consensus_states/2/3"))
}

// Output:
// CreateClient 07-tendermint-1
// CreateClient 07-tendermint-2
//
// ----------- assert render clients
// [{"id":"07-tendermint-1","type":"07-tendermint","creator":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","status":"Active","counterparty_client_id":"07-tendermint-2","counterparty_merke_prefix":["prefix1","prefix2"],"client_state":{"chain_id":"chain-id-2","latest_height":{"revision_number":2,"revision_height":2},"trusting_period":"1h0m0s","unbonding_period":"3h0m0s"},"last_consensus_state":{"height":{"revision_number":2,"revision_height":2},"timestamp":1234567890,"root":"EGkBgr/gwNYq9XkPLS+UWQksRtpw/yCAIRjy3KjUlI8=","next_validators_hash":"TiAQkvoom97x5Cu6r9TvoalcmOXH6Bk+JLVRHGtEAjw="}},{"id":"07-tendermint-2","type":"07-tendermint","creator":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","status":"Active","counterparty_client_id":"","counterparty_merke_prefix":[],"client_state":{"chain_id":"chain-id-2","latest_height":{"revision_number":2,"revision_height":2},"trusting_period":"1h0m0s","unbonding_period":"3h0m0s"},"last_consensus_state":{"height":{"revision_number":2,"revision_height":2},"timestamp":1234567890,"root":"EGkBgr/gwNYq9XkPLS+UWQksRtpw/yCAIRjy3KjUlI8=","next_validators_hash":"TiAQkvoom97x5Cu6r9TvoalcmOXH6Bk+JLVRHGtEAjw="}}]
//
// ----------- assert render clients/xxx
// {"error":"client xxx not found"}
//
// ----------- assert render clients/07-tendermint-1
// {"id":"07-tendermint-1","type":"07-tendermint","creator":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","status":"Active","counterparty_client_id":"07-tendermint-2","counterparty_merke_prefix":["prefix1","prefix2"],"client_state":{"chain_id":"chain-id-2","latest_height":{"revision_number":2,"revision_height":2},"trusting_period":"1h0m0s","unbonding_period":"3h0m0s"},"last_consensus_state":{"height":{"revision_number":2,"revision_height":2},"timestamp":1234567890,"root":"EGkBgr/gwNYq9XkPLS+UWQksRtpw/yCAIRjy3KjUlI8=","next_validators_hash":"TiAQkvoom97x5Cu6r9TvoalcmOXH6Bk+JLVRHGtEAjw="}}
//
// ----------- assert render clients/07-tendermint-1/consensus_states
// [{"height":{"revision_number":2,"revision_height":2},"timestamp":1234567890,"root":"EGkBgr/gwNYq9XkPLS+UWQksRtpw/yCAIRjy3KjUlI8=","next_validators_hash":"TiAQkvoom97x5Cu6r9TvoalcmOXH6Bk+JLVRHGtEAjw="}]
//
// ----------- assert render clients/07-tendermint-1/consensus_states/2/2
// {"height":{"revision_number":2,"revision_height":2},"timestamp":1234567890,"root":"EGkBgr/gwNYq9XkPLS+UWQksRtpw/yCAIRjy3KjUlI8=","next_validators_hash":"TiAQkvoom97x5Cu6r9TvoalcmOXH6Bk+JLVRHGtEAjw="}
//
// ----------- assert render clients/07-tendermint-1/consensus_states/2/3
// {"error":"consensus state not found for height 2/3"}
//
// ----------- assert render clients/07-tendermint-1/consensus_states/xxx/yyy
// {"error":"cant parse height xxx/yyy: invalid revision number. parse err: strconv.ParseUint: parsing \"xxx\": invalid syntax"}
//
// ----------- assert render clients/xxx/consensus_states/2/3
// {"error":"client xxx not found"}

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
//         "value": "2/2"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   },
//   {
//     "type": "create_client",
//     "attrs": [
//       {
//         "key": "client_id",
//         "value": "07-tendermint-2"
//       },
//       {
//         "key": "client_type",
//         "value": "07-tendermint"
//       },
//       {
//         "key": "consensus_heights",
//         "value": "2/2"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   }
// ]
