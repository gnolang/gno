package types

// MsgSendPacket sends an outgoing IBC packet.
type MsgSendPacket struct {
	SourceClient     string    `json:"source_client,omitempty"`
	TimeoutTimestamp uint64    `json:"timeout_timestamp,omitempty"`
	Payloads         []Payload `json:"payloads"`
}

// MsgRecvPacket receives an incoming IBC packet.
type MsgRecvPacket struct {
	Packet          Packet `json:"packet"`
	ProofCommitment []byte `json:"proof_commitment,omitempty"`
	ProofHeight     Height `json:"proof_height"`
}

// MsgAcknowledgement receives incoming IBC acknowledgement.
type MsgAcknowledgement struct {
	Packet          Packet          `json:"packet"`
	Acknowledgement Acknowledgement `json:"acknowledgement"`
	ProofAcked      []byte          `json:"proof_acked,omitempty"`
	ProofHeight     Height          `json:"proof_height"`
}

// MsgTimeout receives timed-out packet
type MsgTimeout struct {
	Packet          Packet `json:"packet"`
	ProofUnreceived []byte `json:"proof_unreceived,omitempty"`
	ProofHeight     Height `json:"proof_height"`
}
