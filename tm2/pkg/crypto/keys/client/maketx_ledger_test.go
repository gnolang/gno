package client

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/internal/ledger"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errPasswordRead = errors.New("password read")

// noInput fails any attempt to read a password.
type noInput struct{}

func (noInput) Read([]byte) (int, error) { return 0, errPasswordRead }

func TestExecSignAndBroadcastPassword(t *testing.T) {
	ledger.Discover = ledger.DiscoverMock
	t.Cleanup(func() { ledger.Discover = ledger.DiscoverDefault })

	home := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	_, err = kb.CreateLedger("ledger-key", keys.Secp256k1, crypto.Bech32AddrPrefix(), 0, 0)
	require.NoError(t, err)
	mnemonic, err := GenerateMnemonic(mnemonicEntropySize)
	require.NoError(t, err)
	_, err = kb.CreateAccount("local-key", mnemonic, "", "password", 0, 0)
	require.NoError(t, err)

	run := func(key string) (error, string) {
		var errOut bytes.Buffer
		io := commands.NewTestIO()
		io.SetIn(noInput{})
		io.SetErr(commands.WriteNopCloser(&errOut))
		cfg := &MakeTxCfg{
			RootCfg: &BaseCfg{BaseOptions: BaseOptions{
				Home:                  home,
				Remote:                "127.0.0.1:1", // nothing listens here
				InsecurePasswordStdin: true,
			}},
			GasWanted: 100000,
			GasFee:    "1000ugnot",
			Broadcast: true,
			Simulate:  SimulateSkip,
			ChainID:   "test",
		}
		err := ExecSignAndBroadcast(cfg, []string{key}, std.Tx{}, io)
		return err, errOut.String()
	}

	// A Ledger key signs on the device: no password is asked for, and the
	// command goes on to query the account.
	err, errOut := run("ledger-key")
	require.Error(t, err)
	assert.NotErrorIs(t, err, errPasswordRead)
	assert.NotContains(t, errOut, "Enter password.")

	// A local key still needs its password.
	err, errOut = run("local-key")
	assert.ErrorIs(t, err, errPasswordRead)
	assert.Contains(t, errOut, "Enter password.")

	// An unknown key fails before any prompt.
	err, errOut = run("missing-key")
	require.Error(t, err)
	assert.NotErrorIs(t, err, errPasswordRead)
	assert.NotContains(t, errOut, "Enter password.")
}
