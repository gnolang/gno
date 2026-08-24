package integration

// Tests for the malformed amino type_url attack path: a transaction whose
// type_url field contains no forward slash used to trigger a hard panic in
// typeURLtoFullname() before the runTx recover block was ever entered.
//
// After the fix, typeURLtoFullname returns an error instead of panicking, and
// BaseApp.CheckTx / DeliverTx recover from any remaining codec panics. Both
// paths must return a tx-decode error rather than crashing.

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bfttypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// buildMalformedTxBytes encodes a valid bank.MsgSend transaction and flips the
// leading '/' (0x2F) in the amino type_url to '}' (0x7D). The result passes
// the IsASCIIText check in the binary decode path but contains no slash,
// which previously triggered a panic in typeURLtoFullname().
func buildMalformedTxBytes(t *testing.T) []byte {
	t.Helper()
	tx := std.Tx{
		Msgs: []std.Msg{
			bank.MsgSend{
				FromAddress: crypto.Address{},
				ToAddress:   crypto.Address{},
				Amount:      std.NewCoins(std.NewCoin("ugnot", 1)),
			},
		},
		Fee: std.NewFee(100000, std.NewCoin("ugnot", 1)),
	}
	validBz, err := amino.Marshal(tx)
	require.NoError(t, err)

	typeURL := amino.GetTypeURL(bank.MsgSend{})
	idx := bytes.Index(validBz, []byte(typeURL))
	require.True(t, idx >= 0, "type_url not found in binary payload")

	mutated := make([]byte, len(validBz))
	copy(mutated, validBz)
	mutated[idx] = '}' // '/' (0x2F) → '}' (0x7D): no slash, previously caused panic
	return mutated
}

// TestMalformedTypeURL_ConsensusDoesNotHalt verifies that a block containing a
// transaction with a malformed amino type_url (no slash) does not halt the
// consensus goroutine. The transaction must be rejected with a decode error and
// the node must continue processing subsequent blocks normally.
//
// Before the fix, the panic in typeURLtoFullname() propagated through:
//
//	BaseApp.DeliverTx → amino.Unmarshal → typeURLtoFullname (panic)
//	→ localClient.DeliverTxAsync (no recover)
//	→ execBlockOnProxyApp → ApplyBlock → finalizeCommit
//	→ receiveRoutine defer/recover → logs CONSENSUS FAILURE!!! → onExit()
//
// A single malicious proposer could deterministically halt every validator by
// including one such transaction in a block.
func TestMalformedTypeURL_ConsensusDoesNotHalt(t *testing.T) {
	t.Parallel()

	rootdir := gnoenv.RootDir()
	config := TestingMinimalNodeConfig(rootdir)
	// Disable empty blocks so the node stays in enterNewRound at heights > 1
	// with an empty mempool, giving a clean window to inject our proposal.
	config.TMConfig.Consensus.CreateEmptyBlocks = false

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	node, _ := TestingInMemoryNode(t, logger, config)
	defer node.Stop()

	mutatedBz := buildMalformedTxBytes(t)
	cs := node.ConsensusState()
	pv := node.PrivValidator()

	// consensusDead is closed by cs.Wait() if receiveRoutine ever exits — the
	// definitive signal that the consensus goroutine has terminated.
	consensusDead := make(chan struct{})
	go func() { cs.Wait(); close(consensusDead) }()

	// Wait for height ≥ 2 with no existing proposal. Height 1 always commits
	// automatically (needProofBlock(1)=true). At height 2+ with
	// CreateEmptyBlocks=false and an empty mempool, the node idles in
	// enterNewRound — cs.Proposal remains nil — providing a clean injection
	// window.
	var (
		targetHeight int64
		targetCommit *bfttypes.Commit
	)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		rs := cs.GetRoundState()
		if rs.Height < 2 || rs.Proposal != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		if c := cs.LoadCommit(rs.Height - 1); c != nil {
			targetHeight = rs.Height
			targetCommit = c
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.NotZero(t, targetHeight, "timed out waiting for injectable consensus height")

	state := cs.GetState()

	// Subscribe before injection so we don't miss the commit event.
	newBlockSub := events.SubscribeToEvent(node.EventSwitch(), "test-malformed-typeurl", bfttypes.EventNewBlock{})

	// Build a block that passes structural validation but contains the
	// amino-malformed transaction. state.MakeBlock uses MedianTime for
	// block.Time, which is exactly what ValidateBlock expects.
	proposerAddr := pv.PubKey().Address()
	block, blockParts := state.MakeBlock(
		targetHeight,
		[]bfttypes.Tx{bfttypes.Tx(mutatedBz)},
		targetCommit,
		proposerAddr,
	)

	proposal := bfttypes.NewProposal(
		targetHeight, 0, -1,
		bfttypes.BlockID{Hash: block.Hash(), PartsHeader: blockParts.Header()},
	)
	require.NoError(t, pv.SignProposal(state.ChainID, proposal))

	// Inject the signed proposal. receiveRoutine picks it up via peerMsgQueue.
	// With the fix applied, DeliverTx returns a decode error; the block is
	// committed with the tx marked as errored and consensus continues.
	require.NoError(t, cs.SetProposalAndBlock(proposal, block, blockParts, "attacker"))

	// Wait for an EventNewBlock at targetHeight: this fires only after the
	// block is fully committed, proving consensus is still alive.  If the fix
	// is absent, consensusDead closes first.
	select {
	case <-consensusDead:
		t.Fatal("consensus goroutine terminated — malformed tx caused an unrecovered panic")
	case ev := <-newBlockSub:
		got := ev.(bfttypes.EventNewBlock)
		require.GreaterOrEqual(t, got.Block.Height, targetHeight)
		t.Logf("OK: consensus survived, committed block %d with malformed type_url transaction", got.Block.Height)
	case <-time.After(15 * time.Second):
		t.Fatal("timed out waiting for consensus to commit block with malformed type_url transaction")
	}
}
