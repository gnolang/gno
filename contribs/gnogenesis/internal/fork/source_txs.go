package fork

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TxsSource provides historical transactions and the latest committed
// height of a source chain. Three implementations live alongside this
// file, picked via mutually-exclusive --source-txs-* flags:
//
//   - rpcTxsSource         (--source-txs-rpc <url>):       source_txs_rpc.go
//   - jsonlFileTxsSource   (--source-txs-jsonl-file PATH): source_txs_jsonl_file.go
//   - dataDirTxsSource     (--source-txs-data-dir DIR):    source_txs_data_dir.go
//
// Genesis is fetched separately via a GenesisSource (see source_genesis.go).
type TxsSource interface {
	// Description returns a human-readable source type label.
	Description() string

	// LatestHeight returns the latest committed block height known to
	// this source. Used to auto-detect halt height when --halt-height
	// is not specified.
	LatestHeight(ctx context.Context) (int64, error)

	// FetchTxs fetches all transactions in [fromHeight, toHeight] with
	// metadata (BlockHeight, Timestamp, ChainID populated). chainID is
	// supplied by the caller (sourced from the GenesisSource) so that
	// TxsSource implementations do not need to read genesis themselves.
	// Progress is reported via io.
	FetchTxs(ctx context.Context, chainID string, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error)

	// Close releases any resources held by the source.
	Close() error
}

// ---- shared sequence-resolution helpers (rpc + data-dir sources)

// signerState tracks per-signer sequence resolution during export.
type signerState struct {
	accNum       uint64
	finalSeq     uint64 // from source state at halt height
	seq          uint64 // current pre-tx sequence counter
	initialized  bool   // true after first brute-force resolves starting seq
	pendingFails []*pendingFailedTx
}

type pendingFailedTx struct {
	txIndex int // index in the output txs slice, for back-patching SignerInfo
	signerI int // index of this signer within the tx's signers
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
// Cosmetic/audit-only — failed txs are skipped during replay.
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

// ---- shared per-block tx processing pipeline

// txStream accumulates transactions across a block range and resolves
// per-signer sequences. Owned and called by individual TxsSource
// implementations (rpcTxsSource, dataDirTxsSource); the dedup lets the
// two sources differ only in how they obtain blocks/results and how they
// query account state.
//
// queryAccount is invoked lazily on the first appearance of each signer.
// The function should return nil for accounts not yet on-chain at the
// halt height; callers typically log a warning in that path.
type txStream struct {
	chainID      string
	queryAccount func(crypto.Address) std.Account
	io           commands.IO

	txs          []gnoland.TxWithMetadata
	signerStates map[crypto.Address]*signerState
}

func newTxStream(chainID string, queryAccount func(crypto.Address) std.Account, io commands.IO) *txStream {
	return &txStream{
		chainID:      chainID,
		queryAccount: queryAccount,
		io:           io,
		signerStates: map[crypto.Address]*signerState{},
	}
}

// getOrCreateSigner returns the cached signerState for addr, creating one
// (via queryAccount) on first access.
func (s *txStream) getOrCreateSigner(addr crypto.Address) *signerState {
	if ss, ok := s.signerStates[addr]; ok {
		return ss
	}
	acc := s.queryAccount(addr)
	ss := &signerState{}
	if acc != nil {
		ss.accNum = acc.GetAccountNumber()
		ss.finalSeq = acc.GetSequence()
	}
	s.signerStates[addr] = ss
	return ss
}

// processBlock decodes each tx in the block, resolves per-signer sequences,
// and appends the resulting TxWithMetadata entries to s.txs. Returns the
// number of txs appended (used by the caller for progress reporting).
//
// rawTxs and deliverTxs are parallel slices: deliverTxs[i].IsErr() decides
// whether tx i is treated as failed (sequence buffered for later
// back-patching) or successful (sequence brute-forced from the signature).
func (s *txStream) processBlock(h, timestamp int64, rawTxs []bft.Tx, deliverTxs []abci.ResponseDeliverTx) int {
	txCount := 0
	for i, rawTx := range rawTxs {
		var stdTx std.Tx
		if err := amino.Unmarshal(rawTx, &stdTx); err != nil {
			s.io.Printf("\n  WARNING: could not decode tx at height %d index %d: %v\n", h, i, err)
			continue
		}

		failed := false
		if i < len(deliverTxs) && deliverTxs[i].IsErr() {
			failed = true
		}

		signers := stdTx.GetSigners()
		sigs := stdTx.GetSignatures()
		txIdx := len(s.txs)

		signerInfos := make([]gnoland.SignerAccountInfo, len(signers))
		for j, signer := range signers {
			ss := s.getOrCreateSigner(signer)
			signerInfos[j] = gnoland.SignerAccountInfo{
				Address:    signer,
				AccountNum: ss.accNum,
				Sequence:   0,
			}
		}

		if !failed {
			for j, signer := range signers {
				ss := s.signerStates[signer]

				if !ss.initialized || len(ss.pendingFails) > 0 {
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
						stdTx, sig, ss.accNum, lo, hi, s.chainID)
					if err != nil {
						s.io.Printf("\n  WARNING: brute-force failed for signer %s at height %d: %v (using counter %d)\n",
							signer, h, err, ss.seq)
						resolvedSeq = ss.seq
					}

					assignFailedTxSequences(s.txs, ss.pendingFails, ss.seq, resolvedSeq)
					ss.pendingFails = nil
					ss.seq = resolvedSeq
					ss.initialized = true
				}

				signerInfos[j].Sequence = ss.seq
				ss.seq++
			}
		} else {
			for j, signer := range signers {
				ss := s.signerStates[signer]
				ss.pendingFails = append(ss.pendingFails, &pendingFailedTx{
					txIndex: txIdx,
					signerI: j,
				})
				signerInfos[j].Sequence = ss.seq
			}
		}

		s.txs = append(s.txs, gnoland.TxWithMetadata{
			Tx: stdTx,
			Metadata: &gnoland.GnoTxMetadata{
				Timestamp:   timestamp,
				BlockHeight: h,
				ChainID:     s.chainID,
				Failed:      failed,
				SignerInfo:  signerInfos,
			},
		})
		txCount++
	}
	return txCount
}

// resolveTrailingFailures back-patches sequence values on any signers that
// ended the stream with buffered failed txs and no later success to anchor
// against.
func (s *txStream) resolveTrailingFailures() {
	for _, ss := range s.signerStates {
		if len(ss.pendingFails) == 0 {
			continue
		}
		if !ss.initialized {
			var consumed uint64
			if ss.finalSeq > ss.seq {
				consumed = ss.finalSeq - ss.seq
			}
			if consumed > uint64(len(ss.pendingFails)) {
				ss.seq = ss.finalSeq - uint64(len(ss.pendingFails))
				consumed = uint64(len(ss.pendingFails))
			}
			assignTrailingFailedTxSequences(s.txs, ss.pendingFails, ss.seq, consumed)
		} else {
			var consumed uint64
			if ss.finalSeq > ss.seq {
				consumed = ss.finalSeq - ss.seq
			}
			assignTrailingFailedTxSequences(s.txs, ss.pendingFails, ss.seq, consumed)
		}
	}
}
