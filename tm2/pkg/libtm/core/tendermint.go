package core

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"

	"github.com/gnolang/libtm/messages"

	"github.com/gnolang/libtm/messages/types"
)

// Tendermint is the single consensus engine instance
type Tendermint struct {
	// store is the message store
	store store

	verifier  Verifier
	node      Node
	broadcast Broadcast
	signer    Signer

	// logger is the consensus engine logger
	logger *slog.Logger

	// timeouts hold state timeout information (constant)
	timeouts map[step]Timeout

	// state is the current Tendermint consensus state
	state state

	// wg is the barrier for keeping all
	// parallel consensus processes synced
	wg sync.WaitGroup
}

// FinalizedProposal is the finalized proposal wrapper, that
// contains the raw proposal data, and the ID of the data (usually hash)
type FinalizedProposal struct {
	Data []byte // the raw proposal data, accepted proposal
	ID   []byte // the ID of the proposal (usually hash)
}

// newFinalizedProposal creates a new finalized proposal wrapper
func newFinalizedProposal(data, id []byte) *FinalizedProposal {
	return &FinalizedProposal{
		Data: data,
		ID:   id,
	}
}

// NewTendermint creates a new instance of the Tendermint consensus engine
func NewTendermint(
	verifier Verifier,
	node Node,
	broadcast Broadcast,
	signer Signer,
	opts ...Option,
) *Tendermint {
	t := &Tendermint{
		state:     newState(),
		store:     newStore(),
		verifier:  verifier,
		node:      node,
		broadcast: broadcast,
		signer:    signer,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		timeouts:  getDefaultTimeoutMap(),
	}

	// Apply any options
	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RunSequence runs the Tendermint consensus sequence for a given height,
// returning only when a proposal has been finalized (consensus reached), or
// the context has been cancelled
func (t *Tendermint) RunSequence(ctx context.Context, h uint64) *FinalizedProposal {
	t.logger.Debug(
		"RunSequence",
		slog.Uint64("height", h),
		slog.String("node", string(t.node.ID())),
	)

	// Initialize the state before starting the sequence
	t.state.setHeight(h)

	// Grab the process view
	view := &types.View{
		Height: h,
		Round:  t.state.getRound(),
	}

	// Drop all old messages
	t.store.dropMessages(view)

	for {
		// set up the round context
		ctxRound, cancelRound := context.WithCancel(ctx)
		teardown := func() {
			cancelRound()
			t.wg.Wait()
		}

		select {
		case proposal := <-t.finalizeProposal(ctxRound):
			teardown()

			// Check if the proposal has been finalized
			if proposal != nil {
				t.logger.Info(
					"RunSequence: proposal finalized",
					slog.Uint64("height", h),
					slog.String("node", string(t.node.ID())),
				)

				return proposal
			}

			t.logger.Info(
				"RunSequence round expired",
				slog.Uint64("height", h),
				slog.Uint64("round", t.state.getRound()),
				slog.String("node", string(t.node.ID())),
			)

			// 65: Function OnTimeoutPrecommit(height, round) :
			// 66: 	if height = hP ∧ round = roundP then
			// 67: 		StartRound(roundP + 1)
			t.state.increaseRound()
			t.state.step.set(propose)
		case recvRound := <-t.watchForRoundJumps(ctxRound):
			teardown()

			t.logger.Info(
				"RunSequence: round jump",
				slog.Uint64("height", h),
				slog.Uint64("from", t.state.getRound()),
				slog.Uint64("to", recvRound),
				slog.String("node", string(t.node.ID())),
			)

			t.state.setRound(recvRound)
			t.state.step.set(propose)
		case <-ctx.Done():
			teardown()

			t.logger.Info(
				"RunSequence: context done",
				slog.Uint64("height", h),
				slog.Uint64("round", t.state.getRound()),
				slog.String("node", string(t.node.ID())),
			)

			return nil
		}
	}
}

// watchForRoundJumps monitors for F+1 (any) messages from a future round, and
// triggers the round switch context (channel) accordingly
func (t *Tendermint) watchForRoundJumps(ctx context.Context) <-chan uint64 {
	var (
		height = t.state.getHeight()
		round  = t.state.getRound()

		ch = make(chan uint64, 1)
	)

	// Signals the round jump to the given channel
	signalRoundJump := func(round uint64) {
		select {
		case <-ctx.Done():
		case ch <- round:
		}
	}

	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		var (
			proposeCh, unsubscribeProposeFn     = t.store.subscribeToPropose()
			prevoteCh, unsubscribePrevoteFn     = t.store.subscribeToPrevote()
			precommitCh, unsubscribePrecommitFn = t.store.subscribeToPrecommit()
		)

		defer func() {
			unsubscribeProposeFn()
			unsubscribePrevoteFn()
			unsubscribePrecommitFn()
		}()

		var (
			isValidProposeFn = func(m *types.ProposalMessage) bool {
				view := m.GetView()

				return view.GetRound() > round && view.GetHeight() == height
			}
			isValidPrevoteFn = func(m *types.PrevoteMessage) bool {
				view := m.GetView()

				return view.GetRound() > round && view.GetHeight() == height
			}
			isValidPrecommitFn = func(m *types.PrecommitMessage) bool {
				view := m.GetView()

				return view.GetRound() > round && view.GetHeight() == height
			}
		)

		var (
			proposeCache   = newMessageCache[*types.ProposalMessage](isValidProposeFn)
			prevoteCache   = newMessageCache[*types.PrevoteMessage](isValidPrevoteFn)
			precommitCache = newMessageCache[*types.PrecommitMessage](isValidPrecommitFn)
		)

		generateRoundMap := func(messages ...[]Message) map[uint64][]Message {
			combined := make([]Message, 0)
			for _, message := range messages {
				combined = append(combined, message...)
			}

			// Group messages by round
			roundMap := make(map[uint64][]Message)

			for _, message := range combined {
				messageRound := message.GetView().GetRound()
				roundMap[messageRound] = append(roundMap[messageRound], message)
			}

			return roundMap
		}

		for {
			select {
			case <-ctx.Done():
				close(ch)

				return
			case getProposeFn := <-proposeCh:
				proposeCache.addMessages(getProposeFn())
			case getPrevoteFn := <-prevoteCh:
				prevoteCache.addMessages(getPrevoteFn())
			case getPrecommitFn := <-precommitCh:
				precommitCache.addMessages(getPrecommitFn())
			}

			var (
				proposeMessages   = proposeCache.getMessages()
				prevoteMessages   = prevoteCache.getMessages()
				precommitMessages = precommitCache.getMessages()
			)

			var (
				convertedPropose   = make([]Message, 0, len(proposeMessages))
				convertedPrevote   = make([]Message, 0, len(prevoteMessages))
				convertedPrecommit = make([]Message, 0, len(precommitMessages))
			)

			messages.ConvertToInterface(proposeMessages, func(m *types.ProposalMessage) {
				convertedPropose = append(convertedPropose, m)
			})

			messages.ConvertToInterface(prevoteMessages, func(m *types.PrevoteMessage) {
				convertedPrevote = append(convertedPrevote, m)
			})

			messages.ConvertToInterface(precommitMessages, func(m *types.PrecommitMessage) {
				convertedPrecommit = append(convertedPrecommit, m)
			})

			// Generate the round map
			roundMap := generateRoundMap(
				convertedPropose,
				convertedPrevote,
				convertedPrecommit,
			)

			// Find the highest round that satisfies an F+1 voting power majority.
			// This max round will always need to be > 0
			maxRound := uint64(0)

			for messageRound, roundMessages := range roundMap {
				if !t.hasFaultyMajority(roundMessages) {
					continue
				}

				if messageRound > maxRound {
					maxRound = messageRound
				}
			}

			// Make sure the max round that has a faulty majority
			// is actually greater than the process round
			if maxRound > round {
				signalRoundJump(maxRound)

				return
			}
		}
	}()

	return ch
}

// finalizeProposal starts the proposal finalization sequence
func (t *Tendermint) finalizeProposal(ctx context.Context) <-chan *FinalizedProposal {
	ch := make(chan *FinalizedProposal, 1)

	t.wg.Add(1)

	go func() {
		defer func() {
			close(ch)
			t.wg.Done()
		}()

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
func (t *Tendermint) startRound(height, round uint64) {
	// 14: if proposer(hp, roundP) = p then
	//
	// The proposal value can either be:
	// - an old (valid / locked) proposal from a previous round
	// - a completely new proposal (built from scratch)
	//
	// 15: 	if validValueP != nil then
	// 16: 		proposal ← validValueP
	var (
		proposal      = t.state.validValue
		proposalRound = t.state.validRound
	)

	// Check if a new proposal needs to be built
	if proposal == nil {
		t.logger.Info(
			"building a proposal",
			slog.Uint64("height", height),
			slog.Uint64("round", round),
			slog.String("node", string(t.node.ID())),
		)

		// No previous valid value present,
		// build a new proposal
		//
		// 17: 	else
		// 18: 		proposal ← getValue()
		proposal = t.node.BuildProposal(height)
	}

	// Build the propose message
	var (
		proposeMessage = t.buildProposalMessage(proposal, proposalRound)
		id             = t.node.Hash(proposal)
	)

	// Broadcast the proposal to other consensus nodes
	//
	// 19: 		broadcast <PROPOSAL, hp, roundP, proposal, validRoundP>
	t.broadcast.BroadcastPropose(proposeMessage)

	// Save the accepted proposal in the state.
	// NOTE: This is different from validValue / lockedValue,
	// since they require a 2F+1 quorum of specific messages
	// in order to be set, whereas this is simply a reference
	// value for different states (prevote, precommit)
	t.state.acceptedProposal = proposal
	t.state.acceptedProposalID = id

	// Build and broadcast the prevote message
	//
	// 24/30: broadcast <PREVOTE, hP, roundP, id(v)>
	t.broadcast.BroadcastPrevote(t.buildPrevoteMessage(id))

	// Since the current process is the proposer,
	// it can directly move to the prevote state
	// 27/33: stepP ← prevote
	t.state.step.set(prevote)
}

// runStates runs the consensus states, depending on the current step
func (t *Tendermint) runStates(ctx context.Context) *FinalizedProposal {
	for {
		currentStep := t.state.step.get()

		select {
		case <-ctx.Done():
			return nil
		default:
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
	var (
		height = t.state.getHeight()
		round  = t.state.getRound()

		lockedRound = t.state.lockedRound
		lockedValue = t.state.lockedValue
	)

	t.logger.Debug(
		"entering propose state",
		slog.Uint64("height", height),
		slog.Uint64("round", round),
		slog.String("node", string(t.node.ID())),
	)

	// Check if the current process is the proposer for this view
	if t.verifier.IsProposer(t.node.ID(), height, round) {
		// Start the round by constructing and broadcasting a proposal
		t.startRound(height, round)

		return
	}

	// The current process is NOT the proposer, schedule a timeout
	//
	// 21: 	schedule OnTimeoutPropose(hP , roundP) to be executed after timeoutPropose(roundP)
	var (
		expiredCh                 = make(chan struct{}, 1)
		timerCtx, cancelTimeoutFn = context.WithCancel(ctx)
		timeoutPropose            = t.timeouts[propose].CalculateTimeout(round)
	)

	// Defer the timeout timer cancellation
	defer cancelTimeoutFn()

	t.logger.Debug(
		"scheduling timeoutPropose",
		slog.Uint64("height", height),
		slog.Uint64("round", round),
		slog.Duration("timeout", timeoutPropose),
		slog.String("node", string(t.node.ID())),
	)

	t.scheduleTimeout(timerCtx, timeoutPropose, expiredCh)

	// Subscribe to all propose messages
	// (=current height; unique; >= current round)
	ch, unsubscribeFn := t.store.subscribeToPropose()
	defer unsubscribeFn()

	// Set up the verification callback.
	// The idea is to get the single proposal from the proposer for the view (height, round),
	// and verify if it is valid.
	// If it turns out the proposal is not valid (the first one received),
	// then the protocol needs to move to the prevote state, after
	// broadcasting a PREVOTE message with a NIL ID
	isFromProposerFn := func(proposal *types.ProposalMessage) bool {
		// Make sure the proposal view matches the process view
		if round != proposal.GetView().GetRound() {
			return false
		}

		// Check if the proposal came from the proposer
		// for the current view
		return t.verifier.IsProposer(proposal.GetSender(), height, round)
	}

	// Validates the proposal by examining the proposal params
	isValidProposal := func(proposal []byte, proposalRound int64) bool {
		// Basic proposal message verification
		if proposalRound < 0 {
			// Make sure there is no locked round (-1), OR
			// that the locked value matches the proposal value
			if lockedRound != -1 && !bytes.Equal(lockedValue, proposal) {
				return false
			}
		} else {
			// Make sure the proposal round is an earlier round
			// than the current process round (old proposal)
			if proposalRound >= int64(round) {
				return false
			}

			// Make sure the locked round value is <= the proposal round, OR
			// that the locked value matches the proposal value
			if lockedRound > proposalRound && !bytes.Equal(lockedValue, proposal) {
				return false
			}
		}

		// Make sure the proposal itself is valid
		return t.verifier.IsValidProposal(proposal, height)
	}

	// Create the message cache (local to this context only)
	cache := newMessageCache[*types.ProposalMessage](isFromProposerFn)

	for {
		select {
		case <-ctx.Done():
			return
		case <-expiredCh:
			// Broadcast a PREVOTE message with a NIL ID
			// 59: broadcast ⟨PREVOTE, hP, roundP, nil⟩
			t.broadcast.BroadcastPrevote(t.buildPrevoteMessage(nil))

			// Move to the prevote state
			// 60: stepP ← prevote
			t.state.step.set(prevote)

			return
		case getMessagesFn := <-ch:
			// Add the messages to the cache
			cache.addMessages(getMessagesFn())

			// Check if at least 1 proposal message is valid,
			// after validation and filtering
			proposalMessages := cache.getMessages()

			if len(proposalMessages) == 0 {
				// No valid proposal message yet
				continue
			}

			proposalMessage := proposalMessages[0]

			// Validate the proposal received
			if !isValidProposal(proposalMessage.Proposal, proposalMessage.ProposalRound) {
				// Broadcast a PREVOTE message with a NIL ID
				// 26: broadcast ⟨PREVOTE, hP, roundP, nil⟩
				// 32: broadcast ⟨PREVOTE, hP, roundP, nil⟩
				t.broadcast.BroadcastPrevote(t.buildPrevoteMessage(nil))

				// Move to the prevote state
				// 27: stepP ← prevote
				// 33: stepP ← prevote
				t.state.step.set(prevote)

				t.logger.Debug(
					"received invalid proposal",
					slog.Uint64("height", height),
					slog.Uint64("round", round),
					slog.String("node", string(t.node.ID())),
				)

				return
			}

			// Get the proposal from the message
			proposal := proposalMessage.GetProposal()

			// Generate the proposal ID
			id := t.node.Hash(proposal)

			// Accept the proposal, since it is valid
			t.state.acceptedProposal = proposal
			t.state.acceptedProposalID = id

			// Broadcast the PREVOTE message with a valid ID
			// 24: broadcast ⟨PREVOTE, hP, roundP, id(v)⟩
			// 30: broadcast ⟨PREVOTE, hP, roundP, id(v)⟩
			t.broadcast.BroadcastPrevote(t.buildPrevoteMessage(id))

			// Move to the prevote state
			// 27: stepP ← prevote
			// 33: stepP ← prevote
			t.state.step.set(prevote)

			return
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
// 44: upon 2f + 1 ⟨PREVOTE, hp, roundP, nil⟩ while stepP = prevote do
// 45: broadcast ⟨PRECOMMIT, hp, roundP, nil⟩
// 46: stepP ← precommit

// - A validator has received 2F+1 PREVOTE messages with any kind of ID (valid / NIL)
// 34: upon 2f + 1 <PREVOTE, hp, roundP, ∗> while stepP = prevote for the first time do
// 35: schedule OnTimeoutPrevote(hP , roundP) to be executed after timeoutPrevote(roundP)
func (t *Tendermint) runPrevote(ctx context.Context) {
	var (
		height             = t.state.getHeight()
		round              = t.state.getRound()
		acceptedProposalID = t.state.acceptedProposalID

		expiredCh                   = make(chan struct{}, 1)
		timeoutCtx, cancelTimeoutFn = context.WithCancel(ctx)
		timeoutPrevote              = t.timeouts[prevote].CalculateTimeout(round)
	)

	t.logger.Debug(
		"entering prevote state",
		slog.Uint64("height", height),
		slog.Uint64("round", round),
		slog.String("node", string(t.node.ID())),
	)

	// Defer the timeout timer cancellation
	defer cancelTimeoutFn()

	// Subscribe to all prevote messages
	// (=current height; unique; >= current round)
	ch, unsubscribeFn := t.store.subscribeToPrevote()
	defer unsubscribeFn()

	var (
		isValidFn = func(prevote *types.PrevoteMessage) bool {
			// Make sure the prevote view matches the process view
			return round == prevote.GetView().GetRound()
		}
		nilMiddleware = func(prevote *types.PrevoteMessage) bool {
			// Make sure the ID is NIL
			return prevote.Identifier == nil
		}
		matchingIDMiddleware = func(prevote *types.PrevoteMessage) bool {
			// Make sure the ID matches the accepted proposal ID
			return bytes.Equal(acceptedProposalID, prevote.Identifier)
		}
	)

	var (
		summedPrevotes = newMessageCache[*types.PrevoteMessage](isValidFn)
		nilCache       = newMessageCache[*types.PrevoteMessage](nilMiddleware)
		nonNilCache    = newMessageCache[*types.PrevoteMessage](matchingIDMiddleware)

		timeoutScheduled = false
	)

	for {
		select {
		case <-ctx.Done():
			return
		case <-expiredCh:
			// Build and broadcast the prevote message, with an ID of NIL
			t.broadcast.BroadcastPrecommit(t.buildPrecommitMessage(nil))

			t.state.step.set(precommit)

			return
		case getMessagesFn := <-ch:
			// Combine the prevote messages (NIL and non-NIL)
			summedPrevotes.addMessages(getMessagesFn())
			prevotes := summedPrevotes.getMessages()

			convertedMessages := make([]Message, 0, len(prevotes))
			messages.ConvertToInterface(
				prevotes,
				func(m *types.PrevoteMessage) {
					convertedMessages = append(convertedMessages, m)
				},
			)

			// Check if there is a super majority for the sum prevotes, to schedule a timeout
			if !timeoutScheduled && t.hasSuperMajority(convertedMessages) {
				// 35: schedule OnTimeoutPrevote(hp, roundP) to be executed after timeoutPrevote(roundP)
				t.logger.Debug(
					"scheduling timeoutPrevote",
					slog.Uint64("round", round),
					slog.Duration("timeout", timeoutPrevote),
					slog.String("node", string(t.node.ID())),
				)

				t.scheduleTimeout(timeoutCtx, timeoutPrevote, expiredCh)

				timeoutScheduled = true
			}

			// Filter the NIL prevote messages
			nilCache.addMessages(prevotes)
			nilPrevotes := nilCache.getMessages()

			convertedMessages = make([]Message, 0, len(nilPrevotes))
			messages.ConvertToInterface(
				nilPrevotes,
				func(m *types.PrevoteMessage) {
					convertedMessages = append(convertedMessages, m)
				},
			)

			// Check if there are 2F+1 NIL prevote messages
			if t.hasSuperMajority(convertedMessages) {
				// 45: broadcast ⟨PRECOMMIT, hp, roundP, nil⟩
				// 46: stepP ← precommit
				t.broadcast.BroadcastPrecommit(t.buildPrecommitMessage(nil))
				t.state.step.set(precommit)

				return
			}

			// Filter the non-NIL prevote messages
			nonNilCache.addMessages(prevotes)
			nonNilPrevotes := nonNilCache.getMessages()

			convertedMessages = make([]Message, 0, len(nonNilPrevotes))
			messages.ConvertToInterface(
				nonNilPrevotes,
				func(m *types.PrevoteMessage) {
					convertedMessages = append(convertedMessages, m)
				},
			)

			// Check if there are 2F+1 non-NIL prevote messages
			if t.hasSuperMajority(convertedMessages) {
				// 38: 	lockedValueP ← v
				// 39: 	lockedRoundP ← roundP
				t.state.lockedRound = int64(round)
				t.state.lockedValue = t.state.acceptedProposal

				// 40: 	broadcast <PRECOMMIT, hP, roundP, id(v))>
				t.broadcast.BroadcastPrecommit(t.buildPrecommitMessage(acceptedProposalID))

				// 41: 	stepP ← precommit
				t.state.step.set(precommit)

				// 42: validValueP ← v
				// 43: validRoundP ← roundP
				t.state.validValue = t.state.acceptedProposal
				t.state.validRound = int64(round)

				return
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
func (t *Tendermint) runPrecommit(ctx context.Context) *FinalizedProposal {
	var (
		height             = t.state.getHeight()
		round              = t.state.getRound()
		acceptedProposalID = t.state.acceptedProposalID

		expiredCh                   = make(chan struct{}, 1)
		timeoutCtx, cancelTimeoutFn = context.WithCancel(ctx)
		timeoutPrecommit            = t.timeouts[precommit].CalculateTimeout(round)
	)

	t.logger.Debug(
		"entering precommit state",
		slog.Uint64("height", height),
		slog.Uint64("round", round),
		slog.String("node", string(t.node.ID())),
	)

	// Defer the timeout timer cancellation
	defer cancelTimeoutFn()

	// Subscribe to all precommit messages
	// (=current height; unique; >= current round)
	ch, unsubscribeFn := t.store.subscribeToPrecommit()
	defer unsubscribeFn()

	var (
		isValidFn = func(precommit *types.PrecommitMessage) bool {
			// Make sure the precommit view matches the process view
			return round == precommit.GetView().GetRound()
		}
		nonNilIDFn = func(precommit *types.PrecommitMessage) bool {
			// Make sure the precommit ID is not nil
			if precommit.Identifier == nil {
				return false
			}

			// Make sure the ID matches the accepted proposal ID
			return bytes.Equal(acceptedProposalID, precommit.Identifier)
		}
	)

	var (
		summedPrecommits = newMessageCache[*types.PrecommitMessage](isValidFn)
		nonNilCache      = newMessageCache[*types.PrecommitMessage](nonNilIDFn)

		timeoutScheduled = false
	)

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, no proposal is finalized
			return nil
		case <-expiredCh:
			// Timeout triggered, no proposal is finalized
			return nil
		case getMessagesFn := <-ch:
			// Combine the precommit messages (NIL and non-NIL)
			summedPrecommits.addMessages(getMessagesFn())
			precommits := summedPrecommits.getMessages()

			convertedMessages := make([]Message, 0, len(precommits))
			messages.ConvertToInterface(
				precommits,
				func(m *types.PrecommitMessage) {
					convertedMessages = append(convertedMessages, m)
				},
			)

			// Check if there is a super majority for the sum precommits, to schedule a timeout
			if !timeoutScheduled && t.hasSuperMajority(convertedMessages) {
				// 48: schedule OnTimeoutPrecommit(hP, roundP) to be executed after timeoutPrecommit(roundP)
				t.logger.Debug(
					"scheduling timeoutPrecommit",
					slog.Uint64("round", round),
					slog.Duration("timeout", timeoutPrecommit),
					slog.String("node", string(t.node.ID())),
				)

				t.scheduleTimeout(timeoutCtx, timeoutPrecommit, expiredCh)

				timeoutScheduled = true
			}

			// Filter the non-NIL precommit messages
			nonNilCache.addMessages(precommits)
			nonNilPrecommits := nonNilCache.getMessages()

			convertedMessages = make([]Message, 0, len(nonNilPrecommits))
			messages.ConvertToInterface(
				nonNilPrecommits,
				func(m *types.PrecommitMessage) {
					convertedMessages = append(convertedMessages, m)
				},
			)

			// Check if there are 2F+1 non-NIL precommit messages
			if t.hasSuperMajority(convertedMessages) {
				return newFinalizedProposal(
					t.state.acceptedProposal,
					t.state.acceptedProposalID,
				)
			}
		}
	}
}
