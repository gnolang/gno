package ed25519

import (
	"crypto/ed25519"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// Sha512BlockSize is the block size of SHA-512 in bytes (used internally by Ed25519).
const Sha512BlockSize int64 = 128

// GasCostEd25519VerifyBase is the fixed gas cost for the elliptic curve operations
// in Ed25519 verification. This dominates for small messages.
// Calibrated via BenchmarkEd25519Verify: ~25,500 ns for 64-byte msg on Apple M5.
const GasCostEd25519VerifyBase int64 = 25000

// GasCostEd25519VerifyPerBlock is the marginal gas cost per 128-byte SHA-512 block
// for hashing the message during Ed25519 verification.
// Calibrated via BenchmarkEd25519Verify: ~63 ns/block on Apple M5.
const GasCostEd25519VerifyPerBlock int64 = 60

func X_verify(m *gno.Machine, publicKey []byte, message []byte, signature []byte) bool {
	if m.GasMeter != nil {
		nBlocks := int64(len(message))/Sha512BlockSize + 1
		gas := overflow.Addp(GasCostEd25519VerifyBase, overflow.Mulp(nBlocks, GasCostEd25519VerifyPerBlock))
		m.GasMeter.ConsumeGas(gas, "ed25519.Verify")
	}
	return ed25519.Verify(publicKey, message, signature)
}
