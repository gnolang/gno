package app

import "gno.land/p/aib/ibc/types"

type IBCApp interface {
	OnSendPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
		signer []byte,
	) error

	OnRecvPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
		relayer []byte,
	) types.RecvPacketResult

	// OnTimeoutPacket is executed when a packet has timed out on the receiving chain.
	OnTimeoutPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload types.Payload,
		relayer []byte,
	) error

	// OnAcknowledgementPacket is executed when a packet gets acknowledged
	OnAcknowledgementPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		acknowledgement []byte,
		payload types.Payload,
		relayer []byte,
	) error
}
