package bptree

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

// Every persisted record ('B' node, 'V' value, 'R' root, 'O' orphan) is
// stored as payload || crc32c(payload). The CRC makes disk/transport
// corruption fail loud at read time — including corruption of fields outside
// the Merkle commitment (inner separators, childSizes, height, raw value
// bytes), which would otherwise be served as wrong-but-hash-valid data. It is
// NOT a defense against an adversarial DB writer (who recomputes it);
// adversarial import streams are validated structurally by the Importer, and
// committed fields are covered by the root hash.
const checksumSize = 4

var crcTable = crc32.MakeTable(crc32.Castagnoli)

// stampChecksum returns a fresh buffer containing payload || crc32c(payload).
// The result never aliases payload: batch backends (memdb, boltdb) retain
// staged slices by reference, so the staged record must not share backing
// with any buffer the package or a caller can still reach.
func stampChecksum(payload []byte) []byte {
	out := make([]byte, len(payload)+checksumSize)
	copy(out, payload)
	binary.BigEndian.PutUint32(out[len(payload):], crc32.Checksum(payload, crcTable))
	return out
}

// sealChecksum writes crc32c over rec's payload (everything before the
// trailing checksumSize bytes) into rec's tail and returns rec. For callers
// that build the payload directly in a full-size fresh buffer — same format
// as stampChecksum without the second payload copy. rec must not alias
// caller-reachable memory (same rule as stampChecksum).
func sealChecksum(rec []byte) []byte {
	payload := rec[:len(rec)-checksumSize]
	binary.BigEndian.PutUint32(rec[len(rec)-checksumSize:], crc32.Checksum(payload, crcTable))
	return rec
}

// verifyChecksum splits a stored record into payload and checksum, verifying
// the CRC. The returned payload is a zero-copy re-slice of data — callers
// that hand it outside the package must copy it (see getCommittedValue). For
// a valid 4-byte record (empty payload) the result is non-nil and empty.
func verifyChecksum(data []byte) ([]byte, error) {
	if len(data) < checksumSize {
		return nil, fmt.Errorf("%w: record too short (%d bytes)", ErrChecksumMismatch, len(data))
	}
	payload := data[:len(data)-checksumSize]
	want := binary.BigEndian.Uint32(data[len(data)-checksumSize:])
	if got := crc32.Checksum(payload, crcTable); got != want {
		return nil, fmt.Errorf("%w: crc %08x != stored %08x over %d payload bytes",
			ErrChecksumMismatch, got, want, len(payload))
	}
	return payload, nil
}
