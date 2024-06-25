package types

import (
	"bytes"
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidMessageView          = errors.New("invalid message view")
	ErrInvalidMessageSender        = errors.New("invalid message sender")
	ErrInvalidMessageProposal      = errors.New("invalid message proposal")
	ErrInvalidMessageProposalRound = errors.New("invalid message proposal round")
)

func (v *View) Equals(view *View) bool {
	if v.GetHeight() != view.GetHeight() {
		return false
	}

	return v.GetRound() == view.GetRound()
}

// GetSignaturePayload returns the sign payload for the proposal message
func (m *ProposalMessage) GetSignaturePayload() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&ProposalMessage{
		View:          m.View,
		Sender:        m.Sender,
		Proposal:      m.Proposal,
		ProposalRound: m.ProposalRound,
	})

	return raw
}

// Marshal returns the marshalled message
func (m *ProposalMessage) Marshal() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&ProposalMessage{
		View:          m.View,
		Sender:        m.Sender,
		Signature:     m.Signature,
		Proposal:      m.Proposal,
		ProposalRound: m.ProposalRound,
	})

	return raw
}

// Verify validates that the given message is valid
func (m *ProposalMessage) Verify() error {
	// Make sure the view is present
	if m.View == nil {
		return ErrInvalidMessageView
	}

	// Make sure the sender is present
	if m.Sender == nil {
		return ErrInvalidMessageSender
	}

	// Make sure the proposal is present
	if m.Proposal == nil {
		return ErrInvalidMessageProposal
	}

	// Make sure the proposal round is
	// for a good round value
	if m.ProposalRound < -1 {
		return ErrInvalidMessageProposalRound
	}

	return nil
}

func (m *ProposalMessage) Equals(message *ProposalMessage) bool {
	if !m.GetView().Equals(message.GetView()) {
		return false
	}

	if !bytes.Equal(m.GetSender(), message.GetSender()) {
		return false
	}

	if !bytes.Equal(m.GetSignature(), message.GetSignature()) {
		return false
	}

	if !bytes.Equal(m.GetProposal(), message.GetProposal()) {
		return false
	}

	return m.GetProposalRound() == message.GetProposalRound()
}

// GetSignaturePayload returns the sign payload for the proposal message
func (m *PrevoteMessage) GetSignaturePayload() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&PrevoteMessage{
		View:       m.View,
		Sender:     m.Sender,
		Identifier: m.Identifier,
	})

	return raw
}

// Marshal returns the marshalled message
func (m *PrevoteMessage) Marshal() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&PrevoteMessage{
		View:       m.View,
		Sender:     m.Sender,
		Signature:  m.Signature,
		Identifier: m.Identifier,
	})

	return raw
}

// Verify validates that the given message is valid
func (m *PrevoteMessage) Verify() error {
	// Make sure the view is present
	if m.View == nil {
		return ErrInvalidMessageView
	}

	// Make sure the sender is present
	if m.Sender == nil {
		return ErrInvalidMessageSender
	}

	return nil
}

func (m *PrevoteMessage) Equals(message *PrevoteMessage) bool {
	if !m.GetView().Equals(message.GetView()) {
		return false
	}

	if !bytes.Equal(m.GetSender(), message.GetSender()) {
		return false
	}

	if !bytes.Equal(m.GetSignature(), message.GetSignature()) {
		return false
	}

	return bytes.Equal(m.GetIdentifier(), message.GetIdentifier())
}

// GetSignaturePayload returns the sign payload for the proposal message
func (m *PrecommitMessage) GetSignaturePayload() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&PrecommitMessage{
		View:       m.View,
		Sender:     m.Sender,
		Identifier: m.Identifier,
	})

	return raw
}

// Marshal returns the marshalled message
func (m *PrecommitMessage) Marshal() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&PrecommitMessage{
		View:       m.View,
		Sender:     m.Sender,
		Signature:  m.Signature,
		Identifier: m.Identifier,
	})

	return raw
}

// Verify validates that the given message is valid
func (m *PrecommitMessage) Verify() error {
	// Make sure the view is present
	if m.View == nil {
		return ErrInvalidMessageView
	}

	// Make sure the sender is present
	if m.Sender == nil {
		return ErrInvalidMessageSender
	}

	return nil
}

func (m *PrecommitMessage) Equals(message *PrecommitMessage) bool {
	if !m.GetView().Equals(message.GetView()) {
		return false
	}

	if !bytes.Equal(m.GetSender(), message.GetSender()) {
		return false
	}

	if !bytes.Equal(m.GetSignature(), message.GetSignature()) {
		return false
	}

	return bytes.Equal(m.GetIdentifier(), message.GetIdentifier())
}
