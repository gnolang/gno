// Package modexp exposes arbitrary-precision modular exponentiation as a
// native binding, matching the semantics of the EIP-198 (0x05) precompile.
//
// Use this when a gno contract needs a single modular reduction or a modular
// inverse over a large prime — both reducible to this primitive without
// pulling in a full big-integer stdlib.
package modexp

// ModExp returns (base^exp) mod modulus. Inputs are big-endian byte slices
// of arbitrary length. The output has len(modulus) bytes, left-padded with
// zeros.
//
// Matches EIP-198 semantics:
//   - If modulus is zero or empty, returns len(modulus) zero bytes.
//   - If exp is zero or empty, returns 1 (or 0 if modulus is 1).
//   - Inputs are taken as non-negative big-endian integers with leading zeros
//     tolerated.
func ModExp(base, exp, modulus []byte) []byte { return modExp(base, exp, modulus) }

func modExp(base, exp, modulus []byte) []byte // injected
