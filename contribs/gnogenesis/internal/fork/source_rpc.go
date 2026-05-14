package fork

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// rpcSource fetches chain state from a live (or recently-halted) node via RPC.
type rpcSource struct {
	rpcURL string
	client *rpcclient.RPCClient
}

func newRPCSource(rpcURL string) (*rpcSource, error) {
	client, err := rpcclient.NewHTTPClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("creating RPC client for %s: %w", rpcURL, err)
	}
	return &rpcSource{rpcURL: rpcURL, client: client}, nil
}

func (s *rpcSource) Description() string { return "RPC" }
func (s *rpcSource) Close() error        { return s.client.Close() }

func (s *rpcSource) FetchGenesis(ctx context.Context) (*bftypes.GenesisDoc, error) {
	res, err := s.client.Genesis(ctx)
	if err != nil {
		return nil, fmt.Errorf("RPC genesis call: %w", err)
	}
	return res.Genesis, nil
}

func (s *rpcSource) LatestHeight(ctx context.Context) (int64, error) {
	res, err := s.client.Status(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("RPC status call: %w", err)
	}
	return res.SyncInfo.LatestBlockHeight, nil
}

// signerState tracks per-signer sequence resolution during export.
type signerState struct {
	accNum       uint64
	finalSeq     uint64 // from RPC query at halt_height
	seq          uint64 // current pre-tx sequence counter
	initialized  bool   // true after first brute-force resolves starting seq
	pendingFails []*pendingFailedTx
}

type pendingFailedTx struct {
	txIndex int // index in the output txs slice, for back-patching SignerInfo
	signerI int // index of this signer within the tx's signers
}

// FetchTxs fetches all transactions in [fromHeight, toHeight] with metadata.
// Includes both successful and failed txs. Failed txs are marked with
// Failed: true and are not re-executed during replay, but their sequence
// impact is tracked.
func (s *rpcSource) FetchTxs(ctx context.Context, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error) {
	var txs []gnoland.TxWithMetadata

	// Get chain ID from genesis (needed for metadata)
	genesis, err := s.FetchGenesis(ctx)
	if err != nil {
		return nil, err
	}
	chainID := genesis.ChainID

	// Per-signer state for sequence tracking
	signerStates := map[crypto.Address]*signerState{}

	getOrCreateSignerState := func(addr crypto.Address) *signerState {
		if ss, ok := signerStates[addr]; ok {
			return ss
		}
		// Query account at halt_height
		acc := s.queryAccountAtHeight(ctx, addr, toHeight, io)
		ss := &signerState{}
		if acc != nil {
			ss.accNum = acc.GetAccountNumber()
			ss.finalSeq = acc.GetSequence()
		}
		signerStates[addr] = ss
		return ss
	}

	total := toHeight - fromHeight + 1
	var processed, txCount int64

	for h := fromHeight; h <= toHeight; h++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		processed++
		if processed%1000 == 0 || processed == total {
			io.Printf("\r  Blocks: %d/%d  Txs: %d", processed, total, txCount)
		}

		// Fetch block
		block, err := s.client.Block(ctx, &h)
		if err != nil {
			return nil, fmt.Errorf("fetching block %d: %w", h, err)
		}

		if len(block.Block.Data.Txs) == 0 {
			continue
		}

		// Fetch block results to check success/failure
		results, err := s.client.BlockResults(ctx, &h)
		if err != nil {
			return nil, fmt.Errorf("fetching block results %d: %w", h, err)
		}

		timestamp := block.Block.Header.Time.Unix()

		for i, rawTx := range block.Block.Data.Txs {
			// Decode the raw transaction bytes
			var stdTx std.Tx
			if err := amino.Unmarshal(rawTx, &stdTx); err != nil {
				io.Printf("\n  WARNING: could not decode tx at height %d index %d: %v\n", h, i, err)
				continue
			}

			failed := false
			if i < len(results.Results.DeliverTxs) && results.Results.DeliverTxs[i].IsErr() {
				failed = true
			}

			signers := stdTx.GetSigners()
			sigs := stdTx.GetSignatures()

			txIdx := len(txs) // index in output slice

			// Build signer info
			signerInfos := make([]gnoland.SignerAccountInfo, len(signers))
			for j, signer := range signers {
				ss := getOrCreateSignerState(signer)
				signerInfos[j] = gnoland.SignerAccountInfo{
					Address:    signer,
					AccountNum: ss.accNum,
					Sequence:   0, // filled below
				}
			}

			if !failed {
				// Successful tx: resolve sequences
				for j, signer := range signers {
					ss := signerStates[signer]

					if !ss.initialized || len(ss.pendingFails) > 0 {
						// Brute-force to find this tx's pre-tx sequence.
						lo := ss.seq
						hi := ss.seq + uint64(len(ss.pendingFails))
						if !ss.initialized {
							lo = 0
							hi = ss.finalSeq
						}

						var sig std.Signature
						if j < len(sigs) {
							sig = sigs[j]
						}

						resolvedSeq, err := bruteForceSignerSequence(
							stdTx, sig, ss.accNum, lo, hi, chainID)
						if err != nil {
							io.Printf("\n  WARNING: brute-force failed for signer %s at height %d: %v (using counter %d)\n",
								signer, h, err, ss.seq)
							resolvedSeq = ss.seq
						}

						// Back-patch buffered failed txs (cosmetic/audit-only)
						assignFailedTxSequences(txs, ss.pendingFails, ss.seq, resolvedSeq)
						ss.pendingFails = nil
						ss.seq = resolvedSeq
						ss.initialized = true
					}

					signerInfos[j].Sequence = ss.seq
					ss.seq++
				}
			} else {
				// Failed tx: buffer for each signer
				for j, signer := range signers {
					ss := signerStates[signer]
					ss.pendingFails = append(ss.pendingFails, &pendingFailedTx{
						txIndex: txIdx,
						signerI: j,
					})
					// Assign current counter as placeholder (will be back-patched)
					signerInfos[j].Sequence = ss.seq
				}
			}

			txs = append(txs, gnoland.TxWithMetadata{
				Tx: stdTx,
				Metadata: &gnoland.GnoTxMetadata{
					Timestamp:   timestamp,
					BlockHeight: h,
					ChainID:     chainID,
					Failed:      failed,
					SignerInfo:  signerInfos,
				},
			})
			txCount++
		}
	}

	// Resolve trailing failures
	for _, ss := range signerStates {
		if len(ss.pendingFails) == 0 {
			continue
		}

		if !ss.initialized {
			// Never had a successful tx. Cap consumed at len(pendingFails).
			var consumed uint64
			if ss.finalSeq > ss.seq {
				consumed = ss.finalSeq - ss.seq
			}
			if consumed > uint64(len(ss.pendingFails)) {
				ss.seq = ss.finalSeq - uint64(len(ss.pendingFails))
				consumed = uint64(len(ss.pendingFails))
			}
			assignTrailingFailedTxSequences(txs, ss.pendingFails, ss.seq, consumed)
		} else {
			var consumed uint64
			if ss.finalSeq > ss.seq {
				consumed = ss.finalSeq - ss.seq
			}
			assignTrailingFailedTxSequences(txs, ss.pendingFails, ss.seq, consumed)
		}
	}

	io.Printf("\r  Blocks: %d/%d  Txs: %d\n", processed, total, txCount)
	return txs, nil
}

// queryAccountAtHeight queries an account's state at a specific block height.
func (s *rpcSource) queryAccountAtHeight(
	ctx context.Context, addr crypto.Address, height int64, io commands.IO,
) std.Account {
	path := fmt.Sprintf("auth/accounts/%s", addr)
	res, err := s.client.ABCIQueryWithOptions(ctx, path, nil, rpcclient.ABCIQueryOptions{
		Height: height,
	})
	if err != nil {
		return nil
	}
	if res.Response.Error != nil {
		return nil
	}
	if len(res.Response.Data) == 0 {
		return nil
	}

	// Response data is amino JSON (the auth query handler returns JSON).
	// Try wrapped form first {"BaseAccount": {...}}, then direct.
	var wrapper struct {
		BaseAccount std.BaseAccount `json:"BaseAccount"`
	}
	if err := amino.UnmarshalJSON(res.Response.Data, &wrapper); err == nil {
		return &wrapper.BaseAccount
	}

	var acc std.BaseAccount
	if err := amino.UnmarshalJSON(res.Response.Data, &acc); err != nil {
		io.Printf("\n  WARNING: could not decode account %s at height %d: %v\n",
			addr, height, err)
		return nil
	}
	return &acc
}

// bruteForceSignerSequence tries sequences in [lo, hi] to find which makes
// the signature verify. Returns the pre-tx sequence (the value used in sign bytes).
func bruteForceSignerSequence(
	tx std.Tx, sig std.Signature, accNum uint64,
	lo, hi uint64, chainID string,
) (uint64, error) {
	pubKey := sig.PubKey
	if pubKey == nil {
		return lo, fmt.Errorf("no pubkey in signature")
	}

	for seq := lo; seq <= hi; seq++ {
		signBytes, err := std.GetSignaturePayload(std.SignDoc{
			ChainID:       chainID,
			AccountNumber: accNum,
			Sequence:      seq,
			Fee:           tx.Fee,
			Msgs:          tx.Msgs,
			Memo:          tx.Memo,
		})
		if err != nil {
			continue
		}
		if pubKey.VerifyBytes(signBytes, sig.Signature) {
			return seq, nil
		}
	}

	return lo, fmt.Errorf("no sequence in [%d, %d] verified for account %d", lo, hi, accNum)
}

// assignFailedTxSequences back-patches sequence values on buffered failed txs.
// This is cosmetic/audit-only — failed txs are skipped during replay and the
// replay loop does not depend on their SignerInfo.Sequence values.
//
// Ordering within the gap is ambiguous: we cannot determine whether a failed tx
// was ante-fail (no sequence consumed) or msg-fail (sequence consumed) without
// re-verifying its signature, which may not be possible if the pubkey was not
// on-chain yet. We approximate by assuming msg-fails (consuming) come first in
// the gap, then ante-fails.
func assignFailedTxSequences(
	txs []gnoland.TxWithMetadata,
	pending []*pendingFailedTx,
	startSeq, resolvedSeq uint64,
) {
	consumed := resolvedSeq - startSeq
	seq := startSeq
	for i, pf := range pending {
		if pf.txIndex < len(txs) && pf.signerI < len(txs[pf.txIndex].Metadata.SignerInfo) {
			txs[pf.txIndex].Metadata.SignerInfo[pf.signerI].Sequence = seq
		}
		if uint64(i) < consumed {
			seq++
		}
	}
}

// assignTrailingFailedTxSequences handles failed txs at the end of the chain
// with no subsequent success to anchor against.
func assignTrailingFailedTxSequences(
	txs []gnoland.TxWithMetadata,
	pending []*pendingFailedTx,
	startSeq, consumed uint64,
) {
	seq := startSeq
	for i, pf := range pending {
		if pf.txIndex < len(txs) && pf.signerI < len(txs[pf.txIndex].Metadata.SignerInfo) {
			txs[pf.txIndex].Metadata.SignerInfo[pf.signerI].Sequence = seq
		}
		if uint64(i) < consumed {
			seq++
		}
	}
}
