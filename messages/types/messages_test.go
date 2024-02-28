package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestProposalMessage_GetSignaturePayload(t *testing.T) {
	t.Parallel()

	// Create the proposal message
	m := &ProposalMessage{
		View: &View{
			Height: 10,
			Round:  10,
		},
		Sender:        []byte("sender"),
		Signature:     []byte("signature"),
		Proposal:      []byte("proposal"),
		ProposalRound: 0,
	}

	// Get the signature payload
	payload := m.GetSignaturePayload()

	var raw ProposalMessage

	require.NoError(t, proto.Unmarshal(payload, &raw))

	// Make sure the signature was not marshalled
	assert.Nil(t, raw.Signature)

	// Make sure other fields are intact
	assert.Equal(t, m.GetView().GetHeight(), raw.GetView().GetHeight())
	assert.Equal(t, m.GetView().GetRound(), raw.GetView().GetRound())
	assert.Equal(t, m.GetSender(), raw.GetSender())
	assert.Equal(t, m.GetProposal(), raw.GetProposal())
	assert.Equal(t, m.GetProposalRound(), raw.GetProposalRound())
}

func TestProposalMessage_Marshal(t *testing.T) {
	t.Parallel()

	// Create the proposal message
	m := &ProposalMessage{
		View: &View{
			Height: 10,
			Round:  10,
		},
		Sender:        []byte("sender"),
		Signature:     []byte("signature"),
		Proposal:      []byte("proposal"),
		ProposalRound: 0,
	}

	// Marshal the message
	marshalled := m.Marshal()

	var raw ProposalMessage

	require.NoError(t, proto.Unmarshal(marshalled, &raw))

	// Make sure other fields are intact
	assert.Equal(t, m.GetView().GetHeight(), raw.GetView().GetHeight())
	assert.Equal(t, m.GetView().GetRound(), raw.GetView().GetRound())
	assert.Equal(t, m.GetSender(), raw.GetSender())
	assert.Equal(t, m.GetSignature(), raw.GetSignature())
	assert.Equal(t, m.GetProposal(), raw.GetProposal())
	assert.Equal(t, m.GetProposalRound(), raw.GetProposalRound())
}

func TestProposalMessage_Verify(t *testing.T) {
	t.Parallel()

	t.Run("invalid view", func(t *testing.T) {
		t.Parallel()

		m := &ProposalMessage{
			View: nil,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageView)
	})

	t.Run("invalid sender", func(t *testing.T) {
		t.Parallel()

		m := &ProposalMessage{
			View:   &View{},
			Sender: nil,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageSender)
	})

	t.Run("invalid proposal", func(t *testing.T) {
		t.Parallel()

		m := &ProposalMessage{
			View:     &View{},
			Sender:   []byte{},
			Proposal: nil,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageProposal)
	})

	t.Run("invalid proposal round", func(t *testing.T) {
		t.Parallel()

		m := &ProposalMessage{
			View:          &View{},
			Sender:        []byte{},
			Proposal:      []byte{},
			ProposalRound: -2,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageProposalRound)
	})

	t.Run("valid proposal message", func(t *testing.T) {
		t.Parallel()

		m := &ProposalMessage{
			View: &View{
				Height: 1,
				Round:  0,
			},
			Sender:        []byte("sender"),
			Proposal:      []byte("proposal"),
			ProposalRound: -1,
		}

		assert.NoError(t, m.Verify())
	})
}

func TestPrevoteMessage_GetSignaturePayload(t *testing.T) {
	t.Parallel()

	// Create the proposal message
	m := &PrevoteMessage{
		View: &View{
			Height: 10,
			Round:  10,
		},
		Sender:    []byte("sender"),
		Signature: []byte("signature"),
	}

	// Get the signature payload
	payload := m.GetSignaturePayload()

	var raw PrevoteMessage

	require.NoError(t, proto.Unmarshal(payload, &raw))

	// Make sure the signature was not marshalled
	assert.Nil(t, raw.Signature)

	// Make sure other fields are intact
	assert.Equal(t, m.GetView().GetHeight(), raw.GetView().GetHeight())
	assert.Equal(t, m.GetView().GetRound(), raw.GetView().GetRound())
	assert.Equal(t, m.GetSender(), raw.GetSender())
}

func TestPrevoteMessage_Marshal(t *testing.T) {
	t.Parallel()

	// Create the proposal message
	m := &PrevoteMessage{
		View: &View{
			Height: 10,
			Round:  10,
		},
		Sender:    []byte("sender"),
		Signature: []byte("signature"),
	}

	// Marshal the message
	marshalled := m.Marshal()

	var raw PrevoteMessage

	require.NoError(t, proto.Unmarshal(marshalled, &raw))

	// Make sure other fields are intact
	assert.Equal(t, m.GetView().GetHeight(), raw.GetView().GetHeight())
	assert.Equal(t, m.GetView().GetRound(), raw.GetView().GetRound())
	assert.Equal(t, m.GetSender(), raw.GetSender())
	assert.Equal(t, m.GetSignature(), raw.GetSignature())
}

func TestPrevoteMessage_Verify(t *testing.T) {
	t.Parallel()

	t.Run("invalid view", func(t *testing.T) {
		t.Parallel()

		m := &PrevoteMessage{
			View: nil,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageView)
	})

	t.Run("invalid sender", func(t *testing.T) {
		t.Parallel()

		m := &PrevoteMessage{
			View:   &View{},
			Sender: nil,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageSender)
	})
}

func TestPrecommitMessage_GetSignaturePayload(t *testing.T) {
	t.Parallel()

	// Create the proposal message
	m := &PrecommitMessage{
		View: &View{
			Height: 10,
			Round:  10,
		},
		Sender:    []byte("sender"),
		Signature: []byte("signature"),
	}

	// Get the signature payload
	payload := m.GetSignaturePayload()

	var raw PrecommitMessage

	require.NoError(t, proto.Unmarshal(payload, &raw))

	// Make sure the signature was not marshalled
	assert.Nil(t, raw.Signature)

	// Make sure other fields are intact
	assert.Equal(t, m.GetView().GetHeight(), raw.GetView().GetHeight())
	assert.Equal(t, m.GetView().GetRound(), raw.GetView().GetRound())
	assert.Equal(t, m.GetSender(), raw.GetSender())
}

func TestPrecommitMessage_Marshal(t *testing.T) {
	t.Parallel()

	// Create the proposal message
	m := &PrecommitMessage{
		View: &View{
			Height: 10,
			Round:  10,
		},
		Sender:    []byte("sender"),
		Signature: []byte("signature"),
	}

	// Marshal the message
	marshalled := m.Marshal()

	var raw PrecommitMessage

	require.NoError(t, proto.Unmarshal(marshalled, &raw))

	// Make sure other fields are intact
	assert.Equal(t, m.GetView().GetHeight(), raw.GetView().GetHeight())
	assert.Equal(t, m.GetView().GetRound(), raw.GetView().GetRound())
	assert.Equal(t, m.GetSender(), raw.GetSender())
	assert.Equal(t, m.GetSignature(), raw.GetSignature())
}

func TestPrecommitMessage_Verify(t *testing.T) {
	t.Parallel()

	t.Run("invalid view", func(t *testing.T) {
		t.Parallel()

		m := &PrecommitMessage{
			View: nil,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageView)
	})

	t.Run("invalid sender", func(t *testing.T) {
		t.Parallel()

		m := &PrecommitMessage{
			View:   &View{},
			Sender: nil,
		}

		assert.ErrorIs(t, m.Verify(), ErrInvalidMessageSender)
	})
}
