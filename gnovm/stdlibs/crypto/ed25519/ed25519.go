package ed25519

import (
	"crypto/ed25519"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

//XXX: benchmark the real cost
const GasCostEd25519VerifyPerByte int64 = 1

func X_verify(m *gno.Machine, publicKey []byte, message []byte, signature []byte) bool {
	if m.GasMeter != nil {
		m.GasMeter.ConsumeGas(overflow.Mulp(int64(len(message)), GasCostEd25519VerifyPerByte), "ed25519.Verify")
	}
	return ed25519.Verify(publicKey, message, signature)
}
