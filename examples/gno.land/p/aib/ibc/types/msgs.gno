package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"gno.land/p/aib/ibc/host"
	"gno.land/p/aib/ics23"
	"gno.land/p/nt/ufmt"
	"gno.land/p/onbloc/json"
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
			if bytes.Equal(a, UniversalErrorAcknowledgement()) {
				return errors.New("cannot have the error acknowledgement in multi acknowledgement list")
			}
		}
	}

	return nil
}

func UniversalErrorAcknowledgement() []byte {
	bz := sha256.Sum256([]byte("UNIVERSAL_ERROR_ACKNOWLEDGEMENT"))
	return bz[:]
}

// AppAcknowledgement is the recommended acknowledgement format to be used by
// app-specific protocols.
type AppAcknowledgement struct {
	Response AppAcknowledgementResponse
}

type AppAcknowledgementResponse struct {
	Result []byte
	Error  string
}

func NewErrorAppAcknowledgement() AppAcknowledgement {
	return AppAcknowledgement{Response: AppAcknowledgementResponse{Error: "error"}}
}

func (a AppAcknowledgement) Success() bool {
	return a.Response.Error == ""
}

func (a *AppAcknowledgement) UnmarshalJSON(bz []byte) error {
	root, err := json.Unmarshal(bz)
	if err != nil {
		return err
	}
	resp, err := root.GetKey("response")
	if err != nil {
		return err
	}
	result, err := resp.GetKey("result")
	if err != nil {
		rerror, err := resp.GetKey("error")
		if err != nil {
			return err
		}
		a.Response = AppAcknowledgementResponse{Error: rerror.MustString()}
		return nil
	}
	resBz, err := base64.StdEncoding.DecodeString(result.MustString())
	if err != nil {
		return err
	}
	a.Response = AppAcknowledgementResponse{Result: resBz}
	return nil
}

// MsgTimeout receives timed-out packet
type MsgTimeout struct {
	Packet          Packet
	ProofUnreceived []ics23.CommitmentProof_Nonexist
	ProofHeight     Height
}
