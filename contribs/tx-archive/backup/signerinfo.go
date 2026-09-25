package backup

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/contribs/tx-archive/backup/client"
)

// signerResolver tracks per-signer account state during backup so that each
// exported tx carries a SignerInfo entry with the (account_num, sequence)
// values used to sign it on the source chain.
//
// Hardfork replay needs these values to force-set account state on the new
// chain before signature verification — see gno.land/pkg/gnoland.InitChainer
// loadAppState (PR #5511).
//
// Brute-force resolution strategy (ported from misc/hardfork/source_rpc.go):
//  1. On first sight of a signer, query auth/accounts/<addr> at the halt
//     height to learn the *final* (accNum, finalSeq).
//  2. On the signer's first *successful* tx in the stream, brute-force
//     sequences in [0, finalSeq] against the tx signature to find the
//     starting sequence at that point in history. Subsequent successful
//     txs simply increment a counter.
//  3. Failed txs buffer until the next success — if any, re-brute-force to
//     figure out how many of them actually consumed sequence (ante-fail =
//     no consume, msg-fail = consume). Trailing failed txs (no later
//     success to anchor) are handled in Finalize().
//
// Failed-tx sequence values are cosmetic — replay skips failed txs — so the
// resolver's fallbacks err on the side of "roughly right" rather than
// re-fetching more RPC state.
type signerResolver struct {
	client     client.Client
	chainID    string
	haltHeight uint64
	states     map[crypto.Address]*signerState
}

type signerState struct {
	accNum       uint64
	finalSeq     uint64 // from RPC query at halt_height
	seq          uint64 // current pre-tx counter
	initialized  bool   // true after first success brute-force resolves start
	pendingFails []*pendingFailedTx
}

type pendingFailedTx struct {
	info    *gnoland.SignerAccountInfo // direct pointer into tx.Metadata.SignerInfo
	ownerSS *signerState
}

func newSignerResolver(c client.Client, chainID string, haltHeight uint64) *signerResolver {
	return &signerResolver{
		client:     c,
		chainID:    chainID,
		haltHeight: haltHeight,
		states:     map[crypto.Address]*signerState{},
	}
}

// Populate fills tx.Metadata.SignerInfo for one tx. Must be called in the
// order txs were produced on the source chain (block-ascending, within-block
// index-ascending).
func (r *signerResolver) Populate(tx *gnoland.TxWithMetadata) error {
	if tx.Metadata == nil {
		tx.Metadata = &gnoland.GnoTxMetadata{}
	}

	stdTx := tx.Tx
	signers := stdTx.GetSigners()
	sigs := stdTx.GetSignatures()
	failed := tx.Metadata.Failed

	infos := make([]gnoland.SignerAccountInfo, len(signers))

	for j, signer := range signers {
		ss, err := r.state(signer)
		if err != nil {
			return err
		}
		infos[j] = gnoland.SignerAccountInfo{
			Address:    signer,
			AccountNum: ss.accNum,
			// Sequence filled below.
		}
	}
	tx.Metadata.SignerInfo = infos

	if failed {
		// Buffer failed tx signer info pointers — sequences are back-patched
		// at the next success (or in Finalize).
		for j, signer := range signers {
			ss := r.states[signer]
			ss.pendingFails = append(ss.pendingFails, &pendingFailedTx{
				info:    &tx.Metadata.SignerInfo[j],
				ownerSS: ss,
			})
			// Placeholder sequence.
			tx.Metadata.SignerInfo[j].Sequence = ss.seq
		}
		return nil
	}

	// Successful tx — resolve sequence per signer.
	for j, signer := range signers {
		ss := r.states[signer]

		needResolve := !ss.initialized || len(ss.pendingFails) > 0
		if needResolve {
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

			resolved, err := bruteForceSignerSequence(
				stdTx, sig, ss.accNum, lo, hi, r.chainID,
			)
			if err != nil {
				// Last resort: keep current counter. Subsequent txs may fail
				// verification, but at least export proceeds.
				resolved = ss.seq
			}

			// Back-patch buffered failed txs now that we know how much
			// sequence was consumed between the last success and this one.
			assignFailedTxSequences(ss.pendingFails, ss.seq, resolved)
			ss.pendingFails = nil
			ss.seq = resolved
			ss.initialized = true
		}

		tx.Metadata.SignerInfo[j].Sequence = ss.seq
		ss.seq++
	}
	return nil
}

// Finalize back-patches any trailing failed txs (those with no successor
// success to anchor against). Must be called once after the last Populate.
func (r *signerResolver) Finalize() {
	for _, ss := range r.states {
		if len(ss.pendingFails) == 0 {
			continue
		}

		var consumed uint64
		if ss.finalSeq > ss.seq {
			consumed = ss.finalSeq - ss.seq
		}
		if !ss.initialized && consumed > uint64(len(ss.pendingFails)) {
			// Never had a successful tx — cap consumed.
			ss.seq = ss.finalSeq - uint64(len(ss.pendingFails))
			consumed = uint64(len(ss.pendingFails))
		}
		assignTrailingFailedTxSequences(ss.pendingFails, ss.seq, consumed)
		ss.pendingFails = nil
	}
}

// state fetches-or-creates the signerState for addr.
func (r *signerResolver) state(addr crypto.Address) (*signerState, error) {
	if ss, ok := r.states[addr]; ok {
		return ss, nil
	}
	accNum, finalSeq, err := r.client.GetAccountAtHeight(addr, r.haltHeight)
	if err != nil {
		return nil, fmt.Errorf("fetch account state for %s at %d: %w",
			addr, r.haltHeight, err)
	}
	ss := &signerState{accNum: accNum, finalSeq: finalSeq}
	r.states[addr] = ss
	return ss, nil
}

// bruteForceSignerSequence tries sequences in [lo, hi] to find the one that
// makes the tx signature verify. Returns the pre-tx sequence (the value that
// was used in GetSignBytes on the source chain).
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
	return lo, fmt.Errorf("no sequence in [%d, %d] verified for account %d",
		lo, hi, accNum)
}

// assignFailedTxSequences back-patches SignerInfo.Sequence on buffered failed
// txs when we finally resolve the next successful-tx sequence.
//
// Cosmetic: failed txs are skipped on replay, so exact values don't matter
// for correctness. We approximate by assuming msg-fails (which consume
// sequence) come first in the gap, then ante-fails (which don't).
func assignFailedTxSequences(pending []*pendingFailedTx, startSeq, resolvedSeq uint64) {
	consumed := resolvedSeq - startSeq
	seq := startSeq
	for i, pf := range pending {
		pf.info.Sequence = seq
		if uint64(i) < consumed {
			seq++
		}
	}
}

// assignTrailingFailedTxSequences handles trailing failed txs with no later
// success to anchor against.
func assignTrailingFailedTxSequences(pending []*pendingFailedTx, startSeq, consumed uint64) {
	seq := startSeq
	for i, pf := range pending {
		pf.info.Sequence = seq
		if uint64(i) < consumed {
			seq++
		}
	}
}
