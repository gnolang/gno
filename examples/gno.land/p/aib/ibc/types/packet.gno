package types

import (
	"strings"

	"gno.land/p/aib/encoding/proto"
	"gno.land/p/aib/ibc/host"
	"gno.land/p/nt/ufmt"
)

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

// NewPacket constructs a new packet.
func NewPacket(sequence uint64, sourceClient, destinationClient string, timeoutTimestamp uint64, payloads ...Payload) Packet {
	return Packet{
		Sequence:          sequence,
		SourceClient:      sourceClient,
		DestinationClient: destinationClient,
		TimeoutTimestamp:  timeoutTimestamp,
		Payloads:          payloads,
	}
}

const MaximumPayloadsSize = 262144 // 256 KiB. This is the maximum size of all payloads combined

// ValidateBasic validates that a Packet satisfies the basic requirements.
func (p Packet) ValidateBasic() error {
	if len(p.Payloads) == 0 {
		return ufmt.Errorf("payload length must be greater than 0")
	}

	totalPayloadsSize := 0
	for i, pd := range p.Payloads {
		if err := pd.ValidateBasic(); err != nil {
			return ufmt.Errorf("invalid Payload #%d: %v", i, err)
		}
		totalPayloadsSize += len(pd.Value)
	}

	if totalPayloadsSize > MaximumPayloadsSize {
		return ufmt.Errorf("packet data bytes cannot exceed %d bytes", MaximumPayloadsSize)
	}

	if err := host.ChannelIdentifierValidator(p.SourceClient); err != nil {
		return ufmt.Errorf("invalid source ID: %v", err)
	}
	if err := host.ChannelIdentifierValidator(p.DestinationClient); err != nil {
		return ufmt.Errorf("invalid destination ID: %v", err)
	}

	if p.Sequence == 0 {
		return ufmt.Errorf("packet sequence cannot be 0")
	}
	if p.TimeoutTimestamp == 0 {
		return ufmt.Errorf("packet timeout timestamp cannot be 0")
	}

	return nil
}

// ProtoMarshal returns the protobuf encoding of a Packet.
//
//	message Packet {
//	  uint64 sequence = 1;
//	  string source_client = 2;
//	  string destination_client = 3;
//	  uint64 timeout_timestamp = 4;
//	  repeated Payload payloads = 5 [(gogoproto.nullable) = false];
//	}
func (p Packet) ProtoMarshal() []byte {
	var buf []byte

	// Field 1: sequence (varint)
	buf = proto.AppendVarint(buf, 1, uint64(p.Sequence))

	// Field 2: source_client (length-delimited)
	buf = proto.AppendLengthDelimited(buf, 2, []byte(p.SourceClient))

	// Field 3: destination_client (length-delimited)
	buf = proto.AppendLengthDelimited(buf, 3, []byte(p.DestinationClient))

	// Field 4: timeout_timestamp (varint)
	buf = proto.AppendVarint(buf, 4, p.TimeoutTimestamp)

	// Field 5: payloads
	for _, payload := range p.Payloads {
		bz := payload.ProtoMarshal()
		buf = proto.AppendLengthDelimited(buf, 5, bz)
	}
	return buf
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

// NewPayload constructs a new Payload
func NewPayload(sourcePort, destPort, version, encoding string, value []byte) Payload {
	return Payload{
		SourcePort:      sourcePort,
		DestinationPort: destPort,
		Version:         version,
		Encoding:        encoding,
		Value:           value,
	}
}

// ValidateBasic validates a Payload.
func (p Payload) ValidateBasic() error {
	if err := host.PortIdentifierValidator(p.SourcePort); err != nil {
		return ufmt.Errorf("invalid source port: %v", err)
	}
	if err := host.PortIdentifierValidator(p.DestinationPort); err != nil {
		return ufmt.Errorf("invalid destination port: %v", err)
	}
	if strings.TrimSpace(p.Version) == "" {
		return ufmt.Errorf("payload version cannot be empty")
	}
	if strings.TrimSpace(p.Encoding) == "" {
		return ufmt.Errorf("payload encoding cannot be empty")
	}
	if len(p.Value) == 0 {
		return ufmt.Errorf("payload value cannot be empty")
	}
	return nil
}

// ProtoMarshal returns the protobuf encoding of a Payload.
//
//	message Payload {
//	  string source_port = 1;
//	  string destination_port = 2;
//	  string version = 3;
//	  string encoding = 4;
//	  bytes value = 5;
//	}
func (p Payload) ProtoMarshal() []byte {
	var buf []byte

	// Field 1: source_port (length-delimited)
	buf = proto.AppendLengthDelimited(buf, 1, []byte(p.SourcePort))

	// Field 2: destination_port (length-delimited)
	buf = proto.AppendLengthDelimited(buf, 2, []byte(p.DestinationPort))

	// Field 3: version (length-delimited)
	buf = proto.AppendLengthDelimited(buf, 3, []byte(p.Version))

	// Field 4: encoding (length-delimited)
	buf = proto.AppendLengthDelimited(buf, 4, []byte(p.Encoding))

	// Field 5: value (length-delimited)
	buf = proto.AppendLengthDelimited(buf, 5, p.Value)

	return buf
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
