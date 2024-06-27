package core

import (
	"errors"
	"fmt"

	"github.com/gnolang/libtm/messages/types"
)

var (
	ErrInvalidMessageSignature = errors.New("invalid message signature")
	ErrMessageFromNonValidator = errors.New("message is from a non-validator")
	ErrEarlierHeightMessage    = errors.New("message is for an earlier height")
	ErrEarlierRoundMessage     = errors.New("message is for an earlier round")
)

// AddProposalMessage verifies and adds a new proposal message to the consensus engine
func (t *Tendermint) AddProposalMessage(message *types.ProposalMessage) error {
	// Verify the incoming message
	if err := t.verifyMessage(message); err != nil {
		return fmt.Errorf("unable to verify proposal message, %w", err)
	}

	// Add the message to the store
	t.store.addProposalMessage(message)

	return nil
}

// AddPrevoteMessage verifies and adds a new prevote message to the consensus engine
func (t *Tendermint) AddPrevoteMessage(message *types.PrevoteMessage) error {
	// Verify the incoming message
	if err := t.verifyMessage(message); err != nil {
		return fmt.Errorf("unable to verify prevote message, %w", err)
	}

	// Add the message to the store
	t.store.addPrevoteMessage(message)

	return nil
}

// AddPrecommitMessage verifies and adds a new precommit message to the consensus engine
func (t *Tendermint) AddPrecommitMessage(message *types.PrecommitMessage) error {
	// Verify the incoming message
	if err := t.verifyMessage(message); err != nil {
		return fmt.Errorf("unable to verify precommit message, %w", err)
	}

	// Add the message to the store
	t.store.addPrecommitMessage(message)

	return nil
}

// verifyMessage is the common base message verification
func (t *Tendermint) verifyMessage(message Message) error {
	// Check if the message is valid
	if err := message.Verify(); err != nil {
		return fmt.Errorf("unable to verify message, %w", err)
	}

	// Make sure the message sender is a validator
	if !t.verifier.IsValidator(message.GetSender()) {
		return ErrMessageFromNonValidator
	}

	// Get the signature payload
	signPayload := message.GetSignaturePayload()

	// Make sure the signature is valid
	if !t.signer.IsValidSignature(signPayload, message.GetSignature()) {
		return ErrInvalidMessageSignature
	}

	// Make sure the message view is valid
	var (
		view = message.GetView()

		currentHeight = t.state.getHeight()
		currentRound  = t.state.getRound()
	)

	// Make sure the height is valid.
	// The message height needs to be the current state height, or greater
	if currentHeight > view.GetHeight() {
		return ErrEarlierHeightMessage
	}

	// Make sure the round is valid.
	// The message rounds needs to be >= the current round
	if currentRound > view.GetRound() {
		return ErrEarlierRoundMessage
	}

	return nil
}
