package types

import (
	"google.golang.org/protobuf/proto"
)

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
	// TODO implement
	return true
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
	// TODO implement
	return true
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
	// TODO implement
	return true
}
