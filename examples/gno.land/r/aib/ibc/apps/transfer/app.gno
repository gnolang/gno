package transfer

import (
	"bytes"
	"chain"
	"chain/banker"
	"chain/runtime"
	"strings"

	"gno.land/p/aib/ibc/app"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
	"gno.land/r/aib/ibc/core"
)

type App struct{}

const (
	// NOTE we must use the same portID as the ibc-go transfer IBC app
	PortID = "transfer"
	// V1 defines first version of the IBC transfer module
	V1               = "ics20-1"
	EncodingProtobuf = "application/x-protobuf"

	denomPrefix          = "ibc"
	escrowAddressVersion = V1
)

func init() {
	// Register the app in the IBC router.
	core.RegisterApp(cross, PortID, &App{})
}

var _ app.IBCApp = &App{}

// Implements app.IBCApp
func (a *App) OnSendPacket(
	banker banker.Banker,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) error {
	// TODO add parameter to disable the app
	// TODO add parameter to block sender addr

	// Enforce that the source and destination portIDs are the same and equal to
	// the transfer portID.
	// Enforce that the source and destination clientIDs are also in the clientID
	// format that transfer expects: {clientid}-{sequence}.
	// This is necessary for IBC v2 since the portIDs (and thus the
	// application-application connection) is not prenegotiated by the channel
	// handshake.
	// This restriction can be removed in a future where the trace hop on receive
	// commits to **both** the source and destination portIDs rather than just
	// the destination port.
	if payload.SourcePort != PortID || payload.DestinationPort != PortID {
		return ufmt.Errorf("payload port ID is invalid: expected %s, got sourcePort: %s destPort: %s", PortID, payload.SourcePort, payload.DestinationPort)
	}
	if !types.IsValidClientID(sourceClient) || !types.IsValidClientID(destinationClient) {
		return ufmt.Errorf("client IDs must be in valid format: {string}-{number}")
	}
	if payload.Version != V1 {
		return ufmt.Errorf("invalid ICS20 version: expected %s, got %s", V1, payload.Version)
	}
	if payload.Encoding != EncodingProtobuf {
		return ufmt.Errorf("invalid encoding: expected %s, got %s", EncodingProtobuf, payload.Encoding)
	}

	var data FungibleTokenPacketData
	if err := data.ProtoUnmarshal(payload.Value); err != nil {
		return ufmt.Errorf("decoding FungibleTokenPacketData: %v", err)
	}
	if err := data.ValidateBasic(); err != nil {
		return ufmt.Errorf("invalid FungibleTokenPacketData: %v", err)
	}

	signer := runtime.OriginCaller().String()
	if data.Sender != signer {
		return ufmt.Errorf("invalid FungibleTokenPacketData: sender %s is different from signer %s", data.Sender, signer)
	}

	// TODO review comment: maybe not relevant in a pure IBCv2 impl.
	//
	// Enforce that the base denom does not contain any slashes
	// Since IBC v2 packets will no longer have channel identifiers, we cannot
	// rely on the channel format to easily divide the trace from the base
	// denomination in ICS20 v1 packets.
	// The simplest way to prevent any potential issues from arising is to simply
	// disallow any slashes in the base denomination.
	// This prevents such denominations from being sent with IBCV v2 packets,
	// however we can still support them in IBC v1 packets.
	// If we enforce that IBC v2 packets are sent with ICS20 v2 and above
	// versions that separate the trace from the base denomination in the packet
	// data, then we can remove this restriction.
	denom := ExtractDenomFromPath(data.Denom)
	if strings.Contains(denom.Base, "/") {
		return ufmt.Errorf("base denomination %s cannot contain slashes for IBC v2 packet", denom.Base)
	}
	token := Token{Denom: denom, Amount: data.Amount}

	if err := sendTransfer(banker, sourceClient, token); err != nil {
		return err
	}

	// Emit events
	chain.Emit(EventTypeTransfer,
		AttributeKeySender, data.Sender,
		AttributeKeyReceiver, data.Receiver,
		AttributeKeyDenom, token.Denom.Path(),
		AttributeKeyAmount, token.Amount,
		AttributeKeyMemo, data.Memo,
	)
	return nil
}

// Implements app.IBCApp
func (a *App) OnRecvPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) types.RecvPacketResult {
	panic("transfer.OnRecvPacket not implemented")
}

// Implements app.IBCApp
func (a *App) OnTimeoutPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) error {
	panic("transfer.OnTimeoutPacket not implemented")
}

// Implements app.IBCApp
func (a *App) OnAcknowledgementPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	acknowledgement []byte,
	payload types.Payload,
) error {
	var ack types.AppAcknowledgement
	// construct an error acknowledgement if the acknowledgement bytes are the sentinel error acknowledgement so we can use the shared transfer logic
	if bytes.Equal(acknowledgement, types.UniversalErrorAcknowledgement()) {
		// the specific error does not matter
		ack = types.NewErrorAppAcknowledgement()
	} else {
		if err := ack.UnmarshalJSON(acknowledgement); err != nil {
			return ufmt.Errorf("cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
		}
		if !ack.Success() {
			return ufmt.Errorf("cannot pass in a custom error acknowledgement with IBC v2")
		}
	}

	var data FungibleTokenPacketData
	if err := data.ProtoUnmarshal(payload.Value); err != nil {
		return ufmt.Errorf("decoding FungibleTokenPacketData: %v", err)
	}
	if err := data.ValidateBasic(); err != nil {
		return ufmt.Errorf("invalid FungibleTokenPacketData: %v", err)
	}
	denom := ExtractDenomFromPath(data.Denom)
	if strings.Contains(denom.Base, "/") {
		return ufmt.Errorf("base denomination %s cannot contain slashes for IBC v2 packet", denom.Base)
	}
	token := Token{Denom: denom, Amount: data.Amount}

	if ack.Success() {
		// the acknowledgement succeeded on the receiving chain so nothing
		// needs to be executed and no error needs to be returned
	} else {
		// TODO refund
		panic("ack refund not implemented")
	}

	// Emit events
	chain.Emit(EventTypePacket,
		AttributeKeySender, data.Sender,
		AttributeKeyReceiver, data.Receiver,
		AttributeKeyDenom, token.Denom.Path(),
		AttributeKeyAmount, token.Amount,
		AttributeKeyMemo, data.Memo,
		AttributeKeyAck, string(acknowledgement),
	)
	if ack.Success() {
		chain.Emit(EventTypePacket,
			AttributeKeyAckSuccess, string(ack.Response.Result),
		)
	} else {
		chain.Emit(EventTypePacket,
			AttributeKeyAckError, ack.Response.Error,
		)
	}
	return nil
}
