package app

import (
	"chain/banker"

	"gno.land/p/aib/ibc/types"
)

type IBCApp interface {
	OnSendPacket(
		banker banker.Banker,
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
	) error

	OnRecvPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
	) types.RecvPacketResult

	// OnTimeoutPacket is executed when a packet has timed out on the receiving chain.
	OnTimeoutPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
	) error

	// OnAcknowledgementPacket is executed when a packet gets acknowledged
	OnAcknowledgementPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		acknowledgement []byte,
		payload types.Payload,
	) error
}
