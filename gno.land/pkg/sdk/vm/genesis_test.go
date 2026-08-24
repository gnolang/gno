package vm

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// TestInitGenesisWarnsOnEmptyRunSubmitters pins the boot-time warning for the
// one configuration that cannot be repaired on-chain.
//
// run_submitters gates MsgRun and fails closed, and govdao proposal creation is
// MsgRun-only, so an empty list disables the very mechanism that would populate
// it. That state cannot be rejected in Validate, because the field's zero value
// IS empty and rejecting it would make DefaultParams() invalid — so a loud log
// at genesis is the strongest available signal. Worth pinning, because a warning
// nobody emits is worse than none: it reads as coverage.
//
// Both directions are asserted. A seeded list must boot QUIETLY, otherwise the
// warning fires on every correctly-configured chain and gets tuned out.
func TestInitGenesisWarnsOnEmptyRunSubmitters(t *testing.T) {
	run := func(t *testing.T, submitters []crypto.Address) string {
		t.Helper()
		env := setupTestEnv()

		var buf bytes.Buffer
		ctx := env.ctx.WithLogger(slog.New(slog.NewTextHandler(&buf, nil)))

		gs := DefaultGenesisState()
		gs.Params.RunSubmitters = submitters
		env.vmk.InitGenesis(ctx, gs)
		return buf.String()
	}

	t.Run("empty list warns and names the fix", func(t *testing.T) {
		out := run(t, nil)
		require.Contains(t, out, "run_submitters is empty",
			"an unrecoverable configuration must not boot silently")
		assert.Contains(t, out, "gnogenesis params set vm.run_submitters",
			"the warning must name the command that fixes it")
	})

	t.Run("seeded list boots quietly", func(t *testing.T) {
		out := run(t, []crypto.Address{crypto.AddressFromPreimage([]byte("runner"))})
		assert.NotContains(t, out, "run_submitters is empty",
			"a correctly configured chain must not be warned, or the warning gets ignored")
	})
}
