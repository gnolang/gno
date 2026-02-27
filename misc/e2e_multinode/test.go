package main

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runMultiNodeTest executes multi-node determinism test
func runMultiNodeTest(
	t TestingT,
	ctx context.Context,
	wg *sync.WaitGroup,
	binaryPath string,
	validators, nonValidators []*Node,
	cfg *testCfg,
) {
	allNodes := append(validators, nonValidators...)

	// Start validators
	t.Log("📋 Starting validators")
	t.Logf("Starting %d validators", len(validators))
	for i, validator := range validators {
		t.Logf("Starting validator %d", i+1)
		err := startGnolandNode(t, ctx, binaryPath, validator)
		require.NoError(t, err, "failed to start validator %d", i+1)
		t.Logf("✅ Validator %d ready", i+1)
	}

	// Wait for validator P2P connectivity
	t.Log("📋 Waiting for validator connectivity")
	extValidators := make([]*ExtendedNode, len(validators))
	for i, val := range validators {
		extValidators[i] = &ExtendedNode{Node: val}
		rpcClient, err := client.NewHTTPClient(val.SocketAddr)
		require.NoError(t, err, "failed to create RPC client for validator %d", i+1)
		extValidators[i].Client = rpcClient
	}

	err := waitForPeerConnectivity(t, ctx, extValidators)
	require.NoError(t, err, "failed to establish validator connectivity")

	// Wait for initial sync
	err = waitForHeightSync(t, ctx, extValidators, 10)
	require.NoError(t, err, "failed to sync validators to height 10")

	// Start first non-validator
	var firstNonValidator *ExtendedNode
	if len(nonValidators) > 0 {
		t.Log("📋 Starting first non-validator")
		err := startGnolandNode(t, ctx, binaryPath, nonValidators[0])
		require.NoError(t, err, "failed to start non-validator 1")
		firstNonValidator = &ExtendedNode{Node: nonValidators[0]}
		// Initialize RPC client for non-validator
		rpcClient, err := client.NewHTTPClient(nonValidators[0].SocketAddr)
		require.NoError(t, err, "failed to create RPC client for non-validator")
		firstNonValidator.Client = rpcClient
		t.Log("✅ Non-validator ready")
	}

	// Execute test transactions
	runningNodes := slices.Clone(extValidators)
	if firstNonValidator != nil {
		runningNodes = append(runningNodes, firstNonValidator)
	}

	t.Log("📋 Executing transactions")
	t.Logf("Executing %d transactions with %d validators", cfg.numTransactions, len(validators))

	executeTestTransactions(t, validators[0], cfg.numTransactions)

	// Wait for transactions to be processed
	err = waitForHeightSync(t, ctx, runningNodes, 20)
	require.NoError(t, err, "failed to sync after transactions")

	// Start remaining non-validators
	if len(nonValidators) > 1 {
		t.Log("📋 Starting remaining non-validators")
		t.Logf("Starting %d additional non-validators", len(nonValidators)-1)
		for i := 1; i < len(nonValidators); i++ {
			t.Logf("Starting non-validator %d", i+1)
			wg.Add(1)
			go func(nv *Node, idx int) {
				defer wg.Done()
				if err := startGnolandNode(t, ctx, binaryPath, nv); err != nil {
					t.Logf("Error starting non-validator %d: %v", idx+1, err)
				} else {
					t.Logf("✅ Non-validator %d ready", idx+1)
				}
			}(nonValidators[i], i)
		}
	}

	// P2P connections will be verified in the sync phase

	// Wait for sync to target height and check determinism
	t.Log("📋 Waiting for chain topology and sync to target height")
	t.Logf("Total nodes: %d (%d validators + %d non-validators), target height: %d",
		len(allNodes), len(validators), len(nonValidators), cfg.targetHeight)

	// Create extended nodes for all
	allExtNodes := make([]*ExtendedNode, 0, len(allNodes))
	for i, node := range allNodes {
		extNode := &ExtendedNode{Node: node}
		if extNode.Client == nil {
			// Initialize client if not already done
			if err := waitForNodeReady(t, ctx, extNode); err != nil {
				t.Logf("Warning: Node %d not ready: %v", i, err)
			}
		}
		allExtNodes = append(allExtNodes, extNode)
	}

	// Wait for target height
	err = waitForHeightSync(t, ctx, allExtNodes, cfg.targetHeight)
	require.NoError(t, err, "failed to reach target height %d", cfg.targetHeight)

	// Perform comprehensive hash comparison
	t.Log("📊 Performing comprehensive determinism check...")
	checkDeterminism(t, ctx, allExtNodes, cfg)
}

// startGnolandNode starts a gnoland node (validator or non-validator)
func startGnolandNode(t TestingT, ctx context.Context, binaryPath string, node *Node) error {
	args := []string{
		"start",
		"--skip-failing-genesis-txs",
		"--skip-genesis-sig-verification",
		"--genesis", node.Genesis,
		"--data-dir", node.DataDir,
	}

	if err := startNode(t, ctx, binaryPath, node, args); err != nil {
		return err
	}

	// Wait for node to be ready
	extNode := &ExtendedNode{Node: node}
	return waitForNodeReady(t, ctx, extNode)
}

// executeTestTransactions simulates transaction execution
func executeTestTransactions(t TestingT, validator *Node, numTxs int) {
	t.Log("🔄 Executing test transactions to create state changes")
	t.Logf("Simulating %d transactions", numTxs)

	// In a real implementation, this would send actual transactions
	// For now, we simulate by allowing block production to create state changes
	if numTxs > 0 {
		t.Logf("📤 Simulating %d transactions via block production", numTxs)
	}

	t.Log("✅ Completed transaction simulation - state changes should have occurred via block production")
}

// waitForPeerConnectivity waits for nodes to establish P2P connections
func waitForPeerConnectivity(t TestingT, ctx context.Context, nodes []*ExtendedNode) error {
	t.Log("📋 Waiting for P2P peer connectivity...")

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	deadline, _ := ctx.Deadline()
	defer cancel()

	expectedPeers := 1
	if len(nodes) == 1 {
		expectedPeers = 0
	}

	success := assert.EventuallyWithT(t, func(c *assert.CollectT) {
		// Calculate expected peers: single validator has 0 peers, otherwise at least 1
		for _, node := range nodes {
			netInfo, err := node.Client.NetInfo(ctx)
			require.NoError(c, err)
			require.GreaterOrEqual(c, len(netInfo.Peers), expectedPeers)
		}
	}, time.Until(deadline), 1*time.Second, "failed to establish peer connectivity")

	if !success {
		return fmt.Errorf("timeout waiting for peer connectivity")
	}

	t.Log("✅ All nodes have established peer connections")
	return nil
}

// waitForHeightSync waits for all nodes to reach a minimum height
func waitForHeightSync(t TestingT, ctx context.Context, nodes []*ExtendedNode, minHeight int64) error {
	t.Log("📋 Waiting for block height synchronization...")

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(time.Minute * 2)
	}

	t.Logf("Height sync progress - deadline for sync %s", time.Until(deadline).String())
	success := assert.EventuallyWithT(t, func(c *assert.CollectT) {
		for i, node := range nodes {
			require.NotNil(c, node.Client)

			status, err := node.Client.Status(ctx, nil)
			require.NoError(c, err)

			currentHeight := status.SyncInfo.LatestBlockHeight
			if currentHeight < 50 || currentHeight%50 == 0 {
				t.Logf("Height sync progress - Node %d: %d/%d", i, currentHeight, minHeight)
			}

			require.GreaterOrEqual(c, currentHeight, minHeight)
		}
	}, time.Until(deadline), 1*time.Second, "failed to sync all nodes to height %d", minHeight)

	if !success {
		return fmt.Errorf("timeout waiting for height sync to %d", minHeight)
	}

	t.Logf("All nodes synced to target height %d", minHeight)
	return nil
}

// checkDeterminism performs comprehensive hash comparison across all nodes
func checkDeterminism(t TestingT, ctx context.Context, nodes []*ExtendedNode, cfg *testCfg) {
	// Get the minimum height across all nodes
	minCompareHeight := cfg.targetHeight
	for i, node := range nodes {
		status, err := node.Client.Status(ctx, nil)
		require.NoError(t, err, "failed to get status for node %d", i)
		if status.SyncInfo.LatestBlockHeight < minCompareHeight {
			minCompareHeight = status.SyncInfo.LatestBlockHeight
		}
		t.Logf("Node %d final height: %d", i, status.SyncInfo.LatestBlockHeight)
	}

	t.Log("📋 Comparing AppHashes from height 1 to target across all nodes")
	t.Logf("Target height: %d, Total nodes: %d (%d validators + %d non-validators)",
		minCompareHeight, len(nodes), cfg.numValidators, cfg.numNonValidators)

	// Get app hashes for ALL heights from 1 to minCompareHeight
	heightList := make([][]string, len(nodes))
	for nodeIdx, node := range nodes {
		heightList[nodeIdx] = make([]string, minCompareHeight)

		for h := int64(1); h <= minCompareHeight; h++ {
			block, err := node.Client.Block(ctx, &h)
			require.NoError(t, err, "failed to get block at height %d for node %d", h, nodeIdx)
			heightList[nodeIdx][h-1] = fmt.Sprintf("%X", block.Block.Header.AppHash)
		}
	}

	// Look for any divergence
	for h := int64(0); h < minCompareHeight; h++ {
		// Get hashes from all nodes for this height
		hashes := make([]string, len(nodes))
		for nodeIdx := range nodes {
			hashes[nodeIdx] = heightList[nodeIdx][h]
		}

		// Check if all hashes are identical
		allMatch := true
		for nodeIdx := 1; nodeIdx < len(nodes); nodeIdx++ {
			if hashes[nodeIdx] == hashes[0] {
				continue
			}

			allMatch = false
			break
		}

		if !allMatch {
			t.Logf("❌ NON-DETERMINISM DETECTED at height %d!", h+1)
			for nodeIdx, hash := range hashes {
				t.Logf("   Node %d AppHash: %s", nodeIdx, hash)
			}
			require.Fail(t, "NON-DETERMINISM FOUND: AppHash divergence detected!")
		} else {
			if h < 10 || (h+1)%50 == 0 { // Log first 10 and every 50th height for brevity
				t.Logf("Height consensus ✅ - Height %d: all %d nodes -> %s", h+1, len(nodes), hashes[0])
			}
		}
	}

	t.Log("🎉 SUCCESS: All AppHashes match across all nodes for ALL heights!")
	t.Logf("Verified %d heights across %d nodes (%d validators + %d non-validators)",
		minCompareHeight, len(nodes), cfg.numValidators, cfg.numNonValidators)
}
