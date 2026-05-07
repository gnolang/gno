package sha256

import (
	"crypto/sha256"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// GasCostSha256PerBlock is the gas cost per 64-byte block processed by SHA-256.
// SHA-256 processes data in 64-byte blocks; padding always adds at least 1 block.
// Calibrated via BenchmarkSha256Sum256: ~18-20 ns/block on Apple M5.
const GasCostSha256PerBlock int64 = 20

func X_sum256(m *gno.Machine, data []byte) [32]byte {
	if m.GasMeter != nil {
		nBlocks := int64(len(data))/64 + 1
		m.GasMeter.ConsumeGas(overflow.Mulp(nBlocks, GasCostSha256PerBlock), "sha256.Sum256")
	}
	return sha256.Sum256(data)
}
