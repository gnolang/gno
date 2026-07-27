package rootmulti_test

// Shared helpers for the fastindex regression tests (gno#6011).

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

var reproCRC = crc32.MakeTable(crc32.Castagnoli)

// reproVerifyChecksum splits a bptree record into payload and crc32c,
// verifying it — an independent reimplementation of bptree's unexported
// verifyChecksum so the audits don't trust the code under test.
func reproVerifyChecksum(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("record too short (%d bytes)", len(data))
	}
	payload := data[:len(data)-4]
	want := binary.BigEndian.Uint32(data[len(data)-4:])
	if got := crc32.Checksum(payload, reproCRC); got != want {
		return nil, fmt.Errorf("crc mismatch")
	}
	return payload, nil
}
