// Package mldsa65 implements a tm2 account key type backed by ML-DSA-65, the
// module-lattice signature scheme standardised in FIPS 204 (security
// category 3).
//
// # Why
//
// Every key tm2 accepts today, ed25519 and secp256k1, rests on the hardness
// of the elliptic curve discrete logarithm problem. A cryptanalytic break of
// ECDLP, or a large enough quantum computer running Shor's algorithm, would
// let an attacker derive the private key behind any public key that has ever
// signed a transaction. ML-DSA rests on lattice problems (Module-LWE and
// Module-SIS) that are not known to fall to either, so an account holding
// this key type is a hedge: its funds stay safe if ECDLP goes.
//
// # Sizes
//
//   - public key: 1952 bytes (ed25519: 32, secp256k1: 33)
//   - signature: 3309 bytes (ed25519: 64, secp256k1: 64)
//   - private key: stored as the 32-byte FIPS 204 seed, expanded on use
//
// Addresses are derived as for the other keys, the SHA256-20 of the raw
// public key bytes, so an address is still 20 bytes.
//
// Signing is deterministic with an empty context string. Implementation:
// github.com/cloudflare/circl/sign/mldsa/mldsa65.
//
// # Not in this package
//
// This is an account key type only. Deliberately left for later work:
//
//   - validator keys: ValidatorParams.PubKeyTypes does not list it, and
//     consensus votes keep using ed25519;
//   - the p2p handshake (SecretConnection), which stays on ed25519 and X25519;
//   - a message to rotate an existing account to a new key, so today only a
//     fresh address can use it;
//   - guidance on hybrid multisig (for example a 2-of-2 of ed25519 and
//     ML-DSA-65), which multisig.PubKeyMultisigThreshold can already express;
//   - tx size limits: a single signature plus its public key add about 5.3 KB
//     to a transaction, which fee and block size parameters should account for.
package mldsa65
