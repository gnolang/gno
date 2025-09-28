package core

import (
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
)

func SendPacket(cur realm, msg types.MsgSendPacket) uint64 /*sequence*/ {
	if err := msg.ValidateBasic(); err != nil {
		panic(err)
	}
	clientID := msg.SourceClient
	c := store.getClient(clientID)
	if c == nil {
		panic(ufmt.Sprintf("client %s not found", clientID))
	}
	// Ensure counterparty has been registered
	if c.counterpartyClientID == "" {
		panic(ufmt.Sprintf("counterparty not found for client: %s", clientID))
	}

	// TODO
	return 0
}

func RecvPacket(cur realm, msg types.MsgRecvPacket) types.ResponseResultType {
	// TODO
	return types.RESPONSE_FAILURE
}

func Acknowledgement(cur realm, msg types.MsgAcknowledgement) types.ResponseResultType {
	// TODO
	return types.RESPONSE_FAILURE
}

func Timeout(cur realm, msg types.MsgTimeout) types.ResponseResultType {
	// TODO
	return types.RESPONSE_FAILURE
}
