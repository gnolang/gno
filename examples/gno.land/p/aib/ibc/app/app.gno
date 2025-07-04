package app

type IBCApp interface {
	OnSendPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload Payload,
		signer []byte,
	) error

	OnRecvPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload Payload,
		relayer []byte,
	) RecvPacketResult

	// OnTimeoutPacket is executed when a packet has timed out on the receiving chain.
	OnTimeoutPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		payload Payload,
		relayer []byte,
	) error

	// OnAcknowledgementPacket is executed when a packet gets acknowledged
	OnAcknowledgementPacket(
		sourceClient string,
		destinationClient string,
		sequence uint64,
		acknowledgement []byte,
		payload Payload,
		relayer []byte,
	) error
}

// TODO moves these types elsewhere

type Payload struct {
	// TODO: does p/demo/json handles field annotations ?
	// specifies the source port of the packet.
	SourcePort string `json:"source_port,omitempty"`
	// specifies the destination port of the packet.
	DestinationPort string `json:"destination_port,omitempty"`
	// version of the specified application.
	Version string `json:"version,omitempty"`
	// the encoding used for the provided value.
	Encoding string `json:"encoding,omitempty"`
	// the raw bytes for the payload.
	Value []byte `json:"value,omitempty"`
}

// RecvPacketResult speecifies the status of a packet as well as the acknowledgement bytes.
type RecvPacketResult struct {
	// status of the packet
	Status PacketStatus `json:"status,omitempty"`
	// acknowledgement of the packet
	Acknowledgement []byte `json:"acknowledgement,omitempty"`
}

// PacketStatus specifies the status of a RecvPacketResult.
type PacketStatus int32

const (
	// PACKET_STATUS_UNSPECIFIED indicates an unknown packet status.
	PacketStatus_NONE PacketStatus = 0
	// PACKET_STATUS_SUCCESS indicates a successful packet receipt.
	PacketStatus_Success PacketStatus = 1
	// PACKET_STATUS_FAILURE indicates a failed packet receipt.
	PacketStatus_Failure PacketStatus = 2
	// PACKET_STATUS_ASYNC indicates an async packet receipt.
	PacketStatus_Async PacketStatus = 3
)
