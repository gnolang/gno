// Package airdrop implements a Merkle airdrop mechanism in Gno.
//
// A Merkle airdrop is a secure and efficient way to distribute tokens or rewards to a list of recipients
// using a Merkle tree for proof verification. This implementation is compliant with the `merkletreejs`
// JavaScript package, ensuring compatibility with proofs generated off-chain.
//
// Note:
// This package expects data to adhere to the structure defined in `gno.land/p/demo/airdrop.AirdropData`.
//
// Compatibility:
// - Proofs and trees generated using the `merkletreejs` package (https://www.npmjs.com/package/merkletreejs) are supported.
// - Uses SHA256 as the hash function with a sorted-pair configuration.
//
// Reference:
// For more information on Merkle trees, refer to the documentation for the `merkletreejs` package:
// https://www.npmjs.com/package/merkletreejs
package grc20_merkle_airdrop
