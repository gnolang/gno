package core

import (
	"context"
	"sync"

	"github.com/gnolang/go-tendermint/messages/types"
)

// TODO define the finalized proposal

type Tendermint struct {
	verifier  Verifier
	node      Node
	broadcast Broadcast
	signer    Signer

	// wg is the barrier for keeping all
	// parallel consensus processes synced
	wg sync.WaitGroup

	// state is the current Tendermint consensus state
	state *state

	// store is the message store
	store *store

	// roundExpired is the channel for signalizing
	// round change events (to the next round, from the current one)
	roundExpired chan struct{}

	// timeouts hold state timeout information (constant)
	timeouts map[step]timeout
}

// RunSequence runs the Tendermint consensus sequence for a given height,
// returning only when a proposal has been finalized (consensus reached), or
// the context has been cancelled
func (t *Tendermint) RunSequence(ctx context.Context, h uint64) []byte {
	// Initialize the state before starting the sequence
	t.state = newState(&types.View{
		Height: h,
		Round:  0,
	})

	for {
		// Set up the round context
		ctxRound, cancelRound := context.WithCancel(ctx)
		teardown := func() {
			cancelRound()
			t.wg.Wait()
		}

		select {
		case proposal := <-t.finalizeProposal(ctxRound):
			teardown()

			return proposal
		case _ = <-t.watchForRoundJumps(ctxRound):
			teardown()

		// TODO start NEW round (that was received)
		case <-t.roundExpired:
			teardown()

			// TODO start NEXT round (view.Round + 1)
		case <-ctx.Done():
			teardown()

			// TODO log
			return nil
		}
	}
}

// watchForRoundJumps monitors for F+1 (any) messages from a future round, and
// triggers the round switch context (channel) accordingly
func (t *Tendermint) watchForRoundJumps(ctx context.Context) <-chan uint64 {
	// TODO make thread safe
	var (
		_  = t.state.view
		ch = make(chan uint64, 1)
	)

	t.wg.Add(1)

	go func() {
		proposeCh, unsubscribeProposeFn := t.store.SubscribeToPropose()
		prevoteCh, unsubscribePrevoteFn := t.store.SubscribeToPrevote()
		precommitCh, unsubscribePrecommitFn := t.store.SubscribeToPrecommit()

		defer func() {
			unsubscribeProposeFn()
			unsubscribePrevoteFn()
			unsubscribePrecommitFn()
		}()

		signalRoundJump := func(round uint64) {
			select {
			case <-ctx.Done():
			case ch <- round:
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case getProposeFn := <-proposeCh:
				// TODO count messages
				_ = getProposeFn()
			case getPrevoteFn := <-prevoteCh:
				// TODO count messages
				_ = getPrevoteFn()
			case getPrecommitFn := <-precommitCh:
				// TODO count messages
				_ = getPrecommitFn()
			}

			// TODO check if the condition (F+1) is met
			// and signal the round jump
			if false {
				signalRoundJump(0)
			}
		}
	}()

	return ch
}

// finalizeProposal starts the proposal finalization sequence
func (t *Tendermint) finalizeProposal(ctx context.Context) <-chan []byte {
	ch := make(chan []byte, 1)

	t.wg.Add(1)

	go func() {
		defer func() {
			close(ch)
			t.wg.Done()
		}()

		// Start the consensus round
		t.startRound(ctx)

		// Run the consensus state machine, and save the finalized proposal (if any)
		if finalizedProposal := t.runStates(ctx); finalizedProposal != nil {
			ch <- finalizedProposal
		}
	}()

	return ch
}

// startRound starts the consensus round.
// It is a required middle step (proposal evaluation) before
// the state machine is in full swing and
// the runs the same flow for everyone (proposer / non-proposers)
func (t *Tendermint) startRound(ctx context.Context) {
	// TODO make thread safe
	// Check if the current process is the proposer for this view
	if !t.verifier.IsProposer(t.node.ID(), t.state.view.Height, t.state.view.Round) {
		// The current process is NOT the proposer, only schedule a timeout
		t.scheduleTimeoutPropose(ctx)

		return
	}

	// The proposal value can either be:
	// - an old (valid / locked) proposal from a previous round
	// - a completely new proposal (built from scratch)
	proposal := t.state.validValue

	// Check if a new proposal needs to be built
	if proposal == nil {
		// No previous valid value present,
		// build a new proposal
		proposal = t.node.BuildProposal(t.state.view.Height)
	}

	// Build the propose message
	var (
		proposeMessage = t.buildProposalMessage(proposal)
		id             = t.node.Hash(proposal)
	)

	// Broadcast the proposal to other consensus nodes
	t.broadcastProposal(proposeMessage)

	// TODO make thread safe
	// Save the accepted proposal in the state.
	// NOTE: This is different from validValue / lockedValue,
	// since they require a 2F+1 quorum of specific messages
	// in order to be set, whereas this is simply a reference
	// value for different states (prevote, precommit)
	t.state.acceptedProposal = proposeMessage
	t.state.acceptedProposalID = id

	// Build and broadcast the prevote message
	t.broadcastPrevote(t.buildPrevoteMessage(id))

	// Since the current process is the proposer,
	// it can directly move to the prevote state
	// TODO make threads safe
	t.state.step = prevote
}

// runStates runs the consensus states, depending on the current step
func (t *Tendermint) runStates(ctx context.Context) []byte {
	for {
		// TODO make thread safe
		switch t.state.step {
		case propose:
			t.runPropose(ctx)
		case prevote:
			t.runPrevote(ctx)
		case precommit:
			return t.runPrecommit(ctx)
		}
	}
}

// runPropose runs the propose state in which the process
// waits for a valid PROPOSE message
func (t *Tendermint) runPropose(ctx context.Context) {
	// TODO make thread safe
	var (
		_ = t.state.view
	)

	ch, unsubscribeFn := t.store.SubscribeToPropose()
	defer unsubscribeFn()

	for {
		select {
		case <-ctx.Done():
			return
		case getMessagesFn := <-ch:
			// TODO filter and verify messages
			_ = getMessagesFn()

			// TODO move to prevote if the proposal is valid
		}
	}
}

// runPrevote runs the prevote state in which the process
// waits for a valid PREVOTE messages
func (t *Tendermint) runPrevote(ctx context.Context) {
	// TODO make thread safe
	var (
		_ = t.state.view
	)

	ch, unsubscribeFn := t.store.SubscribeToPrevote()
	defer unsubscribeFn()

	for {
		select {
		case <-ctx.Done():
			return
		case getMessagesFn := <-ch:
			// TODO filter and verify messages
			_ = getMessagesFn()

			// TODO move to precommit if the proposal is valid
		}
	}
}

// runPrecommit runs the precommit state in which the process
// waits for a valid PRECOMMIT messages
func (t *Tendermint) runPrecommit(ctx context.Context) []byte {
	// TODO make thread safe
	var (
		_ = t.state.view
	)

	ch, unsubscribeFn := t.store.SubscribeToPrecommit()
	defer unsubscribeFn()

	for {
		select {
		case <-ctx.Done():
			return nil
		case getMessagesFn := <-ch:
			// TODO filter and verify messages
			_ = getMessagesFn()

			// TODO move to precommit if the proposal is valid
		}
	}
}

// AddMessage verifies and adds a new message to the consensus engine
func (t *Tendermint) AddMessage(message *types.Message) {
	// Make sure the message is present
	if message == nil {
		return
	}

	// Make sure the message payload is present
	if message.Payload == nil {
		return
	}

	// TODO verify the message sender

	// TODO verify the message signature

	// TODO verify the message height

	// TODO verify the message round

	// TODO verify the message content (fields set)
}
