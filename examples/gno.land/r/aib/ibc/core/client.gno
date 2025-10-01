package core

import (
	"chain"
	"chain/runtime"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
)

// CreateClient generates a new client identifier and invokes the associated
// light client module in order to initialize the client.
func CreateClient(cur realm, clientState lightclient.ClientState, consensusState lightclient.ConsensusState) string {
	if clientState.ClientType() != consensusState.ClientType() {
		panic("client type for client state and consensus state do not match")
	}
	if err := clientState.ValidateBasic(); err != nil {
		panic(err)
	}
	if err := consensusState.ValidateBasic(); err != nil {
		panic(err)
	}
	c := store.addClient(clientState.ClientType(), runtime.OriginCaller())
	err := c.lightClient.Initialize(clientState, consensusState)
	if err != nil {
		panic(err)
	}
	if status := c.lightClient.Status(); status != lightclient.Active {
		panic(ufmt.Sprintf("cannot create client (%s) with status %s", c.id, status))
	}
	// Emit create client event
	chain.Emit(types.EventTypeCreateClient,
		types.AttributeKeyClientID, c.id,
		types.AttributeKeyClientType, c.typ,
		types.AttributeKeyConsensusHeights, c.lightClient.LatestHeight().String(),
	)
	return c.id
}

// RegisterCounterparty will register the IBC v2 counterparty info for the
// given clientID. It must be called by the same relayer that called
// CreateClient.
func RegisterCounterparty(cur realm, clientID string, counterpartyMerklePrefix [][]byte, counterpartyClientID string) {
	c := store.getClient(clientID)
	if c == nil {
		panic(ufmt.Sprintf("client %s not found", clientID))
	}
	if c.creator != runtime.OriginCaller() {
		panic(ufmt.Sprintf("expected same signer as CreateClient submittor %s, got %s", c.creator, runtime.OriginCaller()))
	}
	if c.counterpartyClientID != "" {
		panic("cannot register counterparty once it is already set")
	}
	c.counterpartyClientID = counterpartyClientID
	c.counterpartyMerklePrefix = counterpartyMerklePrefix
}

// UpdateClient will update the given IBC v2 light client with a new header.
// Can also be used to submit a misbehavior (clientMessage can be a header or
// a misbehavior, maybe split into 2 functions would make more sense here).
func UpdateClient(cur realm, clientID string, clientMessage lightclient.ClientMessage) {
	if err := clientMessage.ValidateBasic(); err != nil {
		panic(err)
	}
	c := store.getClient(clientID)
	if c == nil {
		panic(ufmt.Sprintf("client %s not found", clientID))
	}
	if c.typ != clientMessage.ClientType() {
		panic("client type for client state and client message do not match")
	}
	if status := c.lightClient.Status(); status != lightclient.Active {
		panic(ufmt.Sprintf("cannot update client (%s) with status %s", c.id, status))
	}
	// TODO check if relayer address is in whitelist
	// Similarly to ibc-v2 go implementation, for safety we're going to restrict
	// the UpdateClient call to a set of predefined relayers addresses.

	err := c.lightClient.VerifyClientMessage(clientMessage)
	if err != nil {
		panic(err)
	}

	// TODO check for misbehavior

	consensusHeights := c.lightClient.UpdateState(clientMessage)

	// Emit update client event
	var consensusHeightsStr string
	for i, h := range consensusHeights {
		consensusHeightsStr += h.String()
		if i < len(consensusHeights)-1 {
			consensusHeightsStr += ", "
		}
	}
	chain.Emit(types.EventTypeUpdateClient,
		types.AttributeKeyClientID, c.id,
		types.AttributeKeyClientType, c.typ,
		types.AttributeKeyConsensusHeights, consensusHeightsStr,
	)
}

// UpdateClient upgrades the client to a new client state if this new client
// was committed to by the old client at the specified upgrade height.
func UpgradeClient(cur realm, clientID string, clientState, consensusState, proofUpgradeClient, proofUpgradeConsensusState any) {
	// TODO
}

// RecoverClient recovers a frozen or expired client.
// NOTE(tb): This must be a permissioned function called by a gov proposal.
func RecoverClient(cur realm, clientID string) {
	// TODO
}
