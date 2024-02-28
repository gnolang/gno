package core

import (
	"errors"
	"fmt"

	"github.com/gnolang/go-tendermint/messages/types"
)

var (
	ErrMessageNotSet           = errors.New("message not set")
	ErrMessagePayloadNotSet    = errors.New("message payload not set")
	ErrInvalidMessageSignature = errors.New("invalid message signature")
	ErrMessageFromNonValidator = errors.New("message is from a non-validator")
	ErrEarlierHeightMessage    = errors.New("message is for an earlier height")
	ErrEarlierRoundMessage     = errors.New("message is for an earlier round")
)

// AddMessage verifies and adds a new message to the consensus engine
func (t *Tendermint) AddMessage(message *types.Message) error {
	// Verify the incoming message
	if err := t.verifyMessage(message); err != nil {
		return fmt.Errorf("unable to verify message, %w", err)
	}

	// Add the message to the store
	t.store.AddMessage(message)

	return nil
}

// verifyMessage verifies the incoming consensus message (base verification)
func (t *Tendermint) verifyMessage(message *types.Message) error {
	// Make sure the message is present
	if message == nil {
		return ErrMessageNotSet
	}

	// Make sure the message payload is present
	if message.Payload == nil {
		return ErrMessagePayloadNotSet
	}

	// Get the signature payload
	signPayload, err := message.GetSignaturePayload()
	if err != nil {
		return fmt.Errorf("unable to get message signature payload, %w", err)
	}

	// Make sure the signature is valid
	if !t.signer.IsValidSignature(signPayload, message.Signature) {
		return ErrInvalidMessageSignature
	}

	// Extract individual message data
	var (
		sender []byte
		view   *types.View
	)

	switch message.Type {
	case types.MessageType_PROPOSAL:
		// Get the proposal message
		payload := message.GetProposalMessage()
		if payload == nil || !payload.IsValid() {
			return types.ErrInvalidMessagePayload
		}

		sender = payload.GetFrom()
		view = payload.GetView()
	case types.MessageType_PREVOTE:
		// Get the prevote message
		payload := message.GetPrevoteMessage()
		if payload == nil || !payload.IsValid() {
			return types.ErrInvalidMessagePayload
		}

		sender = payload.GetFrom()
		view = payload.GetView()
	case types.MessageType_PRECOMMIT:
		// Get the precommit message
		payload := message.GetPrecommitMessage()
		if payload == nil || !payload.IsValid() {
			return types.ErrInvalidMessagePayload
		}

		sender = payload.GetFrom()
		view = payload.GetView()
	}

	// Make sure the message sender is a validator
	if !t.verifier.IsValidator(sender) {
		return ErrMessageFromNonValidator
	}

	// Make sure the message view is valid
	var (
		currentHeight = t.state.view.GetHeight() // TODO make thread safe
		currentRound  = t.state.view.GetRound()  // TODO make thread safe
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
