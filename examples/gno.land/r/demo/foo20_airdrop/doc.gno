// Package foo20_airdrop implements an airdrop mechanism for the Foo20 GRC20 in Gno.
//
// This package extends the functionality of `gno.land/p/demo/airdrop` to support Foo20 token distribution.
// It enables secure and efficient token distribution to recipients using Merkle proofs.
//
// Usage:
// To execute an airdrop, use the claim function and provide the necessary proof and recipient data.
//
// const { MerkleTree } = require("merkletreejs");
// const crypto = require("crypto");
// const SHA256 = require("crypto-js/sha256");
// const secp256k1 = require("secp256k1");
// const { bech32 } = require("bech32");
//
// Example of using the ClaimJSON function:
//
// // This function reproduce the output of std.DerivePkgAddr
//
//	function derive_pkg_addr(str) {
//	  const hash = crypto
//	    .createHash("sha256")
//	    .update("pkgPath:" + str)
//	    .digest();
//
//	  let hash_bytes = new Uint8Array(hash).slice(0, 20);
//
//	  const words = bech32.toWords(hash_bytes);
//	  return bech32.encode("g", words);
//	}
//
// const addresses = Array.from({ length: 10 }, (_, i) =>
//
//	derive_pkg_addr(`gno.land/r/test/${i}`),
//
// );
//
// const leaves = addresses.map((addr) =>
//
//	SHA256(JSON.stringify({ address: addr, amount: "10000" })),
//
// );
//
// const tree = new MerkleTree(leaves, SHA256);
// const root = tree.getRoot().toString("hex");
//
//	const proof = tree.getProof(leaves[5]).map((p) => ({
//	  data: p.data.toString("hex"),
//	  position: p.position === "left" ? 1 : 0,
//	}));
//
// console.log("proof", proof); // This as JSON can be sent
//
// Data Structure:
// This package relies on the AirdropData structure from `gno.land/p/demo/airdrop`:
//
// Reference Implementation:
// The `foo20_airdrop` package builds on the `airdrop` package's foundation. Refer to its documentation for
// additional details: gno.land/p/demo/airdrop/doc.gno
//
// Notes:
// - Ensure your Merkle tree and proofs are generated off-chain using a compatible library such as `merkletreejs`.
// - Hash function: SHA256
package foo20_airdrop
