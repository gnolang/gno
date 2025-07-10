package core

import (
	"std"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/demo/ufmt"
)

type client struct {
	id                       string
	typ                      string
	creator                  std.Address
	counterpartyClientID     string
	counterpartyMerklePrefix [][]byte
	lightClient              lightclient.LightClientModule
}

// CreateClient generates a new client identifier and invokes the associated
// light client module in order to initialize the client.
//
// TODO: why I don't need to declare 'cur realm' for this func? Because it
// creates a new object and does not modify an existing one ?
func CreateClient(clientType string, clientState, consensusState any) string {
	c := store.addClient(clientType, std.OriginCaller())
	err := c.lightClient.Initialize(c.id, clientState, consensusState)
	if err != nil {
		panic(err)
	}
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
	if c.creator != std.OriginCaller() {
		panic(ufmt.Sprintf("expected same signer as CreateClient submittor %s, got %s", c.creator, std.OriginCaller()))
	}
	c.counterpartyClientID = counterpartyClientID
	c.counterpartyMerklePrefix = counterpartyMerklePrefix
}

// UpdateClient will update the given IBC v2 light client with a new header.
// Can also be used to submit a misbehavior (clientMessage can be a header or
// a misbehavior, maybe split into 2 functions would make more sense here).
func UpdateClient(clientID string, clientMessage []byte) {
	// TODO
}

// UpdateClient upgrades the client to a new client state if this new client
// was committed to by the old client at the specified upgrade height.
func UpgradeClient(clientID string, clientState, consensusState []byte,
	proofUpgradeClient, proofUpgradeConsensusState []byte) {
	// TODO
}

// RecoverClient recovers a frozen or expired client.
// NOTE(tb): This must be a permissioned function called by a gov proposal.
func RecoverClient(clientID string) {
	// TODO
}
