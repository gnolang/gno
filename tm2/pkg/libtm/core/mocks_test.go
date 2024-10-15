package core

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/libtm/messages/types"
)

type (
	broadcastProposeDelegate   func(*types.ProposalMessage)
	broadcastPrevoteDelegate   func(*types.PrevoteMessage)
	broadcastPrecommitDelegate func(*types.PrecommitMessage)
)

type mockBroadcast struct {
	broadcastProposeFn   broadcastProposeDelegate
	broadcastPrevoteFn   broadcastPrevoteDelegate
	broadcastPrecommitFn broadcastPrecommitDelegate
}

func (m *mockBroadcast) BroadcastPropose(message *types.ProposalMessage) {
	if m.broadcastProposeFn != nil {
		m.broadcastProposeFn(message)
	}
}

func (m *mockBroadcast) BroadcastPrevote(message *types.PrevoteMessage) {
	if m.broadcastPrevoteFn != nil {
		m.broadcastPrevoteFn(message)
	}
}

func (m *mockBroadcast) BroadcastPrecommit(message *types.PrecommitMessage) {
	if m.broadcastPrecommitFn != nil {
		m.broadcastPrecommitFn(message)
	}
}

type (
	idDelegate            func() []byte
	hashDelegate          func([]byte) []byte
	buildProposalDelegate func(uint64) []byte
)

type mockNode struct {
	idFn            idDelegate
	hashFn          hashDelegate
	buildProposalFn buildProposalDelegate
}

func (m *mockNode) ID() []byte {
	if m.idFn != nil {
		return m.idFn()
	}

	return nil
}

func (m *mockNode) Hash(proposal []byte) []byte {
	if m.hashFn != nil {
		return m.hashFn(proposal)
	}

	return nil
}

func (m *mockNode) BuildProposal(height uint64) []byte {
	if m.buildProposalFn != nil {
		return m.buildProposalFn(height)
	}

	return nil
}

type (
	signDelegate             func([]byte) []byte
	isValidSignatureDelegate func([]byte, []byte) bool
)

type mockSigner struct {
	signFn             signDelegate
	isValidSignatureFn isValidSignatureDelegate
}

func (m *mockSigner) Sign(data []byte) []byte {
	if m.signFn != nil {
		return m.signFn(data)
	}

	return nil
}

func (m *mockSigner) IsValidSignature(data, signature []byte) bool {
	if m.isValidSignatureFn != nil {
		return m.isValidSignatureFn(data, signature)
	}

	return false
}

type (
	isProposerDelegate          func([]byte, uint64, uint64) bool
	isValidatorDelegate         func([]byte) bool
	isValidProposalDelegate     func([]byte, uint64) bool
	getTotalVotingPowerDelegate func(uint64) uint64
	getSumVotingPowerDelegate   func([]Message) uint64
)

type mockVerifier struct {
	isProposerFn          isProposerDelegate
	isValidatorFn         isValidatorDelegate
	isValidProposalFn     isValidProposalDelegate
	getTotalVotingPowerFn getTotalVotingPowerDelegate
	getSumVotingPowerFn   getSumVotingPowerDelegate
}

func (m *mockVerifier) GetTotalVotingPower(height uint64) uint64 {
	if m.getTotalVotingPowerFn != nil {
		return m.getTotalVotingPowerFn(height)
	}

	return 0
}

func (m *mockVerifier) GetSumVotingPower(msgs []Message) uint64 {
	if m.getSumVotingPowerFn != nil {
		return m.getSumVotingPowerFn(msgs)
	}

	return 0
}

func (m *mockVerifier) IsProposer(id []byte, height, round uint64) bool {
	if m.isProposerFn != nil {
		return m.isProposerFn(id, height, round)
	}

	return false
}

func (m *mockVerifier) IsValidator(from []byte) bool {
	if m.isValidatorFn != nil {
		return m.isValidatorFn(from)
	}

	return false
}

func (m *mockVerifier) IsValidProposal(proposal []byte, height uint64) bool {
	if m.isValidProposalFn != nil {
		return m.isValidProposalFn(proposal, height)
	}

	return false
}

type (
	getViewDelegate             func() *types.View
	getSenderDelegate           func() []byte
	getSignatureDelegate        func() []byte
	getSignaturePayloadDelegate func() []byte
	verifyDelegate              func() error
)

type mockMessage struct {
	getViewFn             getViewDelegate
	getSenderFn           getSenderDelegate
	getSignatureFn        getSignatureDelegate
	getSignaturePayloadFn getSignaturePayloadDelegate
	verifyFn              verifyDelegate
}

func (m *mockMessage) GetView() *types.View {
	if m.getViewFn != nil {
		return m.getViewFn()
	}

	return nil
}

func (m *mockMessage) GetSender() []byte {
	if m.getSenderFn != nil {
		return m.getSenderFn()
	}

	return nil
}

func (m *mockMessage) GetSignature() []byte {
	if m.getSignatureFn != nil {
		return m.getSignatureFn()
	}

	return nil
}

func (m *mockMessage) GetSignaturePayload() []byte {
	if m.getSignaturePayloadFn != nil {
		return m.getSignaturePayloadFn()
	}

	return nil
}

func (m *mockMessage) Verify() error {
	if m.verifyFn != nil {
		return m.verifyFn()
	}

	return nil
}

// mockNodeContext keeps track of the node runtime context
type mockNodeContext struct {
	ctx      context.Context
	cancelFn context.CancelFunc
}

// mockNodeWg is the WaitGroup wrapper for the cluster nodes
type mockNodeWg struct {
	sync.WaitGroup
	count int64
}

func (wg *mockNodeWg) Add(delta int) {
	wg.WaitGroup.Add(delta)
}

func (wg *mockNodeWg) Done() {
	wg.WaitGroup.Done()
	atomic.AddInt64(&wg.count, 1)
}

func (wg *mockNodeWg) getDone() int64 {
	return atomic.LoadInt64(&wg.count)
}

func (wg *mockNodeWg) resetDone() {
	atomic.StoreInt64(&wg.count, 0)
}

type (
	verifierConfigCallback  func(*mockVerifier)
	nodeConfigCallback      func(*mockNode)
	broadcastConfigCallback func(*mockBroadcast)
	signerConfigCallback    func(*mockSigner)
)

// mockCluster represents a mock Tendermint cluster
type mockCluster struct {
	nodes              []*Tendermint        // references to the nodes in the cluster
	ctxs               []mockNodeContext    // context handlers for the nodes in the cluster
	finalizedProposals []*FinalizedProposal // finalized proposals for the nodes

	stoppedWg mockNodeWg
}

// newMockCluster creates a new mock Tendermint cluster
func newMockCluster(
	count uint64,
	verifierCallbackMap map[int]verifierConfigCallback,
	nodeCallbackMap map[int]nodeConfigCallback,
	broadcastCallbackMap map[int]broadcastConfigCallback,
	signerCallbackMap map[int]signerConfigCallback,
	optionsMap map[int][]Option,
) *mockCluster {
	if count < 1 {
		return nil
	}

	nodes := make([]*Tendermint, count)
	nodeCtxs := make([]mockNodeContext, count)

	for index := 0; index < int(count); index++ {
		var (
			verifier  = &mockVerifier{}
			node      = &mockNode{}
			broadcast = &mockBroadcast{}
			signer    = &mockSigner{}
			options   = make([]Option, 0)
		)

		// Execute set callbacks, if any
		if verifierCallbackMap != nil {
			if verifierCallback, isSet := verifierCallbackMap[index]; isSet {
				verifierCallback(verifier)
			}
		}

		if nodeCallbackMap != nil {
			if nodeCallback, isSet := nodeCallbackMap[index]; isSet {
				nodeCallback(node)
			}
		}

		if broadcastCallbackMap != nil {
			if broadcastCallback, isSet := broadcastCallbackMap[index]; isSet {
				broadcastCallback(broadcast)
			}
		}

		if signerCallbackMap != nil {
			if signerCallback, isSet := signerCallbackMap[index]; isSet {
				signerCallback(signer)
			}
		}

		if optionsMap != nil {
			if opts, isSet := optionsMap[index]; isSet {
				options = opts
			}
		}

		// Create a new instance of the Tendermint node
		nodes[index] = NewTendermint(
			verifier,
			node,
			broadcast,
			signer,
			options...,
		)

		// Instantiate context for the nodes
		ctx, cancelFn := context.WithCancel(context.Background())
		nodeCtxs[index] = mockNodeContext{
			ctx:      ctx,
			cancelFn: cancelFn,
		}
	}

	return &mockCluster{
		nodes:              nodes,
		ctxs:               nodeCtxs,
		finalizedProposals: make([]*FinalizedProposal, count),
	}
}

// runSequence runs the cluster sequence for the given height
func (m *mockCluster) runSequence(height uint64) {
	m.stoppedWg.resetDone()

	for nodeIndex, node := range m.nodes {
		m.stoppedWg.Add(1)

		go func(
			ctx context.Context,
			node *Tendermint,
			nodeIndex int,
			height uint64,
		) {
			defer m.stoppedWg.Done()

			// Start the main run loop for the node
			finalizedProposal := node.RunSequence(ctx, height)

			m.finalizedProposals[nodeIndex] = finalizedProposal
		}(m.ctxs[nodeIndex].ctx, node, nodeIndex, height)
	}
}

// awaitCompletion waits for completion of all
// nodes in the cluster
func (m *mockCluster) awaitCompletion() {
	// Wait for all main run loops to signalize
	// that they're finished
	m.stoppedWg.Wait()
}

// ensureShutdown ensures the cluster is shutdown within the given duration
func (m *mockCluster) ensureShutdown(timeout time.Duration) {
	ch := time.After(timeout)

	for {
		select {
		case <-ch:
			m.forceShutdown()

			return
		default:
			if m.stoppedWg.getDone() == int64(len(m.nodes)) {
				// All nodes are finished
				return
			}
		}
	}
}

// forceShutdown sends a stop signal to all running nodes
// in the cluster, and awaits their completion
func (m *mockCluster) forceShutdown() {
	// Send a stop signal to all the nodes
	for _, ctx := range m.ctxs {
		ctx.cancelFn()
	}

	// Wait for all the nodes to finish
	m.awaitCompletion()
}

// pushProposalMessage relays the proposal message to all nodes in the cluster
func (m *mockCluster) pushProposalMessage(message *types.ProposalMessage) error {
	for _, node := range m.nodes {
		if err := node.AddProposalMessage(message); err != nil {
			return err
		}
	}

	return nil
}

// pushPrevoteMessage relays the prevote message to all nodes in the cluster
func (m *mockCluster) pushPrevoteMessage(message *types.PrevoteMessage) error {
	for _, node := range m.nodes {
		if err := node.AddPrevoteMessage(message); err != nil {
			return err
		}
	}

	return nil
}

// pushPrecommitMessage relays the precommit message to all nodes in the cluster
func (m *mockCluster) pushPrecommitMessage(message *types.PrecommitMessage) error {
	for _, node := range m.nodes {
		if err := node.AddPrecommitMessage(message); err != nil {
			return err
		}
	}

	return nil
}

// areAllNodesOnRound checks to make sure all nodes
// are on the same specified round
func (m *mockCluster) areAllNodesOnRound(round uint64) bool {
	for _, node := range m.nodes {
		if node.state.getRound() != round {
			return false
		}
	}

	return true
}
