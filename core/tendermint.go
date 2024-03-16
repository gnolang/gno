package core

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gnolang/go-tendermint/messages"

	"github.com/gnolang/go-tendermint/messages/types"
)

type Tendermint struct {
	// store is the message store
	store store

	verifier  Verifier
	node      Node
	broadcast Broadcast
	signer    Signer

	// logger is the consensus engine logger
	logger *slog.Logger

	// roundExpired is the channel for signalizing
	// round change events (to the next round, from the current one)
	roundExpired chan struct{}

	// timeouts hold state timeout information (constant)
	timeouts map[step]timeout

	// state is the current Tendermint consensus state
	state state

	// wg is the barrier for keeping all
	// parallel consensus processes synced
	wg sync.WaitGroup
}

// RunSequence runs the Tendermint consensus sequence for a given height,
// returning only when a proposal has been finalized (consensus reached), or
// the context has been cancelled
func (t *Tendermint) RunSequence(ctx context.Context, h uint64) []byte {
	t.logger.Debug(
		"RunSequence",
		slog.Uint64("height", h),
		slog.String("node", string(t.node.ID())),
	)

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

			// Check if the proposal has been finalized
			if proposal == nil {
				t.logger.Info(

					"RunSequence received empty proposal",
					slog.Uint64("height", h),
					slog.Uint64("round", t.state.LoadRound()),
				)
				// 65: Function OnTimeoutPrecommit(height, round) :
				// 66: 	if height = hP ∧ round = roundP then
				// 67: 		StartRound(roundP + 1)
				t.state.IncRound()

				continue
			}

			t.logger.Info(
				"RunSequence: proposal finalized",
				slog.Uint64("height", h),
			)

			return proposal
		case recvRound := <-t.watchForRoundJumps(ctxRound):
			t.logger.Info(
				"RunSequence: round jump",
				slog.Uint64("height", h),
				slog.Uint64("from", t.state.LoadRound()),
				slog.Uint64("to", recvRound),
			)

			teardown()
			t.state.SetRound(recvRound)
		case <-t.roundExpired:
			t.logger.Info(
				"RunSequence: round expired",
				slog.Uint64("height", h),
				slog.Uint64("round", t.state.LoadRound()),
			)

			teardown()
			t.state.IncRound()
		case <-ctx.Done():
			teardown()

			t.logger.Info(
				"RunSequence: context done",
				slog.Uint64("height", h),
				slog.Uint64("round", t.state.LoadRound()),
			)

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
			var majority bool

			select {
			case <-ctx.Done():
				return
			case getProposeFn := <-proposeCh:
				prpMsgs := getProposeFn()
				msgs := make([]Message, 0)

				messages.ConvertToInterface(prpMsgs, func(m *types.ProposalMessage) {
					msgs = append(msgs, m)
				})

				majority = t.verifier.Quorum(msgs)
			case getPrevoteFn := <-prevoteCh:
				prvMsgs := getPrevoteFn()
				msgs := make([]Message, 0)

				messages.ConvertToInterface(prvMsgs, func(m *types.PrevoteMessage) {
					msgs = append(msgs, m)
				})

				majority = t.verifier.Quorum(msgs)
			case getPrecommitFn := <-precommitCh:
				prcMsgs := getPrecommitFn()
				msgs := make([]Message, 0)

				messages.ConvertToInterface(prcMsgs, func(m *types.PrecommitMessage) {
					msgs = append(msgs, m)
				})

				majority = t.verifier.Quorum(msgs)
			}

			// check if the condition (F+1) is met
			// and signal the round jump
			if majority {
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
	height := t.state.LoadHeight()
	round := t.state.LoadRound()

	// Check if the current process is the proposer for this view
	if !t.verifier.IsProposer(t.node.ID(), height, round) {
		// The current process is NOT the proposer, only schedule a timeout
		//
		// 21: 	schedule OnTimeoutPropose(hP , roundP) to be executed after timeoutPropose(roundP)
		var (
			callback = func() {
				t.onTimeoutPropose(round)
			}
			timeoutPropose = t.timeouts[propose].calculateTimeout(round)
		)

		t.logger.Debug(
			"scheduling timeoutPropose",
			slog.Uint64("height", height),
			slog.Uint64("round", round),
			slog.Duration("timeout", timeoutPropose),
		)

		t.scheduleTimeout(ctx, timeoutPropose, callback)

		return
	}

	// 14: if proposer(hp, roundP) = p then
	//
	// The proposal value can either be:
	// - an old (valid / locked) proposal from a previous round
	// - a completely new proposal (built from scratch)
	//
	// 15: 	if validValueP != nil then
	// 16: 		proposal ← validValueP
	proposal := t.state.validValue

	// Check if a new proposal needs to be built
	if proposal == nil {
		t.logger.Info(
			"building a proposal",
			slog.Uint64("height", height),
			slog.Uint64("round", round),
		)
		// No previous valid value present,
		// build a new proposal.
		//
		// 17: 	else
		// 18: 		proposal ← getValue()
		proposal = t.node.BuildProposal(height)
	}

	// Build the propose message
	var (
		proposeMessage = t.buildProposalMessage(proposal)
		id             = t.node.Hash(proposal)
	)

	// Broadcast the proposal to other consensus nodes
	//
	// 19: 		broadcast <PROPOSAL, hp, roundP, proposal, validRoundP>
	t.broadcast.BroadcastProposal(proposeMessage)

	// Save the accepted proposal in the state.
	// NOTE: This is different from validValue / lockedValue,
	// since they require a 2F+1 quorum of specific messages
	// in order to be set, whereas this is simply a reference
	// value for different states (prevote, precommit)
	t.state.acceptedProposal = proposeMessage
	t.state.acceptedProposalID = id

	// Build and broadcast the prevote message
	//
	// 24/30: broadcast <PREVOTE, hP, roundP, id(v)>
	t.broadcast.BroadcastPrevote(t.buildPrevoteMessage(id))

	// Since the current process is the proposer,
	// it can directly move to the prevote state
	// 27/33: stepP ← prevote
	t.state.step.Set(prevote)
}

// runStates runs the consensus states, depending on the current step
func (t *Tendermint) runStates(ctx context.Context) []byte {
	for {
		currentStep := t.state.step.Load()

		switch currentStep {
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
// waits for a valid PROPOSE message.
// This state handles the following situations:
//
// - The proposer for view (hP, roundP) has proposed a value with a proposal round -1 (first ever proposal for height)
// 22: upon <PROPOSAL, hP, roundP, v, −1> from proposer(hP, roundP) while stepP = propose do
// 23: 	if valid(v) ∧ (lockedRoundP = −1 ∨ lockedValueP = v) then
// 24: 		broadcast <PREVOTE, hP, roundP, id(v)>
// 25: 	else
// 26: 		broadcast <PREVOTE, hP, roundP, nil>
// 27: 	stepP ← prevote
//
// - The proposer for view (hP, roundP) has proposed a value that was accepted in some previous round
// 28: upon <PROPOSAL, hP, roundP, v, vr> from proposer(hP, roundP) AND 2f + 1 <PREVOTE, hP, vr, id(v)>
// while stepP = propose ∧ (vr >= 0 ∧ vr < roundP) do
// 29: if valid(v) ∧ (lockedRoundP ≤ vr ∨ lockedValueP = v) then
// 30: 	broadcast <PREVOTE, hp, roundP, id(v)>
// 31: else
// 32: 	broadcast <PREVOTE, hp, roundP, nil>
// 33: stepP ← prevote
//
// NOTE: the proposer for view (height, round) will send ONLY 1 proposal, be it a new one or an old agreed value
func (t *Tendermint) runPropose(ctx context.Context) {
	// TODO make thread safe
	_ = t.state.view

	ch, unsubscribeFn := t.store.SubscribeToPropose()
	defer unsubscribeFn()

	for {
		select {
		case <-ctx.Done():
			return
		case getMessagesFn := <-ch:
			prpMsgs := getMessagesFn()
			msgs := make([]Message, 0)

			messages.ConvertToInterface(prpMsgs, func(m *types.ProposalMessage) {
				msgs = append(msgs, m)
			})

			majority := t.verifier.Quorum(msgs)

			if majority {
				t.state.step.Set(prevote)
			}
		}
	}
}

// runPrevote runs the prevote state in which the process
// waits for a valid PREVOTE messages.
// This state handles the following situations:
//
// - A validator has received 2F+1 PREVOTE messages with a valid ID for the previously accepted proposal
// 36: upon ... AND 2f + 1 <PREVOTE, hP, roundP, id(v)> while valid(v) ∧ stepP ≥ prevote for the first time do
// 37: if stepP = prevote then
// 38: 	lockedValueP ← v
// 39: 	lockedRoundP ← roundP
// 40: 	broadcast <PRECOMMIT, hP, roundP, id(v))>
// 41: 	stepP ← precommit
// 42: validValueP ← v
// 43: validRoundP ← roundP
//
// - A validator has received 2F+1 PREVOTE messages with a NIL ID
// 34: upon 2f + 1 <PREVOTE, hp, roundP, ∗> while stepP = prevote for the first time do
// 35: schedule OnTimeoutPrevote(hP , roundP) to be executed after timeoutPrevote(roundP)
func (t *Tendermint) runPrevote(ctx context.Context) {
	// TODO make thread safe
	_ = t.state.view

	ch, unsubscribeFn := t.store.SubscribeToPrevote()
	defer unsubscribeFn()

	for {
		select {
		case <-ctx.Done():
			return
		case getMessagesFn := <-ch:
			prvMsgs := getMessagesFn()
			msgs := make([]Message, 0)

			messages.ConvertToInterface(prvMsgs, func(m *types.PrevoteMessage) {
				msgs = append(msgs, m)
			})

			majority := t.verifier.Quorum(msgs)

			if majority {
				t.state.step.Set(precommit)
			}
		}
	}
}

// runPrecommit runs the precommit state in which the process
// waits for a valid PRECOMMIT messages.
// This state handles the following situations:
//
// - A validator has received 2F+1 PRECOMMIT messages with a valid ID for the previously accepted proposal
// 49: upon <PROPOSAL, hP, r, v, ∗> from proposer(hP, r) AND 2f + 1 <PRECOMMIT, hP, r, id(v)>
// while decisionP[hP] = nil do
// 50: if valid(v) then
// 51: 	decisionP[hp] = v
// 52: 	hP ← hP + 1
// 53: 	reset lockedRoundP, lockedValueP, validRoundP and validValueP to initial values and empty message log
// 54: 	StartRound(0)
//
// - A validator has received 2F+1 PRECOMMIT messages with any value (valid ID or NIL)
// 47: upon 2f + 1 <PRECOMMIT, hP, roundP, ∗> for the first time do
// 48: schedule OnTimeoutPrecommit(hP , roundP) to be executed after timeoutPrecommit(roundP)
//
// TODO @zivkovicmilos: @petar-dambovaliev, I think we can just return nil from this method
// in case OnTimeoutPrecommit triggers and we still don't have 2F+1 valid PRECOMMIT messages.
// This makes it easy to handle it in the top-level run loop that parses the finalized proposal:
// 65: Function OnTimeoutPrecommit(height, round) :
// 66: 	if height = hP ∧ round = roundP then
// 67: 		StartRound(roundP + 1)
func (t *Tendermint) runPrecommit(ctx context.Context) []byte {
	// TODO make thread safe
	_ = t.state.view

	ch, unsubscribeFn := t.store.SubscribeToPrecommit()
	defer unsubscribeFn()

	for {
		select {
		case <-ctx.Done():
			return nil
		case getMessagesFn := <-ch:
			prcMsgs := getMessagesFn()
			msgs := make([]Message, 0)

			messages.ConvertToInterface(prcMsgs, func(m *types.PrecommitMessage) {
				msgs = append(msgs, m)
			})

			majority := t.verifier.Quorum(msgs)

			if majority {
				return nil
			}
		}
	}
}
