package fork

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/commands"
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
