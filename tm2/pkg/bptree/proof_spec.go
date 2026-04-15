package bptree

import ics23 "github.com/cosmos/ics23/go"

// BptreeSpec is the ICS23 ProofSpec for the B+ tree.
// The mini merkle collapses into a uniform chain of binary InnerOps.
var BptreeSpec = &ics23.ProofSpec{
	LeafSpec: &ics23.LeafOp{
		Prefix:       []byte{DomainLeaf},      // 0x00 (RFC 6962 leaf domain separator)
		PrehashKey:   ics23.HashOp_NO_HASH,
		PrehashValue: ics23.HashOp_SHA256,
		Hash:         ics23.HashOp_SHA256,
		Length:       ics23.LengthOp_VAR_PROTO,
	},
	InnerSpec: &ics23.InnerSpec{
		ChildOrder:      []int32{0, 1},           // binary merkle
		MinPrefixLength: 1,                       // the 0x01 domain separator
		MaxPrefixLength: 1,                       // just the 0x01 byte; ICS23 adds maxLeftChildBytes internally
		ChildSize:       int32(HashSize),          // 32
		EmptyChild:      sentinelHash[:],          // SHA256(0x02)
		Hash:            ics23.HashOp_SHA256,
	},
	MinDepth: 5,  // at least one mini-merkle traversal (log2(B) = 5)
	MaxDepth: 60, // supports trees up to ~12 inner levels (~billions of entries)
}
