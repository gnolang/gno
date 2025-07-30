package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// runMultiNodeTest executes the multi-node determinism test
func runMultiNodeTest(
	ctx context.Context,
	wg *sync.WaitGroup,
	binaryPath string,
	validators, nonValidators []*Node,
	cfg *testCfg,
) error {
	allNodes := append(validators, nonValidators...)

	// Step 1: Start all validators
	slog.Info("📋 Step 1: Starting validators", "count", len(validators))
	for i, validator := range validators {
		slog.Info("Starting validator", "index", i+1)
		if err := startValidatorNode(ctx, wg, binaryPath, validator); err != nil {
			return fmt.Errorf("failed to start validator %d: %w", i+1, err)
		}
		slog.Info("✅ Validator ready", "index", i+1)
	}

	// Step 2: Wait for validator connectivity
	slog.Info("📋 Step 2: Waiting for validator connectivity")
	extValidators := make([]*ExtendedNode, len(validators))
	for i, val := range validators {
		extValidators[i] = &ExtendedNode{Node: val}
	}

	if err := waitForPeerConnectivity(ctx, extValidators); err != nil {
		return fmt.Errorf("failed to establish validator connectivity: %w", err)
	}

	// Wait for initial sync
	if err := waitForHeightSync(ctx, extValidators, 10); err != nil {
		return fmt.Errorf("failed to sync validators to height 10: %w", err)
	}

	// Step 3: Start first non-validator
	var firstNonValidator *ExtendedNode
	if len(nonValidators) > 0 {
		slog.Info("📋 Step 3: Starting first non-validator")
		if err := startNonValidatorNode(ctx, wg, binaryPath, nonValidators[0]); err != nil {
			return fmt.Errorf("failed to start non-validator 1: %w", err)
		}
		firstNonValidator = &ExtendedNode{Node: nonValidators[0]}
		slog.Info("✅ Non-validator ready", "index", 1)
	}

	// Step 4: Execute test transactions
	runningNodes := append(extValidators, firstNonValidator)
	if firstNonValidator == nil {
		runningNodes = extValidators
	}

	slog.Info("📋 Step 4: Executing transactions", "count", cfg.numTransactions, "validators", len(validators))

	executeTestTransactions(validators[0], cfg.numTransactions)

	// Wait for transactions to be processed
	if err := waitForHeightSync(ctx, runningNodes, 20); err != nil {
		return fmt.Errorf("failed to sync after transactions: %w", err)
	}

	// Step 5: Start remaining non-validators
	if len(nonValidators) > 1 {
		slog.Info("📋 Step 5: Starting remaining non-validators", "count", len(nonValidators)-1)
		for i := 1; i < len(nonValidators); i++ {
			slog.Info("Starting non-validator", "index", i+1)
			wg.Add(1)
			go func(nv *Node, idx int) {
				defer wg.Done()
				if err := startNonValidatorNode(ctx, nil, binaryPath, nv); err != nil {
					slog.Error("Error starting non-validator", "index", idx+1, "error", err)
				} else {
					slog.Info("✅ Non-validator ready", "index", idx+1)
				}
			}(nonValidators[i], i)

			time.Sleep(2 * time.Second) // Stagger node starts
		}
	}

	// Wait for P2P connections
	slog.Info("📋 Waiting for P2P connections to be established...")
	time.Sleep(10 * time.Second)

	// Step 6: Wait for sync to target height and check determinism
	slog.Info("📋 Step 6: Waiting for chain topology (%d nodes: %d validators + %d non-validators) and sync to height ~%d...",
		len(allNodes), len(validators), len(nonValidators), cfg.targetHeight)

	// Create extended nodes for all
	allExtNodes := make([]*ExtendedNode, len(allNodes))
	for i, node := range allNodes {
		allExtNodes[i] = &ExtendedNode{Node: node}
		if allExtNodes[i].Client == nil {
			// Initialize client if not already done
			if err := waitForNodeReady(ctx, allExtNodes[i]); err != nil {
				slog.Info("Warning: Node %d not ready: %v", i, err)
			}
		}
	}

	// Wait for target height
	if err := waitForHeightSync(ctx, allExtNodes, cfg.targetHeight); err != nil {
		return fmt.Errorf("failed to reach target height %d: %w", cfg.targetHeight, err)
	}

	// Perform comprehensive hash comparison
	slog.Info("📊 Performing comprehensive determinism check...")
	if err := checkDeterminism(ctx, allExtNodes, cfg); err != nil {
		return fmt.Errorf("determinism check failed: %w", err)
	}

	return nil
}

// startValidatorNode starts a validator node
func startValidatorNode(ctx context.Context, wg *sync.WaitGroup, binaryPath string, node *Node) error {
	args := []string{
		"start",
		"--genesis", node.Genesis,
		"--skip-start",
		"--data-dir", node.DataDir,
	}

	if err := startNode(ctx, binaryPath, node, args); err != nil {
		return err
	}

	// Wait for node to be ready
	extNode := &ExtendedNode{Node: node}
	return waitForNodeReady(ctx, extNode)
}

// startNonValidatorNode starts a non-validator node
func startNonValidatorNode(ctx context.Context, wg *sync.WaitGroup, binaryPath string, node *Node) error {
	args := []string{
		"start",
		"--genesis", node.Genesis,
		"--skip-start",
		"--data-dir", node.DataDir,
	}

	if err := startNode(ctx, binaryPath, node, args); err != nil {
		return err
	}

	// Wait for node to be ready
	extNode := &ExtendedNode{Node: node}
	return waitForNodeReady(ctx, extNode)
}

// executeTestTransactions simulates transaction execution
func executeTestTransactions(validator *Node, numTxs int) {
	slog.Info("🔄 Executing %d test transactions to create state changes...", numTxs)

	// In a real implementation, this would send actual transactions
	// For now, we simulate by waiting for block production
	for i := 1; i <= numTxs; i++ {
		slog.Info("   📤 Simulating transaction %d (waiting for block production)...", i)
		time.Sleep(2 * time.Second)
	}

	slog.Info("✅ Completed transaction simulation - state changes should have occurred via block production")
}

// waitForPeerConnectivity waits for nodes to establish P2P connections
func waitForPeerConnectivity(ctx context.Context, nodes []*ExtendedNode) error {
	slog.Info("📋 Waiting for P2P peer connectivity...")

	maxAttempts := 30
	for attempt := 0; attempt < maxAttempts; attempt++ {
		allConnected := true

		for i, node := range nodes {
			netInfo, err := node.Client.NetInfo()
			if err != nil {
				allConnected = false
				continue
			}

			expectedPeers := 1 // Each validator should have at least 1 peer
			if len(netInfo.Peers) < expectedPeers {
				allConnected = false
				// Note: verbose logging removed for now - could be passed in cfg if needed
				slog.Info("Node %d has %d peers (expected: %d)", i, len(netInfo.Peers), expectedPeers)
			}
		}

		if allConnected {
			slog.Info("✅ All nodes have established peer connections")
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for peer connectivity")
		case <-time.After(1 * time.Second):
			// Continue waiting
		}
	}

	return fmt.Errorf("timeout waiting for peer connectivity")
}

// waitForHeightSync waits for all nodes to reach a minimum height
func waitForHeightSync(ctx context.Context, nodes []*ExtendedNode, minHeight int64) error {
	slog.Info("📋 Waiting for block height synchronization...")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for height sync")
		case <-ticker.C:
			allSynced := true

			for i, node := range nodes {
				if node.Client == nil {
					allSynced = false
					continue
				}

				status, err := node.Client.Status()
				if err != nil {
					allSynced = false
					continue
				}

				currentHeight := status.SyncInfo.LatestBlockHeight
				if currentHeight < minHeight {
					allSynced = false
					if i == 0 { // Only log for first node to reduce noise
						slog.Info("%d: current height [%d] vs %d", i, currentHeight, minHeight)
					}
				}
			}

			if allSynced {
				slog.Info("node[0] - configured to min height %d", minHeight)
				return nil
			}
		}
	}
}

// checkDeterminism performs comprehensive hash comparison across all nodes
func checkDeterminism(ctx context.Context, nodes []*ExtendedNode, cfg *testCfg) error {
	// Get the minimum height across all nodes
	minCompareHeight := cfg.targetHeight
	for i, node := range nodes {
		status, err := node.Client.Status()
		if err != nil {
			return fmt.Errorf("failed to get status for node %d: %w", i, err)
		}
		if status.SyncInfo.LatestBlockHeight < minCompareHeight {
			minCompareHeight = status.SyncInfo.LatestBlockHeight
		}
		slog.Info("   Node %d final height: %d", i, status.SyncInfo.LatestBlockHeight)
	}

	slog.Info("📋 Comparing EVERY SINGLE AppHash from height 1 to %d across all %d nodes (%d validators + %d non-validators)...",
		minCompareHeight, len(nodes), cfg.numValidators, cfg.numNonValidators)

	// Get app hashes for ALL heights from 1 to minCompareHeight
	heightList := make([][]string, len(nodes))
	for nodeIdx, node := range nodes {
		heightList[nodeIdx] = make([]string, minCompareHeight)

		for h := int64(1); h <= minCompareHeight; h++ {
			block, err := node.Client.Block(&h)
			if err != nil {
				return fmt.Errorf("failed to get block at height %d for node %d: %w", h, nodeIdx, err)
			}
			heightList[nodeIdx][h-1] = fmt.Sprintf("%X", block.Block.Header.AppHash)
		}
	}

	// Look for any divergence at ANY height
	divergenceFound := false
	for h := int64(0); h < minCompareHeight; h++ {
		// Get hashes from all nodes for this height
		hashes := make([]string, len(nodes))
		for nodeIdx := range nodes {
			hashes[nodeIdx] = heightList[nodeIdx][h]
		}

		// Check if all hashes are identical
		allMatch := true
		for nodeIdx := 1; nodeIdx < len(nodes); nodeIdx++ {
			if hashes[nodeIdx] != hashes[0] {
				allMatch = false
				break
			}
		}

		if !allMatch {
			slog.Info("❌ NON-DETERMINISM DETECTED at height %d!", h+1)
			for nodeIdx, hash := range hashes {
				slog.Info("   Node %d AppHash: %s", nodeIdx, hash)
			}
			divergenceFound = true
			break
		} else {
			if h < 10 || h%50 == 0 { // Log first 10 and every 50th height for brevity
				slog.Info("H[%d] all %d nodes -> %s ✅", h+1, len(nodes), hashes[0])
			}
		}
	}

	if !divergenceFound {
		slog.Info("🎉 PERFECT DETERMINISM: All AppHashes match across all %d nodes (%d validators + %d non-validators) for ALL %d heights!",
			len(nodes), cfg.numValidators, cfg.numNonValidators, minCompareHeight)
	} else {
		return fmt.Errorf("💥 NON-DETERMINISM FOUND: AppHash divergence detected!")
	}

	return nil
}
