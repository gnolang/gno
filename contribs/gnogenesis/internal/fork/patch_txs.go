package fork

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// txKey uniquely identifies a historical tx by its replay coordinates. The
// (signer, sequence) pair is globally unique per chain (nonce contract) so
// combined with block_height this gives a stable per-tx identity that we can
// match across the source stream and patch files.
type txKey struct {
	height int64
	signer crypto.Address
	seq    uint64
}

// keyOf derives a txKey from metadata. Returns false if any required field is
// missing — historical txs from the source stream and patch entries must both
// carry block_height + at least one SignerInfo entry.
func keyOf(meta *gnoland.GnoTxMetadata) (txKey, bool) {
	if meta == nil || meta.BlockHeight == 0 || len(meta.SignerInfo) == 0 {
		return txKey{}, false
	}
	si := meta.SignerInfo[0]
	return txKey{height: meta.BlockHeight, signer: si.Address, seq: si.Sequence}, true
}

func (k txKey) String() string {
	return fmt.Sprintf("h=%d sender=%s seq=%d", k.height, k.signer, k.seq)
}

// applyPatchTxs loads AnnotatedTx entries from each path and rewrites matching
// historical txs in-place. Each patch is keyed by (block_height, signer[0],
// sequence). The pre-patch tx is stored on the entry's metadata as OriginalTx,
// Source is set to "patched", and Note is set to the patch's Reason.
//
// Returns the number of txs rewritten. Fails fast on any of:
//   - patch entry missing block_height or signer_info
//   - duplicate key within one file or across files
//   - patch key with no matching source tx (typo guard)
func applyPatchTxs(txs []gnoland.TxWithMetadata, paths []string, io commands.IO) (int, error) {
	if len(paths) == 0 {
		return 0, nil
	}

	patchByKey := make(map[txKey]*AnnotatedTx)
	originPath := make(map[txKey]string)
	for _, path := range paths {
		entries, err := readAnnotatedTxs(path)
		if err != nil {
			return 0, fmt.Errorf("read %s: %w", path, err)
		}
		for i, at := range entries {
			k, ok := keyOf(at.Metadata)
			if !ok {
				return 0, fmt.Errorf("%s line %d: patch entry needs block_height + signer_info[0]", path, i+1)
			}
			if prev, dup := originPath[k]; dup {
				return 0, fmt.Errorf("duplicate patch key (%s): %s and %s", k, prev, path)
			}
			at2 := at
			patchByKey[k] = &at2
			originPath[k] = path
		}
	}

	matched := make(map[txKey]bool, len(patchByKey))
	patchCount := 0
	for i := range txs {
		k, ok := keyOf(txs[i].Metadata)
		if !ok {
			continue
		}
		patch, has := patchByKey[k]
		if !has {
			continue
		}
		if matched[k] {
			return 0, fmt.Errorf("source stream has duplicate (%s) — chain history corrupted?", k)
		}
		original := txs[i].Tx
		txs[i].Tx = patch.Tx
		if txs[i].Metadata == nil {
			txs[i].Metadata = &gnoland.GnoTxMetadata{}
		}
		txs[i].Metadata.Source = gnoland.SourcePatched
		txs[i].Metadata.Note = patch.Reason
		txs[i].Metadata.OriginalTx = &original
		matched[k] = true
		patchCount++
	}

	if patchCount != len(patchByKey) {
		var unmatched []string
		for k := range patchByKey {
			if !matched[k] {
				unmatched = append(unmatched, fmt.Sprintf("%s (%s)", originPath[k], k))
			}
		}
		sort.Strings(unmatched)
		return 0, fmt.Errorf("%d patch(es) did not match any source tx: %s", len(unmatched), strings.Join(unmatched, ", "))
	}

	io.Printf("  patched %d historical tx(s) from %d patch file(s)\n", patchCount, len(paths))
	return patchCount, nil
}
