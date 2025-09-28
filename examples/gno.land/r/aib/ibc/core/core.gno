package core

import (
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
)

const maxTimeoutDelta time.Duration = 24 * time.Hour

func SendPacket(cur realm, msg types.MsgSendPacket) uint64 /*sequence*/ {
	if err := msg.ValidateBasic(); err != nil {
		panic(err)
	}
	clientID := msg.SourceClient
	c := store.getClient(clientID)
	if c == nil {
		panic(ufmt.Sprintf("client %s not found", clientID))
	}
	if status := c.lightClient.Status(); status != lightclient.Active {
		panic(ufmt.Sprintf("client (%s) status is %s", clientID, status))
	}

	// Ensure counterparty has been registered
	if c.counterpartyClientID == "" {
		panic(ufmt.Sprintf("counterparty not found for client: %s", clientID))
	}

	// timeoutTimestamp must be greater than current block time and less or equal
	// to current block time + maxTimeoutDelta.
	var (
		timeout    = time.Unix(int64(msg.TimeoutTimestamp), 0)
		minTimeout = time.Now()
		maxTimeout = minTimeout.Add(maxTimeoutDelta)
	)
	if !timeout.After(minTimeout) {
		panic(ufmt.Sprintf(
			"timeout is less than or equal the current block timestamp, %d <= %d",
			msg.TimeoutTimestamp, minTimeout.Unix(),
		))
	}
	if timeout.After(maxTimeout) {
		panic(ufmt.Sprintf(
			"timeout is after the max allowed timeout, %d > %d",
			msg.TimeoutTimestamp, maxTimeout.Unix(),
		))
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
