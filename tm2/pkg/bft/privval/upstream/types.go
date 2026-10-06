// Package upstream defines tm2 types whose amino wire-format is byte-identical
// to upstream Tendermint v0.34's protobuf encoding. They exist only for the
// privval-wire boundary — the path between a gno.land validator and an
// external upstream-protocol KMS such as tmkms.
//
// The chain's own Vote/Proposal/BlockID/PartSetHeader types in
// tm2/pkg/bft/types remain unchanged; they're used in p2p messages and
// blocks, where any byte-level shift would be a network-coordinated upgrade.
// The translator (translator.go) converts between the two shapes.
//
// CanonicalVote and CanonicalProposal are NOT redefined here — they live in
// tm2/pkg/bft/types and were re-tagged (POLRound varint;
// CanonicalPartSetHeader uint32 Total first) to be byte-identical to
// upstream's canonical.proto.
package upstream

import (
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// Vote matches upstream Tendermint v0.34's types.proto Vote message:
//
//	message Vote {
//	  SignedMsgType type              = 1;
//	  int64         height            = 2;
//	  int32         round             = 3;
//	  BlockID       block_id          = 4;
//	  google.protobuf.Timestamp timestamp = 5;
//	  bytes         validator_address = 6;
//	  int32         validator_index   = 7;
//	  bytes         signature         = 8;
//	}
//
// Differences from tm2/pkg/bft/types.Vote:
//   - height: int64 plain varint (tm2: sint64 zigzag)
//   - round: int32 plain varint (tm2: sint64)
//   - validator_address: raw bytes (tm2: bech32 string via Address.MarshalAmino)
//   - validator_index: int32 plain varint (tm2: sint64)
type Vote struct {
	Type             types.SignedMsgType
	Height           int64 `binary:"varint"`
	Round            int32 `binary:"varint"`
	BlockID          BlockID
	Timestamp        time.Time
	ValidatorAddress []byte
	ValidatorIndex   int32 `binary:"varint"`
	Signature        []byte
}

// Proposal matches upstream Tendermint v0.34's types.proto Proposal:
//
//	message Proposal {
//	  SignedMsgType type      = 1;
//	  int64         height    = 2;
//	  int32         round     = 3;
//	  int32         pol_round = 4;
//	  BlockID       block_id  = 5;
//	  google.protobuf.Timestamp timestamp = 6;
//	  bytes         signature = 7;
//	}
type Proposal struct {
	Type      types.SignedMsgType
	Height    int64 `binary:"varint"`
	Round     int32 `binary:"varint"`
	POLRound  int32 `binary:"varint"`
	BlockID   BlockID
	Timestamp time.Time
	Signature []byte
}

// BlockID matches upstream's types.proto BlockID:
//
//	message BlockID {
//	  bytes         hash            = 1;
//	  PartSetHeader part_set_header = 2;
//	}
//
// The Go field name is PartSetHeader (matching upstream) rather than tm2's
// PartsHeader. Only field POSITION matters for wire bytes.
type BlockID struct {
	Hash          []byte
	PartSetHeader PartSetHeader
}

// PartSetHeader matches upstream's types.proto PartSetHeader:
//
//	message PartSetHeader {
//	  uint32 total = 1;
//	  bytes  hash  = 2;
//	}
//
// Difference from tm2/pkg/bft/types.PartSetHeader:
//   - total: uint32 plain varint (tm2: int → sint64 zigzag)
//
// Field order matches tm2 (Total first); only the type changes.
type PartSetHeader struct {
	Total uint32
	Hash  []byte
}
