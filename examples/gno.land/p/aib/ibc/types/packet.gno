package types

// FIXME: does p/demo/json handles field annotations ?

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

// Packet defines a type that carries data across different chains through IBC
type Packet struct {
	// number corresponds to the order of sends and receives, where a Packet
	// with an earlier sequence number must be sent and received before a Packet
	// with a later sequence number.
	Sequence uint64 `json:"sequence,omitempty"`
	// identifies the sending client on the sending chain.
	SourceClient string `json:"source_client,omitempty"`
	// identifies the receiving client on the receiving chain.
	DestinationClient string `json:"destination_client,omitempty"`
	// timeout timestamp in seconds after which the packet times out.
	TimeoutTimestamp uint64 `json:"timeout_timestamp,omitempty"`
	// a list of payloads, each one for a specific application.
	Payloads []Payload `json:"payloads"`
}

type Payload struct {
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

// Acknowledgement contains a list of all ack results associated with a single packet.
// In the case of a successful receive, the acknowledgement will contain an app acknowledgement
// for each application that received a payload in the same order that the payloads were sent
// in the packet.
// If the receive is not successful, the acknowledgement will contain a single app acknowledgment
// which will be a constant error acknowledgment as defined by the IBC v2 protocol.
type Acknowledgement struct {
	AppAcknowledgements [][]byte `json:"app_acknowledgements,omitempty"`
}
