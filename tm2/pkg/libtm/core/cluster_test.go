package core

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/gnolang/libtm/messages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateNodeAddresses generates dummy node addresses
func generateNodeAddresses(count uint64) [][]byte {
	addresses := make([][]byte, count)

	for index := range addresses {
		addresses[index] = []byte(fmt.Sprintf("node %d", index))
	}

	return addresses
}

// TestConsensus_ValidFlow tests the following scenario:
// N = 4
//
// - Node 0 is the proposer for block 1, round 0
// - Node 0 proposes a valid block B
// - All nodes go through the consensus states to insert the valid block B
func TestConsensus_ValidFlow(t *testing.T) {
	t.Parallel()

	var (
		broadcastProposeFn   func(message *types.ProposalMessage)
		broadcastPrevoteFn   func(message *types.PrevoteMessage)
		broadcastPrecommitFn func(message *types.PrecommitMessage)

		proposal     = []byte("proposal")
		proposalHash = []byte("proposal hash")
		signature    = []byte("signature")
		numNodes     = uint64(4)
		nodes        = generateNodeAddresses(numNodes)

		defaultTimeout = Timeout{
			Initial: 2 * time.Second,
			Delta:   200 * time.Millisecond,
		}
	)

	// commonBroadcastCallback is the common method modification
	// required for Broadcast, for all nodes
	commonBroadcastCallback := func(broadcast *mockBroadcast) {
		broadcast.broadcastProposeFn = func(message *types.ProposalMessage) {
			broadcastProposeFn(message)
		}

		broadcast.broadcastPrevoteFn = func(message *types.PrevoteMessage) {
			broadcastPrevoteFn(message)
		}

		broadcast.broadcastPrecommitFn = func(message *types.PrecommitMessage) {
			broadcastPrecommitFn(message)
		}
	}

	// commonNodeCallback is the common method modification required
	// for the Node, for all nodes
	commonNodeCallback := func(node *mockNode, nodeIndex int) {
		node.idFn = func() []byte {
			return nodes[nodeIndex]
		}

		node.hashFn = func(_ []byte) []byte {
			return proposalHash
		}
	}

	// commonSignerCallback is the common method modification required
	// for the Signer, for all nodes
	commonSignerCallback := func(signer *mockSigner) {
		signer.signFn = func(_ []byte) []byte {
			return signature
		}

		signer.isValidSignatureFn = func(_, sig []byte) bool {
			return bytes.Equal(sig, signature)
		}
	}

	// commonVerifierCallback is the common method modification required
	// for the Verifier, for all nodes
	commonVerifierCallback := func(verifier *mockVerifier) {
		verifier.isProposerFn = func(from []byte, _ uint64, _ uint64) bool {
			return bytes.Equal(from, nodes[0])
		}

		verifier.isValidProposalFn = func(newProposal []byte, _ uint64) bool {
			return bytes.Equal(newProposal, proposal)
		}

		verifier.isValidatorFn = func(_ []byte) bool {
			return true
		}

		verifier.getTotalVotingPowerFn = func(_ uint64) uint64 {
			return numNodes
		}

		verifier.getSumVotingPowerFn = func(messages []Message) uint64 {
			return uint64(len(messages))
		}
	}

	var (
		verifierCallbackMap = map[int]verifierConfigCallback{
			0: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
			1: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
			2: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
			3: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
		}

		nodeCallbackMap = map[int]nodeConfigCallback{
			0: func(node *mockNode) {
				commonNodeCallback(node, 0)

				node.buildProposalFn = func(_ uint64) []byte {
					return proposal
				}
			},
			1: func(node *mockNode) {
				commonNodeCallback(node, 1)
			},
			2: func(node *mockNode) {
				commonNodeCallback(node, 2)
			},
			3: func(node *mockNode) {
				commonNodeCallback(node, 3)
			},
		}

		broadcastCallbackMap = map[int]broadcastConfigCallback{
			0: commonBroadcastCallback,
			1: commonBroadcastCallback,
			2: commonBroadcastCallback,
			3: commonBroadcastCallback,
		}

		signerCallbackMap = map[int]signerConfigCallback{
			0: commonSignerCallback,
			1: commonSignerCallback,
			2: commonSignerCallback,
			3: commonSignerCallback,
		}

		commonOptions = []Option{
			WithProposeTimeout(defaultTimeout),
			WithPrevoteTimeout(defaultTimeout),
			WithPrecommitTimeout(defaultTimeout),
		}

		optionsCallbackMap = map[int][]Option{
			0: commonOptions,
			1: commonOptions,
			2: commonOptions,
			3: commonOptions,
		}
	)

	// Create the mock cluster
	cluster := newMockCluster(
		numNodes,
		verifierCallbackMap,
		nodeCallbackMap,
		broadcastCallbackMap,
		signerCallbackMap,
		optionsCallbackMap,
	)

	broadcastProposeFn = func(message *types.ProposalMessage) {
		require.NoError(t, cluster.pushProposalMessage(message))
	}

	broadcastPrevoteFn = func(message *types.PrevoteMessage) {
		require.NoError(t, cluster.pushPrevoteMessage(message))
	}

	broadcastPrecommitFn = func(message *types.PrecommitMessage) {
		require.NoError(t, cluster.pushPrecommitMessage(message))
	}

	// Start the main run loops
	cluster.runSequence(0)

	// Wait until the main run loops finish
	cluster.ensureShutdown(5 * time.Second)

	// Make sure the finalized proposals match what node 0 proposed
	for _, finalizedProposal := range cluster.finalizedProposals {
		require.NotNil(t, finalizedProposal)

		assert.True(t, bytes.Equal(finalizedProposal.Data, proposal))
		assert.True(t, bytes.Equal(finalizedProposal.ID, proposalHash))
	}
}

// TestConsensus_InvalidBlock tests the following scenario:
// N = 4
//
// - Node 0 is the proposer for block 1, round 0
// - Node 0 proposes an invalid block B
// - Other nodes should verify that the block is invalid
// - All nodes should move to round 1, and start a new consensus round
// - Node 1 is the proposer for block 1, round 1
// - Node 1 proposes a valid block B'
// - All nodes go through the consensus states to insert the valid block B'
func TestConsensus_InvalidFlow(t *testing.T) {
	t.Parallel()

	var (
		broadcastProposeFn   func(message *types.ProposalMessage)
		broadcastPrevoteFn   func(message *types.PrevoteMessage)
		broadcastPrecommitFn func(message *types.PrecommitMessage)

		proposals = [][]byte{
			[]byte("proposal 1"), // proposed by node 0
			[]byte("proposal 2"), // proposed by node 1
		}

		proposalHashes = [][]byte{
			[]byte("proposal hash 1"), // for proposal 1
			[]byte("proposal hash 2"), // for proposal 2
		}

		signature = []byte("signature")
		numNodes  = uint64(4)
		nodes     = generateNodeAddresses(numNodes)

		defaultTimeout = Timeout{
			Initial: 2 * time.Second,
			Delta:   200 * time.Millisecond,
		}

		precommitTimeout = Timeout{
			Initial: 300 * time.Millisecond, // low timeout, so a new round is started quicker
			Delta:   200 * time.Millisecond,
		}
	)

	// commonBroadcastCallback is the common method modification
	// required for Broadcast, for all nodes
	commonBroadcastCallback := func(broadcast *mockBroadcast) {
		broadcast.broadcastProposeFn = func(message *types.ProposalMessage) {
			broadcastProposeFn(message)
		}

		broadcast.broadcastPrevoteFn = func(message *types.PrevoteMessage) {
			broadcastPrevoteFn(message)
		}

		broadcast.broadcastPrecommitFn = func(message *types.PrecommitMessage) {
			broadcastPrecommitFn(message)
		}
	}

	// commonNodeCallback is the common method modification required
	// for the Node, for all nodes
	commonNodeCallback := func(node *mockNode, nodeIndex int) {
		node.idFn = func() []byte {
			return nodes[nodeIndex]
		}

		node.hashFn = func(proposal []byte) []byte {
			if bytes.Equal(proposal, proposals[0]) {
				return proposalHashes[0]
			}

			return proposalHashes[1]
		}
	}

	// commonSignerCallback is the common method modification required
	// for the Signer, for all nodes
	commonSignerCallback := func(signer *mockSigner) {
		signer.signFn = func(_ []byte) []byte {
			return signature
		}

		signer.isValidSignatureFn = func(_, sig []byte) bool {
			return bytes.Equal(sig, signature)
		}
	}

	// commonVerifierCallback is the common method modification required
	// for the Verifier, for all nodes
	commonVerifierCallback := func(verifier *mockVerifier) {
		verifier.isProposerFn = func(from []byte, _ uint64, round uint64) bool {
			// Node 0 is the proposer for round 0
			// Node 1 is the proposer for round 1
			return bytes.Equal(from, nodes[round])
		}

		verifier.isValidProposalFn = func(newProposal []byte, _ uint64) bool {
			// Node 1 is the proposer for round 1,
			// and their proposal is the only one that's valid
			return bytes.Equal(newProposal, proposals[1])
		}

		verifier.isValidatorFn = func(_ []byte) bool {
			return true
		}

		verifier.getTotalVotingPowerFn = func(_ uint64) uint64 {
			return numNodes
		}

		verifier.getSumVotingPowerFn = func(messages []Message) uint64 {
			return uint64(len(messages))
		}
	}

	var (
		verifierCallbackMap = map[int]verifierConfigCallback{
			0: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
			1: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
			2: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
			3: func(verifier *mockVerifier) {
				commonVerifierCallback(verifier)
			},
		}

		nodeCallbackMap = map[int]nodeConfigCallback{
			0: func(node *mockNode) {
				commonNodeCallback(node, 0)

				node.buildProposalFn = func(_ uint64) []byte {
					return proposals[0]
				}
			},
			1: func(node *mockNode) {
				commonNodeCallback(node, 1)

				node.buildProposalFn = func(_ uint64) []byte {
					return proposals[1]
				}
			},
			2: func(node *mockNode) {
				commonNodeCallback(node, 2)
			},
			3: func(node *mockNode) {
				commonNodeCallback(node, 3)
			},
		}

		broadcastCallbackMap = map[int]broadcastConfigCallback{
			0: commonBroadcastCallback,
			1: commonBroadcastCallback,
			2: commonBroadcastCallback,
			3: commonBroadcastCallback,
		}

		signerCallbackMap = map[int]signerConfigCallback{
			0: commonSignerCallback,
			1: commonSignerCallback,
			2: commonSignerCallback,
			3: commonSignerCallback,
		}

		commonOptions = []Option{
			WithProposeTimeout(defaultTimeout),
			WithPrevoteTimeout(defaultTimeout),
			WithPrecommitTimeout(precommitTimeout),
		}

		optionsCallbackMap = map[int][]Option{
			0: commonOptions,
			1: commonOptions,
			2: commonOptions,
			3: commonOptions,
		}
	)

	// Create the mock cluster
	cluster := newMockCluster(
		numNodes,
		verifierCallbackMap,
		nodeCallbackMap,
		broadcastCallbackMap,
		signerCallbackMap,
		optionsCallbackMap,
	)

	broadcastProposeFn = func(message *types.ProposalMessage) {
		_ = cluster.pushProposalMessage(message) //nolint:errcheck // No need to check
	}

	broadcastPrevoteFn = func(message *types.PrevoteMessage) {
		_ = cluster.pushPrevoteMessage(message) //nolint:errcheck // No need to check
	}

	broadcastPrecommitFn = func(message *types.PrecommitMessage) {
		_ = cluster.pushPrecommitMessage(message) //nolint:errcheck // No need to check
	}

	// Start the main run loops
	cluster.runSequence(0)

	// Wait until the main run loops finish
	cluster.ensureShutdown(5 * time.Second)

	// Make sure the nodes switched to the new round
	assert.True(t, cluster.areAllNodesOnRound(1))

	// Make sure the finalized proposals match what node 0 proposed
	for _, finalizedProposal := range cluster.finalizedProposals {
		require.NotNil(t, finalizedProposal)

		assert.True(t, bytes.Equal(finalizedProposal.Data, proposals[1]))
		assert.True(t, bytes.Equal(finalizedProposal.ID, proposalHashes[1]))
	}
}
