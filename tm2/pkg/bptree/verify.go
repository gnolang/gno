package bptree

import (
	"bytes"
	"fmt"
)

// verifyMismatchCap bounds the per-report sample of mismatches so auditing a
// pathologically corrupt index cannot exhaust memory; MismatchCount always
// reflects the true total.
const verifyMismatchCap = 256

// FastIndexMismatchKind classifies why a persisted 'F' entry disagrees with the
// authoritative tree.
type FastIndexMismatchKind string

const (
	// MismatchStaleValue: the entry's key exists in the tree but its inlined
	// value differs from the tree's authoritative value — the gno#6011 signature
	// (a fast read would return the stale value).
	MismatchStaleValue FastIndexMismatchKind = "stale-value"
	// MismatchOrphan: the entry's key is absent from the tree entirely (a
	// leftover entry for a removed key).
	MismatchOrphan FastIndexMismatchKind = "orphan"
	// MismatchCorrupt: the entry's record failed checksum/length validation.
	MismatchCorrupt FastIndexMismatchKind = "corrupt"
)

// FastIndexMismatch is one 'F' entry that disagrees with the authoritative tree
// at the audited version.
type FastIndexMismatch struct {
	Key         []byte
	Kind        FastIndexMismatchKind
	FastVersion int64 // version stamped in the 'F' entry (0 if corrupt)
	TreeVersion int64 // version of the authoritative valueKey (0 unless stale-value)
}

func (m FastIndexMismatch) String() string {
	switch m.Kind {
	case MismatchStaleValue:
		return fmt.Sprintf("stale-value key=%X fast_version=%d tree_version=%d", m.Key, m.FastVersion, m.TreeVersion)
	case MismatchOrphan:
		return fmt.Sprintf("orphan     key=%X fast_version=%d (absent from tree)", m.Key, m.FastVersion)
	default:
		return fmt.Sprintf("corrupt    key=%X", m.Key)
	}
}

// FastIndexReport summarizes a VerifyFastIndex audit.
type FastIndexReport struct {
	Version       int64 // the tree version audited (the loaded version)
	Stamp         int64 // the fast-index completeness stamp
	StampPresent  bool  // whether a stamp exists
	Entries       int   // total 'F' entries scanned
	MismatchCount int   // total mismatches found (may exceed len(Mismatches))
	// Mismatches is a bounded sample (up to verifyMismatchCap) of the mismatches
	// found, for reporting; MismatchCount is the true total.
	Mismatches []FastIndexMismatch
}

// Healthy reports whether the index is trustworthy at the audited version: the
// stamp is current (== version) and no entry disagrees with the tree. A stamp
// BEHIND the version is not a corruption — those entries would be rebuilt on the
// next live Load — so callers that only care about "index behind" versus
// "index current but wrong" should consult Stamp/Version and MismatchCount
// directly.
func (r *FastIndexReport) Healthy() bool {
	return r.StampPresent && r.Stamp == r.Version && r.MismatchCount == 0
}

// VerifyFastIndex audits the persisted fast index against the authoritative
// tree at the loaded version: every 'F' entry's inlined value must byte-equal
// the tree's committed value for that key. It is READ-ONLY — it never writes,
// stages, or rebuilds — so it is safe to run on a snapshot, a read-only DB, or
// an operator's captured node state.
//
// The expected value is resolved via a raw treeLookup + committed value read,
// never via Get/GetVersioned/ImmutableTree (which would consult the very index
// under audit on a clean tree, making the check circular). A missing 'F' entry
// is never a fault — a miss falls back to the authoritative walk — so only
// PRESENT-but-wrong entries are reported.
//
// The tree must be loaded at the version the index reflects (the latest
// committed version); callers Load or LoadReadonly first. When the stamp is
// BEHIND the loaded version the index is merely incomplete (feature toggled off
// for some versions) and its stale entries would be healed by the next live
// Load; the report surfaces Stamp/Version so the caller can distinguish that
// benign case from a stamp-current index that nonetheless disagrees (the
// gno#6011 corruption).
func (t *MutableTree) VerifyFastIndex() (*FastIndexReport, error) {
	report := &FastIndexReport{Version: t.version}
	stamp, ok, err := t.ndb.getFastIndexVersion()
	if err != nil {
		return nil, err
	}
	report.Stamp, report.StampPresent = stamp, ok

	itr, err := t.ndb.db.Iterator([]byte{PrefixFast}, []byte{PrefixFast + 1})
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		report.Entries++
		userKey := append([]byte(nil), itr.Key()[1:]...) // strip PrefixFast

		payload, cerr := verifyChecksum(itr.Value())
		if cerr != nil || len(payload) < 8 {
			report.record(FastIndexMismatch{Key: userKey, Kind: MismatchCorrupt})
			continue
		}
		fastVer := vkVersion(payload)

		if t.root == nil {
			report.record(FastIndexMismatch{Key: userKey, Kind: MismatchOrphan, FastVersion: fastVer})
			continue
		}
		_, _, vk, found, lerr := treeLookup(t.root, userKey)
		if lerr != nil {
			return nil, fmt.Errorf("tree lookup %X: %w", userKey, lerr)
		}
		if !found {
			report.record(FastIndexMismatch{Key: userKey, Kind: MismatchOrphan, FastVersion: fastVer})
			continue
		}
		want, verr := t.ndb.getCommittedValue(vk)
		if verr != nil {
			return nil, fmt.Errorf("committed value for %X: %w", userKey, verr)
		}
		if !bytes.Equal(payload[8:], want) {
			report.record(FastIndexMismatch{
				Key: userKey, Kind: MismatchStaleValue,
				FastVersion: fastVer, TreeVersion: vkVersion(vk),
			})
		}
	}
	if err := itr.Error(); err != nil {
		return nil, err
	}
	return report, nil
}

func (r *FastIndexReport) record(m FastIndexMismatch) {
	r.MismatchCount++
	if len(r.Mismatches) < verifyMismatchCap {
		r.Mismatches = append(r.Mismatches, m)
	}
}
