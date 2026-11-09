# ADR-prxxxx: Compute a simple Merkle root with no heap allocation per node

## Context

`SimpleHashFromByteSlices` returned a fresh slice from every node of the tree, and `leafHash` and
`innerHash` allocated two and three times, to concatenate the prefix and the input and for the sum.
`Txs.Hash` allocated a 32-byte slice per tx on top. A block's `ValidateBasic` hashes its
transactions this way, and a peer can send a block of four million empty transactions in one 8 MB
message: `Txs.Hash` over 4M txs measured 1190 MB in 24M allocations.

## Decision

The tree recursion returns each node by value, as a `[tmhash.Size]byte`, and only the root is copied
to a slice. A leaf is hashed through a SHA-256 digest the compiler keeps on the stack, and an inner
node from a stack buffer holding the prefix and both children. `Txs.Hash` writes every tx hash into
one backing array.

Every root stays byte-equal for every input, which a test checks against
`SimpleHashFromByteSlicesIterative`, built on the unchanged `leafHash` and `innerHash`, over item
counts from 0 to 4097 and leaf lengths from nil to 1000 bytes. The gnovm native
`crypto/merkle.hashFromByteSlices` charges gas from its table entry, so no gas figure moves.

Measured on 4M empty txs: `Txs.Hash` 213 MB in 3 allocations, and a block's `ValidateBasic` 213 MB
where it was 1190 MB.

## Alternatives considered

A stack buffer for a leaf of up to 64 bytes, and the old concatenation above that. Rejected: `go
build -gcflags=-m` keeps the digest of `tmhash.New` on the stack, so the digest hashes every leaf
length with no allocation and no branch.

A Merkle entry point taking leaves by index, so `Txs.Hash` builds no slice of hashes. It would
remove the 213 MB left, the 32-byte hash and the slice header per tx, and add an exported API to
`merkle` for one caller. Left for a change that needs it.

A cap on the number of transactions in a block. Rejected for this change: no stateless bound
separates an honest block from a stuffed one, and a cap as a consensus parameter is a hardfork.

## Consequences

Hashing still costs one SHA-256 per leaf and per inner node, about a second of CPU for four million
transactions. Decoding such a block still allocates in amino, which this change does not touch.
