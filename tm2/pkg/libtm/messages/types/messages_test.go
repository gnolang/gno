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

func TestView_Equals(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name        string
		views       []*View
		shouldEqual bool
	}{
		{
			"equal views",
			[]*View{
				{
					Height: 10,
					Round:  10,
				},
				{
					Height: 10,
					Round:  10,
				},
			},
			true,
		},
		{
			"not equal views",
			[]*View{
				{
					Height: 10,
					Round:  10,
				},
				{
					Height: 10,
					Round:  5, // different round
				},
			},
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				testCase.shouldEqual,
				testCase.views[0].Equals(testCase.views[1]),
			)
		})
	}
}

func TestProposalMessage_Equals(t *testing.T) {
	t.Parallel()

	t.Run("equal proposal messages", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender        = []byte("sender")
			signature     = []byte("signature")
			proposal      = []byte("proposal")
			proposalRound = int64(-1)

			left = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}

			right = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}
		)

		assert.True(t, left.Equals(right))
	})

	t.Run("view mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender        = []byte("sender")
			signature     = []byte("signature")
			proposal      = []byte("proposal")
			proposalRound = int64(-1)

			left = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}

			right = &ProposalMessage{
				View: &View{
					Height: view.Height,
					Round:  view.Round + 1, // round mismatch
				}, Sender: sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("sender mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender        = []byte("sender")
			signature     = []byte("signature")
			proposal      = []byte("proposal")
			proposalRound = int64(-1)

			left = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}

			right = &ProposalMessage{
				View:          view,
				Sender:        []byte("different sender"), // sender mismatch
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("signature mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender        = []byte("sender")
			signature     = []byte("signature")
			proposal      = []byte("proposal")
			proposalRound = int64(-1)

			left = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}

			right = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     []byte("different signature"), // signature mismatch
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("proposal mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender        = []byte("sender")
			signature     = []byte("signature")
			proposal      = []byte("proposal")
			proposalRound = int64(-1)

			left = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}

			right = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      []byte("different proposal"), // proposal mismatch
				ProposalRound: proposalRound,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("proposal round mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender        = []byte("sender")
			signature     = []byte("signature")
			proposal      = []byte("proposal")
			proposalRound = int64(-1)

			left = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound,
			}

			right = &ProposalMessage{
				View:          view,
				Sender:        sender,
				Signature:     signature,
				Proposal:      proposal,
				ProposalRound: proposalRound + 1, // proposal round mismatch
			}
		)

		assert.False(t, left.Equals(right))
	})
}

func TestPrevoteMessage_Equals(t *testing.T) {
	t.Parallel()

	t.Run("equal prevote messages", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}
		)

		assert.True(t, left.Equals(right))
	})

	t.Run("view mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrevoteMessage{
				View: &View{
					Height: view.Height,
					Round:  view.Round + 1, // round mismatch
				},
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("sender mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrevoteMessage{
				View:       view,
				Sender:     []byte("different sender"), // sender mismatch
				Signature:  signature,
				Identifier: id,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("signature mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  []byte("different signature"), // signature mismatch
				Identifier: id,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("identifier mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrevoteMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: []byte("different identifier"), // identifier mismatch
			}
		)

		assert.False(t, left.Equals(right))
	})
}

func TestPrecommitMessage_Equals(t *testing.T) {
	t.Parallel()

	t.Run("equal precommit messages", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}
		)

		assert.True(t, left.Equals(right))
	})

	t.Run("view mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrecommitMessage{
				View: &View{
					Height: view.Height,
					Round:  view.Round + 1, // round mismatch
				},
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("sender mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrecommitMessage{
				View:       view,
				Sender:     []byte("different sender"), // sender mismatch
				Signature:  signature,
				Identifier: id,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("signature mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  []byte("different signature"), // signature mismatch
				Identifier: id,
			}
		)

		assert.False(t, left.Equals(right))
	})

	t.Run("identifier mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			view = &View{
				Height: 10,
				Round:  0,
			}
			sender    = []byte("sender")
			signature = []byte("signature")
			id        = []byte("id")

			left = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: id,
			}

			right = &PrecommitMessage{
				View:       view,
				Sender:     sender,
				Signature:  signature,
				Identifier: []byte("different identifier"), // identifier mismatch
			}
		)

		assert.False(t, left.Equals(right))
	})
}
