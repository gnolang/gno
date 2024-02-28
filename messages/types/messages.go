package types

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidMessageView          = errors.New("invalid message view")
	ErrInvalidMessageSender        = errors.New("invalid message sender")
	ErrInvalidMessageProposal      = errors.New("invalid message proposal")
	ErrInvalidMessageProposalRound = errors.New("invalid message proposal round")
)

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
