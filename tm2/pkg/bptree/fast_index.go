package bptree

import (
	"encoding/binary"
	"fmt"
)

// The fast index is an OPTIONAL, latest-version, inline read accelerator. It maps
// user-key → version‖value ('F'‖key → version(8)‖value‖crc32c, the standard
// record framing), so a
// point Get of a PRESENT key against committed state resolves in 1 DB read
// (the entry carries the value) instead of a full tree descent plus the
// out-of-line value read. Two read surfaces consult it: committed snapshots
// (ImmutableTree.Get) and the CLEAN working tree (MutableTree.Get while the
// session has no staged mutations — see MutableTree.fastReadable), which is
// byte-identical to the committed snapshot at its version.
//
// Trust contract: index currency is verified by Load (ensureFastIndex rebuilds
// on a stamp mismatch) and preserved from then on by eager same-batch
// maintenance (Set/Remove/SaveVersion) and by Import dropping the index up
// front. A tree reached ONLY via LoadVersion — never Load — over a DB whose
// later versions were committed with the feature off is outside the contract:
// nothing re-verifies the stamp there. The in-repo store layer always goes
// through Load.
//
// Properties:
//   - Not in the Merkle commitment — an unauthenticated accelerator, like cosmos
//     IAVL fast nodes. App hash and proofs are identical with it on or off, so a
//     node may toggle it without forking.
//   - Inline — stores the latest value a second time (the authoritative copy
//     still lives under PrefixVal). Both are written from the same value in the
//     same batch, so they can never disagree. Worth it for small values; the
//     duplication scales with value size.
//   - Maintained in the SAME batch as the tree (Set/Remove stage into ndb.batch,
//     committed atomically by SaveVersion → Commit), so it can never disagree
//     with the committed tree, even across a crash.
//   - ADVISORY on read: a hit is trusted only when its version ≤ the snapshot
//     version (else the entry is newer than the reader's snapshot); a miss,
//     a too-new entry, or a corrupt/too-short entry all fall back to the
//     authoritative tree walk. Index completeness is therefore a performance
//     property, never a correctness one.

// fastDBKey builds the fast-index DB key: PrefixFast ‖ userKey.
func fastDBKey(userKey []byte) []byte {
	key := make([]byte, 1+len(userKey))
	key[0] = PrefixFast
	copy(key[1:], userKey)
	return key
}

// metaFastVersionKey is the PrefixMeta key stamping the version the persisted
// fast index is complete for (used to decide rebuild-on-Load). Constant; only
// ever used read-only as a DB key, so the shared backing is safe.
var metaFastVersionKey = append([]byte{PrefixMeta}, "fastidx"...)

// setFastIndex stages userKey → version‖value in the batch (the version is the
// valueKey's — the version the value was written at). No-op when the fast index
// is disabled, so Set's call site is unconditional and zero-cost when off.
func (ndb *nodeDB) setFastIndex(userKey, vk, value []byte) error {
	if !ndb.opts.FastIndex {
		return nil
	}
	// Build the record in one buffer (per-mutation hot path — avoids
	// stampChecksum's second copy of the value bytes).
	rec := make([]byte, 8+len(value)+checksumSize)
	copy(rec[:8], vk[:8]) // version prefix, from the valueKey
	copy(rec[8:], value)
	return ndb.batch.Set(fastDBKey(userKey), sealChecksum(rec))
}

// deleteFastIndex stages removal of userKey from the index. No-op when disabled.
// This delete is LOAD-BEARING for correctness: a missing entry is always safe
// (→ tree walk), but a leftover entry for a removed key whose vkVersion ≤ a
// reader's snapshot would be wrongly trusted and return the pre-removal value.
func (ndb *nodeDB) deleteFastIndex(userKey []byte) error {
	if !ndb.opts.FastIndex {
		return nil
	}
	return ndb.batch.Delete(fastDBKey(userKey))
}

// fastGet attempts an advisory fast-index read for committed state at version
// s (a committed snapshot, or the clean working tree at its committed
// version). Returns (value, true) on a trusted hit, or (nil, false) to fall
// back to the tree walk (miss / corrupt / too-short / entry newer than s). It
// reads committed state only (never pendingVals or the staged batch).
func (ndb *nodeDB) fastGet(userKey []byte, s int64) ([]byte, bool) {
	data, err := ndb.db.Get(fastDBKey(userKey))
	if err != nil || data == nil {
		return nil, false
	}
	payload, err := verifyChecksum(data)
	if err != nil || len(payload) < 8 {
		return nil, false
	}
	if vkVersion(payload) > s {
		return nil, false // entry newer than the snapshot
	}
	// payload = version(8) ‖ value; copy out (the re-slice aliases db storage).
	return copyKey(payload[8:]), true
}

// setFastIndexVersion stages the index-complete-through stamp into the batch
// (committed atomically with the index entries it describes). No-op when the
// fast index is disabled, so SaveVersion can call it unconditionally.
func (ndb *nodeDB) setFastIndexVersion(v int64) error {
	if !ndb.opts.FastIndex {
		return nil
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	return ndb.batch.Set(metaFastVersionKey, stampChecksum(b[:]))
}

// getFastIndexVersion reads the stamp. The bool is false when no stamp exists
// (index never built / built without the feature).
func (ndb *nodeDB) getFastIndexVersion() (int64, bool, error) {
	data, err := ndb.db.Get(metaFastVersionKey)
	if err != nil {
		return 0, false, err
	}
	if data == nil {
		return 0, false, nil
	}
	payload, err := verifyChecksum(data)
	if err != nil {
		return 0, false, fmt.Errorf("fast index version: %w", err)
	}
	if len(payload) != 8 {
		return 0, false, fmt.Errorf("fast index version: bad length %d", len(payload))
	}
	return int64(binary.BigEndian.Uint64(payload)), true, nil
}

// fastRebuildFlush bounds the batch during a full rebuild/clear so a 100M-key
// rebuild doesn't hold an unbounded batch in memory.
const fastRebuildFlush = 1 << 16

// clearFastIndex stages deletion of every existing 'F' entry, flushing in
// bounded chunks. After each chunk Commits, the deleted keys are gone from the
// DB, so re-opening the iterator from the range start finds the remaining keys.
// Leaves a fresh batch for the caller.
func (ndb *nodeDB) clearFastIndex() error {
	prefix := []byte{PrefixFast}
	end := []byte{PrefixFast + 1}
	for {
		itr, err := ndb.db.Iterator(prefix, end)
		if err != nil {
			return err
		}
		n := 0
		for ; itr.Valid() && n < fastRebuildFlush; itr.Next() {
			k := itr.Key()
			kc := make([]byte, len(k))
			copy(kc, k)
			if err := ndb.batch.Delete(kc); err != nil {
				itr.Close()
				return err
			}
			n++
		}
		ierr := itr.Error()
		itr.Close()
		if ierr != nil {
			return ierr
		}
		if n == 0 {
			return nil // range exhausted
		}
		if err := ndb.Commit(); err != nil {
			return err
		}
		if n < fastRebuildFlush {
			return nil
		}
	}
}

// dropFastIndex removes the fast index entirely: the completeness stamp first
// (with its own commit — clearFastIndex's empty-range path returns without
// committing, so the stamp delete must not ride a chunk commit), then every
// 'F' entry in bounded chunks. Stamp-first ordering makes an abort at any
// point safe: the stamp is already gone, so the next Load rebuilds; a
// partially-cleared index is only ever a perf loss, never a wrong read.
// No-op when the feature is off. Precondition: the batch holds no unrelated
// staged state (the chunked clear self-commits).
func (ndb *nodeDB) dropFastIndex() (err error) {
	if !ndb.opts.FastIndex {
		return nil
	}
	// A failed Commit recycles the batch on its own path, so the trailing
	// discard there is a harmless no-op (same convention as rebuildFastIndex).
	defer func() {
		if err != nil {
			ndb.DiscardBatch()
		}
	}()
	if err = ndb.batch.Delete(metaFastVersionKey); err != nil {
		return err
	}
	if err = ndb.Commit(); err != nil {
		return err
	}
	return ndb.clearFastIndex()
}

// ensureFastIndex rebuilds the fast index from the latest root if it is absent
// or stale (the stamp != the loaded version). Called from Load when the feature
// is on. The index is advisory, so a stale/missing index is never wrong, only
// slower; a rebuild error is returned to Load's caller (the loaded tree is still
// usable, and a retry Load re-attempts the rebuild).
func (t *MutableTree) ensureFastIndex() error {
	if !t.ndb.opts.FastIndex {
		return nil
	}
	stamp, ok, err := t.ndb.getFastIndexVersion()
	if err != nil {
		return err
	}
	if ok && stamp == t.version {
		return nil // already complete through the loaded version
	}
	return t.rebuildFastIndex()
}

// rebuildFastIndex clears any stale 'F' entries, then re-derives the index from
// the latest committed root (one ordered leaf walk, resolving each live value),
// and stamps it complete through t.version. All staging rides ndb.batch and is
// dropped by DiscardBatch on any error, so a crash mid-rebuild leaves the stamp
// stale and the next Load rebuilds. The clear handles the one stale-PRESENT route
// (an externally-stale index whose removed keys the live-only walk wouldn't
// touch).
func (t *MutableTree) rebuildFastIndex() (err error) {
	// Any error leaves staged writes uncommitted; drop them so a later Commit
	// can't flush a partial index. (Commit recycles the batch on its own path,
	// so a trailing discard after a failed Commit is a harmless no-op.)
	defer func() {
		if err != nil {
			t.ndb.DiscardBatch()
		}
	}()

	// A first-enable rebuild on a large DB reads and re-stores every live value,
	// so it can take a while; log start/progress/done (no-op under NopLogger).
	// Logged before clearFastIndex so the (also potentially long) clear isn't
	// silent on a re-backfill over an existing index.
	t.ndb.logger.Info("bptree: rebuilding fast index", "version", t.version)
	if err = t.ndb.clearFastIndex(); err != nil {
		return err
	}
	n := 0
	if t.root != nil {
		_, walkErr := iterateNodeResolved(t.root, func(key, vk []byte) bool {
			var value []byte
			if value, err = t.ndb.getCommittedValue(vk); err != nil {
				return true // resolve failed; err is the named return
			}
			if err = t.ndb.setFastIndex(key, vk, value); err != nil {
				return true // stop; err is the named return
			}
			n++
			if n%fastRebuildFlush == 0 {
				err = t.ndb.Commit()
				if err == nil && n%(fastRebuildFlush*64) == 0 {
					t.ndb.logger.Info("bptree: fast index rebuild progress", "entries", n)
				}
			}
			return err != nil
		})
		if walkErr != nil { // child-load error from the walk itself
			return walkErr
		}
		if err != nil { // staging/commit error captured in the callback
			return err
		}
	}
	if err = t.ndb.setFastIndexVersion(t.version); err != nil {
		return err
	}
	if err = t.ndb.Commit(); err != nil {
		return err
	}
	t.ndb.logger.Info("bptree: fast index rebuilt", "entries", n)
	return nil
}
