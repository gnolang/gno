package sha256

import (
	"crypto/sha256"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// XXX: benchmark the real cost
const GasCostSha256PerByte int64 = 1

func X_sum256(m *gno.Machine, data []byte) [32]byte {
	if m.GasMeter != nil {
		m.GasMeter.ConsumeGas(overflow.Mulp(int64(len(data)), GasCostSha256PerByte), "sha256.Sum256")
	}
	return sha256.Sum256(data)
}
