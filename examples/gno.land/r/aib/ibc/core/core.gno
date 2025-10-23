package core

import (
	"bytes"
	"chain"
	"chain/banker"
	"encoding/hex"
	"time"

	"gno.land/p/aib/ibc/host"
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
	// NOTE: ideally we should use banker.NewBanker(banker.BankerTypeOriginSend)
	// but this kind of banker "can only be instantiated by the origin package".
	// Does that mean we should pass such banker in the realm function parameters
	// directly ?
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
	// TODO check if relayer address is in whitelist
	return types.RESPONSE_FAILURE
}

func Acknowledgement(cur realm, msg types.MsgAcknowledgement) types.ResponseResultType {
	if err := msg.ValidateBasic(); err != nil {
		panic(err)
	}
	// TODO check if relayer address is in whitelist

	clientID := msg.Packet.SourceClient
	c := store.getClient(clientID)
	if c == nil {
		panic(ufmt.Sprintf("client %s not found", clientID))
	}
	if status := c.lightClient.Status(); status != lightclient.Active {
		panic(ufmt.Sprintf("client (%s) status is %s", clientID, status))
	}
	// lookup counterparty from packet identifiers
	// note this will be either the client identifier for IBC V2 paths
	// or an aliased channel identifier for IBC V1 paths
	if c.counterpartyClientID != msg.Packet.DestinationClient {
		panic(ufmt.Sprintf(
			"counterparty id (%s) does not match packet destination id (%s)",
			c.counterpartyClientID, msg.Packet.DestinationClient,
		))
	}

	commitment := c.getPacketCommitment(msg.Packet.Sequence)
	if len(commitment) == 0 {
		// This error indicates that the acknowledgement has already been relayed
		// or there is a misconfigured relayer attempting to prove an
		// acknowledgement for a packet never sent. Core IBC will treat this error
		// as a no-op in order to prevent an entire relay transaction from failing
		// and consuming unnecessary fees.
		return types.RESPONSE_NOOP
	}

	packetCommitment := types.CommitPacket(msg.Packet)

	// verify we sent the packet and haven't cleared it out yet
	if !bytes.Equal(commitment, packetCommitment) {
		h1, h2 := hex.EncodeToString(packetCommitment), hex.EncodeToString(commitment)
		panic(ufmt.Sprintf(
			"commitment bytes are not equal: got (%v), expected (%v)", h1, h2,
		))
	}

	// Verify existence of the acknowledgement commitment bytes in the proofs.
	var (
		key        = host.PacketAcknowledgementKey(msg.Packet.DestinationClient, msg.Packet.Sequence)
		merklePath = types.BuildMerklePath(c.counterpartyMerklePrefix, key)
		value      = types.CommitAcknowledgement(msg.Acknowledgement)
	)
	if err := c.lightClient.VerifyMembership(
		msg.ProofHeight, msg.ProofAcked, merklePath, value,
	); err != nil {
		panic(ufmt.Sprintf(
			"failed packet acknowledgement verification for client (%s): %v",
			clientID, err,
		))
	}

	c.deletePacketCommitment(msg.Packet.Sequence)

	chain.Emit(types.EventTypeAcknowledgePacket,
		types.AttributeKeySrcClient, msg.Packet.SourceClient,
		types.AttributeKeyDstClient, msg.Packet.DestinationClient,
		types.AttributeKeySequence, ufmt.Sprintf("%d", msg.Packet.Sequence),
		types.AttributeKeyTimeoutTimestamp, ufmt.Sprintf("%d", msg.Packet.TimeoutTimestamp),
		types.AttributeKeyEncodedPacketHex, hex.EncodeToString(msg.Packet.ProtoMarshal()),
	)

	// Invoke registed app OnAcknowledgementPacket() for each payload.SourcePort.
	recvSuccess := !bytes.Equal(msg.Acknowledgement.AppAcknowledgements[0],
		types.UniversalErrorAcknowledgement())
	for i, payload := range msg.Packet.Payloads {
		app := store.route(payload.SourcePort)
		// if recv was successful, each payload should have its own acknowledgement
		// so we send each individual acknowledgment to the application otherwise,
		// the acknowledgement only contains the sentinel error acknowledgement
		// which we send to the application. The application is responsible for
		// knowing that this is an error acknowledgement and executing the
		// appropriate logic.
		var ack []byte
		if recvSuccess {
			ack = msg.Acknowledgement.AppAcknowledgements[i]
		} else {
			ack = types.UniversalErrorAcknowledgement()
		}
		err := app.OnAcknowledgementPacket(
			msg.Packet.SourceClient, c.counterpartyClientID, msg.Packet.Sequence,
			ack, payload,
		)
		if err != nil {
			panic(ufmt.Sprintf(
				"acknowledgement packet failed for app %s: %v", payload.SourcePort, err,
			))
		}
	}

	return types.RESPONSE_SUCCESS
}

func Timeout(cur realm, msg types.MsgTimeout) types.ResponseResultType {
	// TODO
	// TODO check if relayer address is in whitelist
	return types.RESPONSE_FAILURE
}
