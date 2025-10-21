package types

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"gno.land/p/aib/ibc/host"
	"gno.land/p/aib/ics23"
	"gno.land/p/nt/ufmt"
)

// MsgSendPacket sends an outgoing IBC packet.
type MsgSendPacket struct {
	SourceClient string
	// TimeoutTimestamp holds the seconds after which the packet times out.
	TimeoutTimestamp uint64
	Payloads         []Payload
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
	Packet          Packet
	ProofCommitment []byte
	ProofHeight     Height
}

// MsgAcknowledgement receives incoming IBC acknowledgement.
type MsgAcknowledgement struct {
	Packet          Packet
	Acknowledgement Acknowledgement
	ProofAcked      []ics23.CommitmentProof_Exist
	ProofHeight     Height
}

// ValidateBasic performs basic checks on a MsgAcknowledgement.
func (msg *MsgAcknowledgement) ValidateBasic() error {
	if len(msg.ProofAcked) == 0 {
		return errors.New("cannot submit an empty acknowledgement proof")
	}
	for i, proof := range msg.ProofAcked {
		if proof.Exist == nil {
			return ufmt.Errorf("invalid commitment proof #%d: proof.Exist is nil", i)
		}
	}

	if err := msg.Acknowledgement.Validate(); err != nil {
		return err
	}

	return msg.Packet.ValidateBasic()
}

func ErrorAcknowledgement() []byte {
	bz := sha256.Sum256([]byte("UNIVERSAL_ERROR_ACKNOWLEDGEMENT"))
	return bz[:]
}

// Validate performs a basic validation of the acknowledgement
func (ack Acknowledgement) Validate() error {
	// acknowledgement list should be non-empty
	if len(ack.AppAcknowledgements) == 0 {
		return errors.New("app acknowledgements must be non-empty")
	}

	for _, a := range ack.AppAcknowledgements {
		// Each app acknowledgement should be non-empty
		if len(a) == 0 {
			return errors.New("app acknowledgement cannot be empty")
		}

		// Ensure that the app acknowledgement contains ErrorAcknowledgement
		// **if and only if** the app acknowledgement list has a single element
		if len(ack.AppAcknowledgements) > 1 {
			if bytes.Equal(a, ErrorAcknowledgement()) {
				return errors.New("cannot have the error acknowledgement in multi acknowledgement list")
			}
		}
	}

	return nil
}

// MsgTimeout receives timed-out packet
type MsgTimeout struct {
	Packet          Packet
	ProofUnreceived []ics23.CommitmentProof_Nonexist
	ProofHeight     Height
}
