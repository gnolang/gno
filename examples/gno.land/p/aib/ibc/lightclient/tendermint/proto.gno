package tendermint

import (
	"encoding/binary"
	"time"
)

// VoteBytesToSign returns the bytes-to-sign for a validator's vote.
// These bytes are produced from a protobuf marshaled type [CanonicalVote][1].
// Since proto is not available in Gno, the function tries to generate these
// bytes manually, using the same logic as protobuf.
//
// [1]: https://buf.build/tendermint/tendermint/docs/main:tendermint.types#tendermint.types.CanonicalVote
func VoteBytesToSign(
	chainID string,
	height, round int64,
	blockIDHash []byte,
	partSetTotal uint32,
	partSetHash []byte,
	blockIDFlag BlockIDFlag,
	timestamp time.Time,
) []byte {
	var buf []byte

	// Field 1: type (SignedMsgType - varint)
	buf = appendTag(buf, 1, 0) // field 1, wire type 0 (varint)
	buf = appendVarint(buf, 2) // type precommit

	// Field 2: height (sfixed64 - fixed64)
	if height != 0 {
		buf = appendTag(buf, 2, 1) // field 2, wire type 1 (fixed64)
		buf = appendFixed64(buf, uint64(height))
	}

	// Field 3: round (sfixed64 - fixed64)
	if round != 0 {
		buf = appendTag(buf, 3, 1) // field 3, wire type 1 (fixed64)
		buf = appendFixed64(buf, uint64(round))
	}

	// Field 4: block_id (CanonicalBlockID - length-delimited)
	if blockIDFlag == BlockIDFlagCommit {
		blockIDBytes := serializeCanonicalBlockID(blockIDHash, partSetTotal, partSetHash)
		if len(blockIDBytes) > 0 {
			buf = appendTag(buf, 4, 2) // field 4, wire type 2 (length-delimited)
			buf = appendVarint(buf, uint64(len(blockIDBytes)))
			buf = append(buf, blockIDBytes...)
		}
	}

	// Field 5: timestamp (google.protobuf.Timestamp - length-delimited)
	timestampBytes := serializeTimestamp(timestamp)
	if len(timestampBytes) > 0 {
		buf = appendTag(buf, 5, 2) // field 5, wire type 2 (length-delimited)
		buf = appendVarint(buf, uint64(len(timestampBytes)))
		buf = append(buf, timestampBytes...)
	}

	// Field 6: chain_id (string - length-delimited)
	if len(chainID) > 0 {
		buf = appendTag(buf, 6, 2) // field 6, wire type 2 (length-delimited)
		buf = appendVarint(buf, uint64(len(chainID)))
		buf = append(buf, []byte(chainID)...)
	}

	// Prefix with total length
	length := len(buf)
	buf = append(appendVarint(nil, uint64(length)), buf...)

	return buf
}

// serializeCanonicalBlockID serializes a CanonicalBlockID message
func serializeCanonicalBlockID(hash []byte, partSetTotal uint32, partSetHash []byte) []byte {
	var buf []byte

	// Field 1: hash (bytes)
	if len(hash) > 0 {
		buf = appendTag(buf, 1, 2) // field 1, wire type 2 (length-delimited)
		buf = appendVarint(buf, uint64(len(hash)))
		buf = append(buf, hash...)
	}

	// Field 2: part_set_header (CanonicalPartSetHeader - embedded message)
	partSetHeaderBytes := serializeCanonicalPartSetHeader(partSetTotal, partSetHash)
	if len(partSetHeaderBytes) > 0 {
		buf = appendTag(buf, 2, 2) // field 2, wire type 2 (length-delimited)
		buf = appendVarint(buf, uint64(len(partSetHeaderBytes)))
		buf = append(buf, partSetHeaderBytes...)
	}

	return buf
}

// serializeCanonicalPartSetHeader serializes a CanonicalPartSetHeader message
func serializeCanonicalPartSetHeader(total uint32, hash []byte) []byte {
	var buf []byte

	// Field 1: total (uint32 - varint)
	if total != 0 {
		buf = appendTag(buf, 1, 0) // field 1, wire type 0 (varint)
		buf = appendVarint(buf, uint64(total))
	}

	// Field 2: hash (bytes)
	if len(hash) > 0 {
		buf = appendTag(buf, 2, 2) // field 2, wire type 2 (length-delimited)
		buf = appendVarint(buf, uint64(len(hash)))
		buf = append(buf, hash...)
	}

	return buf
}

// serializeTimestamp serializes a google.protobuf.Timestamp
func serializeTimestamp(t time.Time) []byte {
	var buf []byte

	seconds := t.Unix()
	nanos := int32(t.Nanosecond())

	// Field 1: seconds (int64 - varint)
	if seconds != 0 {
		buf = appendTag(buf, 1, 0) // field 1, wire type 0 (varint)
		buf = appendVarint(buf, uint64(seconds))
	}

	// Field 2: nanos (int32 - varint)
	if nanos != 0 {
		buf = appendTag(buf, 2, 0) // field 2, wire type 0 (varint)
		buf = appendVarint(buf, uint64(nanos))
	}

	return buf
}

// appendTag appends a protobuf tag (field number and wire type)
func appendTag(buf []byte, fieldNum int, wireType int) []byte {
	tag := (fieldNum << 3) | wireType
	return appendVarint(buf, uint64(tag))
}

// appendVarint appends a varint-encoded value
func appendVarint(buf []byte, v uint64) []byte {
	for v >= 0x80 {
		buf = append(buf, byte(v)|0x80)
		v >>= 7
	}
	return append(buf, byte(v))
}

// appendFixed64 appends a fixed 64-bit value (little-endian)
func appendFixed64(buf []byte, v uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return append(buf, b[:]...)
}
