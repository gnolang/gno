package tendermint

import (
	"time"

	"gno.land/p/aib/encoding/proto"
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
	buf = proto.AppendVarint(buf, 1, 2) // type precommit

	// Field 2: height (sfixed64 - fixed64)
	buf = proto.AppendFixed64(buf, 2, uint64(height))

	// Field 3: round (sfixed64 - fixed64)
	buf = proto.AppendFixed64(buf, 3, uint64(round))

	// Field 4: block_id (CanonicalBlockID - length-delimited)
	if blockIDFlag == BlockIDFlagCommit {
		buf = appendCanonicalBlockID(buf, 4, blockIDHash, partSetTotal, partSetHash)
	}

	// Field 5: timestamp (google.protobuf.Timestamp - length-delimited)
	buf = proto.AppendTime(buf, 5, timestamp)

	// Field 6: chain_id (string - length-delimited)
	buf = proto.AppendLengthDelimited(buf, 6, []byte(chainID))

	// Prefix with total length
	length := len(buf)
	buf = append(proto.Varint(uint64(length)), buf...)

	return buf
}

// appendCanonicalBlockID serializes a CanonicalBlockID message
func appendCanonicalBlockID(buf []byte, fieldNum int, hash []byte, partSetTotal uint32, partSetHash []byte) []byte {
	var _buf []byte

	// Field 1: hash (bytes)
	_buf = proto.AppendLengthDelimited(_buf, 1, hash)

	// Field 2: part_set_header (CanonicalPartSetHeader - embedded message)
	_buf = appendCanonicalPartSetHeader(_buf, 2, partSetTotal, partSetHash)

	return proto.AppendLengthDelimited(buf, fieldNum, _buf)
}

// appendCanonicalPartSetHeader serializes a CanonicalPartSetHeader message
func appendCanonicalPartSetHeader(buf []byte, fieldNum int, total uint32, hash []byte) []byte {
	var _buf []byte

	// Field 1: total (uint32 - varint)
	_buf = proto.AppendVarint(_buf, 1, uint64(total))

	// Field 2: hash (bytes)
	_buf = proto.AppendLengthDelimited(_buf, 2, hash)

	return proto.AppendLengthDelimited(buf, fieldNum, _buf)
}
