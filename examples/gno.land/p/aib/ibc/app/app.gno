package app

import (
	"chain/banker"

	"gno.land/p/aib/ibc/types"
)

// IBCApp defines an interface that implements all the callbacks that apps must
// define as specified in IBC Protocol V2.
type IBCApp interface {
	// OnSendPacket is executed when a packet is being sent from sending chain.
	// This callback is provided with the source and destination IDs, the signer,
	// the packet sequence and the packet data for this specific application.
	OnSendPacket(
		banker banker.Banker,
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
	) error

	// OnRecvPacket is executed when a packet is received from receiving chain.
	OnRecvPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
	) types.RecvPacketResult

	// OnTimeoutPacket is executed when a packet has timed out on the receiving
	// chain.
	OnTimeoutPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
	) error

	// OnAcknowledgementPacket is executed when a packet gets acknowledged.
	OnAcknowledgementPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		acknowledgement []byte,
		payload types.Payload,
	) error
}
