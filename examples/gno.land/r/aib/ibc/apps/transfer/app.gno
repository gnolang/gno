package transfer

import (
	"gno.land/p/aib/ibc/app"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
	"gno.land/r/aib/ibc/core"
)

type App struct{}

var portID string

func init() {
	portID = core.RegisterApp(cross, &App{})
}

var _ app.IBCApp = &App{}

// Implements app.IBCApp
func (a *App) OnSendPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) error {
	// Enforce that the source and destination portIDs are the same and equal to
	// the transfer portID
	// Enforce that the source and destination clientIDs are also in the clientID
	// format that transfer expects: {clientid}-{sequence}
	// This is necessary for IBC v2 since the portIDs (and thus the
	// application-application connection) is not prenegotiated by the channel
	// handshake.
	// This restriction can be removed in a future where the trace hop on receive
	// commits to **both** the source and destination portIDs rather than just
	// the destination port
	if payload.SourcePort != portID || payload.DestinationPort != portID {
		return ufmt.Errorf("payload port ID is invalid: expected %s, got sourcePort: %s destPort: %s", portID, payload.SourcePort, payload.DestinationPort)
	}
	if !types.IsValidClientID(sourceClient) || !types.IsValidClientID(destinationClient) {
		return ufmt.Errorf("client IDs must be in valid format: {string}-{number}")
	}

	// TODO decode payload.Value
	// data, err := types.UnmarshalPacketData(payload.Value, payload.Version, payload.Encoding)
	// if err != nil {
	// return err
	// }

	// TODO check sender
	// sender := data.Sender
	// signer := runtime.OriginCaller()
	// if sender != signer {
	// return ufmt.Errorf("sender %s is different from signer %s", sender, signer)
	// }

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
	// TODO
	// if strings.Contains(data.Token.Denom.Base, "/") {
	//   return errorsmod.Wrapf(types.ErrInvalidDenomForTransfer, "base denomination %s cannot contain slashes for IBC v2 packet", data.Token.Denom.Base)
	// }
	//
	// if err := im.keeper.SendTransfer(ctx, payload.SourcePort, sourceChannel, data.Token, signer); err != nil {
	// return err
	// }
	//
	// events.EmitTransferEvent(ctx, data.Sender, data.Receiver, data.Token, data.Memo)

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
	panic("transfer.OnAcknowledgementPacket not implemented")
}
