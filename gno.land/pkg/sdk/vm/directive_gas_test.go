package vm

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Validation reads every .gno file end to end -- FindDirectiveComment must,
// since a pragma can sit on any declaration -- so a package that fails
// validation still costs the validator a full scan of its source. The
// preprocess charge is therefore taken before validation: without it, the only
// price of that scan is the generic per-tx-size fee, and a rejected package is
// the cheapest way to buy validator CPU.
func TestAddPkgChargesPreprocessGasBeforeValidation(t *testing.T) {
	env := setupTestEnv()
	// Non-genesis block: genesis uses an infinite gas meter.
	ctx := env.ctx.WithBlockHeader(&bft.Header{Height: int64(1)})

	addr := crypto.AddressFromPreimage([]byte("test1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))

	const pkgPath = "gno.land/r/test/rejected"
	// Large body so the charge is unmistakably byte-proportional, and a
	// directive at the very end so the scan cannot stop early.
	body := "package rejected\n\n" +
		strings.Repeat("// filler\nvar _ = 1\n", 2000) +
		"//go:noinline\nfunc F() {}\n"

	msg := NewMsgAddPackage(addr, pkgPath, []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "rejected.gno", Body: body},
	})

	gctx := auth.SetGasMeter(ctx.WithMode(sdk.RunTxModeDeliver), 100_000_000_000)
	gctx = env.vmk.MakeGnoTransactionStore(gctx)
	before := gctx.GasMeter().GasConsumed()

	err := env.vmk.AddPackage(gctx, msg)
	// The rejection reason is asserted in the gnovm and txtar tests; here the
	// point is only that a rejected package was still charged for its bytes.
	require.Error(t, err, "the package carries a directive and must be rejected")

	charged := gctx.GasMeter().GasConsumed() - before
	srcBytes := int64(len(body))
	params := env.vmk.GetParams(gctx)
	assert.GreaterOrEqual(t, charged, params.PreprocessGasPerByte*srcBytes,
		"a rejected package must still pay for the source bytes validation had to read")
}
