// Package bn254 exposes the EVM precompile operations over the BN254 curve as
// native bindings. Input and output layouts mirror the Ethereum precompiles
// exactly (EIP-196 for G1 add/mul, EIP-197 for the pairing check), so gno
// contracts that want the familiar "pack into a byte buffer, read a byte
// buffer out" precompile pattern can use this directly.
//
// Each function returns nil on invalid input (e.g. wrong length, point not on
// curve, coordinate not reduced) — matching staticcall failure in the EVM.
package bn254

// G1AddInputSize is the input byte length for G1Add: x1|y1|x2|y2, each
// 32-byte big-endian.
const G1AddInputSize = 128

// G1MulInputSize is the input byte length for G1Mul: x|y|scalar, each
// 32-byte big-endian.
const G1MulInputSize = 96

// G1OutputSize is the output byte length for G1Add and G1Mul: x|y, each
// 32-byte big-endian.
const G1OutputSize = 64

// PairingChunkSize is the stride of a single (G1, G2) pair in the
// PairingCheck input: G1 (64 bytes: x|y) followed by G2 (128 bytes, field
// extension in imaginary-first order: x_imag|x_real|y_imag|y_real).
const PairingChunkSize = 192

// PairingOutputSize is the output byte length for PairingCheck:
// 32-byte big-endian 0x...01 (success) or 0x...00 (failure).
const PairingOutputSize = 32

// G1Add returns the affine sum of two G1 points.
//
// input must be exactly G1AddInputSize bytes. The output is G1OutputSize
// bytes, or nil if input is malformed / coordinates are not reduced / points
// are not on the curve.
func G1Add(input []byte) []byte { return g1Add(input) }

// G1Mul returns scalar * G1 in affine form.
//
// input must be exactly G1MulInputSize bytes (x|y|scalar). The output is
// G1OutputSize bytes, or nil on malformed input / invalid point. The scalar
// is reduced modulo the group order internally; it does not need to be below
// the scalar field modulus.
func G1Mul(input []byte) []byte { return g1Mul(input) }

// PairingCheck evaluates the product of pairings and returns 1 iff the
// product equals the identity in GT.
//
// input must be a multiple of PairingChunkSize bytes (possibly zero). Each
// 192-byte chunk is (g1.x, g1.y, g2.x_imag, g2.x_real, g2.y_imag, g2.y_real),
// all 32-byte big-endian. The output is PairingOutputSize bytes:
// 0x...01 on success (product == 1), 0x...00 otherwise. Returns nil on
// malformed input / invalid points / coordinates not reduced.
func PairingCheck(input []byte) []byte { return pairingCheck(input) }

func g1Add(input []byte) []byte         // injected
func g1Mul(input []byte) []byte         // injected
func pairingCheck(input []byte) []byte  // injected
