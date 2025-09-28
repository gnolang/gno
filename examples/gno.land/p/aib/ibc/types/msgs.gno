package types

import (
	"gno.land/p/aib/ibc/host"
	"gno.land/p/nt/ufmt"
)

// MsgSendPacket sends an outgoing IBC packet.
type MsgSendPacket struct {
	SourceClient     string    `json:"source_client,omitempty"`
	TimeoutTimestamp uint64    `json:"timeout_timestamp,omitempty"`
	Payloads         []Payload `json:"payloads"`
}

// NewMsgSendPacket creates a new MsgSendPacket instance.
func NewMsgSendPacket(sourceClient string, timeoutTimestamp uint64, payloads ...Payload) *MsgSendPacket {
	return &MsgSendPacket{
		SourceClient:     sourceClient,
		TimeoutTimestamp: timeoutTimestamp,
		Payloads:         payloads,
	}
}

// ValidateBasic performs basic checks on a MsgSendPacket.
func (msg *MsgSendPacket) ValidateBasic() error {
	if err := host.ClientIdentifierValidator(msg.SourceClient); err != nil {
		return err
	}

	if msg.TimeoutTimestamp == 0 {
		return ufmt.Errorf("timeout must not be 0")
	}

	if len(msg.Payloads) == 0 {
		return ufmt.Errorf("payload length must be greater than 0")
	}

	for i, pd := range msg.Payloads {
		if err := pd.ValidateBasic(); err != nil {
			return ufmt.Errorf("invalid payload #%d: %v", i, err)
		}
	}
	return nil
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
