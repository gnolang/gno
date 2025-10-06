package core

import (
	"chain"
	"chain/banker"
	"encoding/hex"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/aib/limitedbanker"
	"gno.land/p/nt/ufmt"
)

const maxTimeoutDelta time.Duration = 24 * time.Hour

func SendPacket(cur realm, msg types.MsgSendPacket) (sequence uint64) {
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
		panic(ufmt.Sprintf("counterparty not found for client %s", clientID))
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

	// check if the latest consensus timestamp is lower than timeoutTimestamp
	latestTimestamp, err := c.lightClient.TimestampAtHeight(c.lightClient.LatestHeight())
	if err != nil {
		panic(err)
	}
	if latestTimestamp >= msg.TimeoutTimestamp {
		panic(ufmt.Errorf("latest timestamp: %d, timeout timestamp: %d",
			latestTimestamp, msg.TimeoutTimestamp))
	}

	// construct packet from given fields
	sequence = uint64(c.sendSeq.Next())
	packet := types.NewPacket(sequence, msg.SourceClient, c.counterpartyClientID, msg.TimeoutTimestamp, msg.Payloads...)
	if err := packet.ValidateBasic(); err != nil {
		panic(ufmt.Errorf("constructed packet failed basic validation: %v", err))
	}

	// set the packet commitment
	c.setPacketCommitment(packet)

	// emit events
	chain.Emit(types.EventTypeSendPacket,
		types.AttributeKeySrcClient, packet.SourceClient,
		types.AttributeKeyDstClient, packet.DestinationClient,
		types.AttributeKeySequence, ufmt.Sprintf("%d", packet.Sequence),
		types.AttributeKeyTimeoutTimestamp, ufmt.Sprintf("%d", packet.TimeoutTimestamp),
		types.AttributeKeyEncodedPacketHex, hex.EncodeToString(packet.ProtoMarshal()),
	)

	// Build a banker limited to spend OriginSend.
	banker := limitedbanker.New(
		banker.NewBanker(banker.BankerTypeRealmSend),
		banker.OriginSend(),
	)
	// Invoke registed app OnSendPacket() for each payload.SourcePort.
	for _, payload := range msg.Payloads {
		app := store.route(payload.SourcePort)
		err := app.OnSendPacket(
			banker, msg.SourceClient, c.counterpartyClientID, sequence, payload,
		)
		if err != nil {
			panic(ufmt.Sprintf(
				"send packet failed for app %s: %v", payload.SourcePort, err,
			))
		}
	}

	return sequence
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
