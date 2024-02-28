package types

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidMessagePayload = errors.New("invalid message payload")
	ErrInvalidMessageType    = errors.New("invalid message type")
)

// GetSignaturePayload returns the signature payload for the message
func (m *Message) GetSignaturePayload() ([]byte, error) {
	switch m.Type {
	case MessageType_PROPOSAL:
		payload := m.GetProposalMessage()
		if payload == nil {
			return nil, ErrInvalidMessagePayload
		}

		return payload.Marshal(), nil
	case MessageType_PREVOTE:
		payload := m.GetPrevoteMessage()
		if payload == nil {
			return nil, ErrInvalidMessagePayload
		}

		return payload.Marshal(), nil
	case MessageType_PRECOMMIT:
		payload := m.GetPrecommitMessage()
		if payload == nil {
			return nil, ErrInvalidMessagePayload
		}

		return payload.Marshal(), nil
	default:
		return nil, ErrInvalidMessageType
	}
}

// Marshal returns the marshalled message
func (m *ProposalMessage) Marshal() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&ProposalMessage{
		View:          m.View,
		From:          m.From,
		Proposal:      m.Proposal,
		ProposalRound: m.ProposalRound,
	})

	return raw
}

// IsValid validates that the given message is valid
func (m *ProposalMessage) IsValid() bool {
	// Make sure the view is present
	if m.View == nil {
		return false
	}

	// Make sure the sender is present
	if m.From == nil {
		return false
	}

	// Make sure the proposal is present
	if m.Proposal == nil {
		return false
	}

	// Make sure the proposal round is
	// for a good value
	return m.ProposalRound >= -1
}

// Marshal returns the marshalled message
func (m *PrevoteMessage) Marshal() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&PrevoteMessage{
		View:       m.View,
		From:       m.From,
		Identifier: m.Identifier,
	})

	return raw
}

// IsValid validates that the given message is valid
func (m *PrevoteMessage) IsValid() bool {
	// Make sure the view is present
	if m.View == nil {
		return false
	}

	// Make sure the sender is present
	return m.From != nil
}

// Marshal returns the marshalled message
func (m *PrecommitMessage) Marshal() []byte {
	//nolint:errcheck // No need to verify the error
	raw, _ := proto.Marshal(&PrecommitMessage{
		View:       m.View,
		From:       m.From,
		Identifier: m.Identifier,
	})

	return raw
}

// IsValid validates that the given message is valid
func (m *PrecommitMessage) IsValid() bool {
	// Make sure the view is present
	if m.View == nil {
		return false
	}

	// Make sure the sender is present
	return m.From != nil
}
