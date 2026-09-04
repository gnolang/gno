package vm

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
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

// TestValidateGenesisRejectsNonRealmParamKeys covers realm-param keys that are
// not realm paths.
//
// InitGenesis writes realm params as "vm:"+key AFTER SetParams, so a key whose
// first segment is a vm submodule rather than a realm path overwrites a
// validated vm parameter. The genesis loader turns a [vm:p] section into
// exactly that shape, and it bypasses the scalar [vm] allowlist, which admits
// only chain_domain and sysnames_pkgpath.
func TestValidateGenesisRejectsNonRealmParamKeys(t *testing.T) {
	t.Parallel()

	mk := func(key string, value any) GenesisState {
		gs := DefaultGenesisState()
		gs.RealmParams = []params.Param{params.NewParam(key, value)}
		return gs
	}

	t.Run("vm submodule key is refused", func(t *testing.T) {
		t.Parallel()
		err := ValidateGenesis(mk("p:run_submitters", []string{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}))
		require.Error(t, err, "a [vm:p] genesis section must not be able to overwrite a vm parameter")
		assert.Contains(t, err.Error(), "not a realm path")
	})

	t.Run("malformed key is refused", func(t *testing.T) {
		t.Parallel()
		require.Error(t, ValidateGenesis(mk("nocolon", "x")))
	})

	t.Run("a real realm path is accepted", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, ValidateGenesis(mk("gno.land/r/demo/foo:bar", "x")))
	})

	t.Run("a second colon is refused", func(t *testing.T) {
		t.Parallel()
		// Only genesis can build this. A realm writing at runtime goes through
		// sys/params' prmkey, which panics on a colon in the name. It is also
		// ambiguous: this splits on the first colon, realmFromKey on the last.
		err := ValidateGenesis(mk("gno.land/r/demo/foo:a:b", "x"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not contain a colon")
	})
}
