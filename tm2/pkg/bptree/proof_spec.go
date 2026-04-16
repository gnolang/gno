package bptree

import ics23 "github.com/cosmos/ics23/go"

// BptreeSpec is the ICS23 ProofSpec for the B+ tree.
// The mini merkle collapses into a uniform chain of binary InnerOps.
//
// Depth constants (Finding #22):
//   - MinDepth = 5: one full mini-merkle traversal is log2(B) = log2(32) = 5
//     InnerOps. Even the smallest non-trivial tree (a single leaf inside a
//     single inner node) has at least one mini-merkle chain.
//   - MaxDepth = 60: a B+32 tree has log32(n) inner levels and each level
//     contributes log2(B) = 5 mini-merkle InnerOps, so MaxDepth covers 12
//     inner levels = ~ 32**12 ≈ 1.2e18 entries. Well beyond any realistic
//     gno.land application store; the bound exists so a malicious proof
//     cannot trigger unbounded ICS23 verification work.
var BptreeSpec = &ics23.ProofSpec{
	LeafSpec: &ics23.LeafOp{
		Prefix:       []byte{DomainLeaf}, // 0x00 (RFC 6962 leaf domain separator)
		PrehashKey:   ics23.HashOp_NO_HASH,
		PrehashValue: ics23.HashOp_SHA256,
		Hash:         ics23.HashOp_SHA256,
		Length:       ics23.LengthOp_VAR_PROTO,
	},
	InnerSpec: &ics23.InnerSpec{
		ChildOrder: []int32{0, 1}, // binary merkle
		// MinPrefixLength / MaxPrefixLength both = 1: each mini-merkle
		// InnerOp emits exactly one prefix byte — the 0x01 inner-node
		// domain separator. ICS23's verifier adds `maxLeftChildBytes`
		// (== ChildSize == 32 here) to the prefix bound internally when
		// checking op.Prefix length, so declaring 1/1 here is correct
		// and not accidentally too tight. See Finding #22.
		MinPrefixLength: 1,
		MaxPrefixLength: 1,
		ChildSize:       int32(HashSize), // 32
		EmptyChild:      sentinelHash[:], // SHA256(0x02)
		Hash:            ics23.HashOp_SHA256,
	},
	MinDepth: 5,
	MaxDepth: 60,
}
