package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/ledger"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// digestSignEnv is a keybase holding one key, and a tx file that key is to
// sign. The memo stands in for the oversized field, since this package has no
// message type carrying a file body.
type digestSignEnv struct {
	kbHome   string
	keyName  string
	password string
	txPath   string
	tx       std.Tx
}

func newDigestSignEnv(t *testing.T, memo string) digestSignEnv {
	t.Helper()

	const (
		keyName  = "digest-key"
		password = "encrypt"
	)

	kbHome := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	info, err := kb.CreateAccount(keyName, generateTestMnemonic(t), "", password, 0, 0)
	require.NoError(t, err)

	tx := std.Tx{
		Fee:  std.Fee{GasWanted: 10, GasFee: std.Coin{Denom: "ugnot", Amount: 10}},
		Msgs: []std.Msg{bank.MsgSend{FromAddress: info.GetAddress()}},
		Memo: memo,
	}

	encoded, err := amino.MarshalJSON(tx)
	require.NoError(t, err)

	txPath := tempTxPath(t)
	require.NoError(t, os.WriteFile(txPath, encoded, 0o600))

	return digestSignEnv{
		kbHome:   kbHome,
		keyName:  keyName,
		password: password,
		txPath:   txPath,
		tx:       tx,
	}
}

func tempTxPath(t *testing.T) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "tx-*.json")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	return file.Name()
}

// run signs env's tx with the given extra arguments, returning what the command
// printed.
func (env digestSignEnv) run(t *testing.T, extra ...string) (string, error) {
	t.Helper()

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&out))
	io.SetIn(strings.NewReader(fmt.Sprintf("%s\n%s\n", env.password, env.password)))

	args := append([]string{
		"sign",
		"--insecure-password-stdin",
		"--home", env.kbHome,
		"--tx-path", env.txPath,
	}, extra...)
	args = append(args, env.keyName)

	err := NewRootCmdWithBaseConfig(io, BaseOptions{
		InsecurePasswordStdin: true,
		Home:                  env.kbHome,
		Quiet:                 true,
	}).ParseAndRun(ctx, args)

	return out.String(), err
}

// signature reads back the signature the command wrote into the tx file.
func (env digestSignEnv) signature(t *testing.T) std.Signature {
	t.Helper()

	raw, err := os.ReadFile(env.txPath)
	require.NoError(t, err)

	var signed std.Tx
	require.NoError(t, amino.UnmarshalJSON(raw, &signed))
	require.Len(t, signed.Signatures, 1)

	return signed.Signatures[0]
}

// A misspelled mode must not fall back to full, which would sign something the
// device then refuses -- the error the flag exists to avoid, reported as if the
// flag had not been passed.
func TestSignRejectsUnknownSignMode(t *testing.T) {
	t.Parallel()

	env := newDigestSignEnv(t, "hello")

	_, err := env.run(t, "--sign-mode", "hash")
	assert.ErrorContains(t, err, `invalid sign mode "hash"`)
}

// The mode signs the digest rendering, and the chain reads it as such.
func TestSignDigestModeSignsTheDigestRendering(t *testing.T) {
	t.Parallel()

	memo := strings.Repeat("a", std.MaxSignDocStringLen+1)
	env := newDigestSignEnv(t, memo)

	_, err := env.run(t, "--sign-mode", SignModeDigest)
	require.NoError(t, err)

	sig := env.signature(t)
	doc := env.tx.SignDoc("dev", 0, 0)

	rendering, err := std.VerifySignaturePayload(sig.PubKey, doc, sig.Signature)
	require.NoError(t, err)
	assert.Equal(t, std.PayloadRenderingDigest, rendering)
}

// Full mode is the default, so a signer who passes nothing keeps the device
// display they had before.
func TestSignDefaultsToFullRendering(t *testing.T) {
	t.Parallel()

	memo := strings.Repeat("a", std.MaxSignDocStringLen+1)
	env := newDigestSignEnv(t, memo)

	out, err := env.run(t)
	require.NoError(t, err)
	assert.NotContains(t, out, "sign mode digest")

	sig := env.signature(t)
	rendering, err := std.VerifySignaturePayload(sig.PubKey, env.tx.SignDoc("dev", 0, 0), sig.Signature)
	require.NoError(t, err)
	assert.Equal(t, std.PayloadRenderingCurrent, rendering)
}

// WHAT THE SIGNER IS TOLD. Comparing the digest against the device screen is
// the whole security of the mode, so the field and its digest have to be
// printed before the device asks for approval -- and in a form a person can
// actually compare, which is why the hex is grouped.
func TestSignDigestModeReportsWhatItHides(t *testing.T) {
	t.Parallel()

	memo := strings.Repeat("a", std.MaxSignDocStringLen+1)
	env := newDigestSignEnv(t, memo)

	out, err := env.run(t, "--sign-mode", SignModeDigest)
	require.NoError(t, err)

	_, fields, err := std.GetSignaturePayloadDigest(env.tx.SignDoc("dev", 0, 0))
	require.NoError(t, err)
	require.Len(t, fields, 1)

	assert.Contains(t, out, "memo")
	assert.Contains(t, out, fmt.Sprintf("%d bytes", len(memo)))
	assert.Contains(t, out, groupHex(fields[0].Sum))
}

// And it says so when it hides nothing, rather than leaving a signer to wonder
// which fields the silence covered.
func TestSignDigestModeReportsAnUntouchedTx(t *testing.T) {
	t.Parallel()

	env := newDigestSignEnv(t, "short memo")

	out, err := env.run(t, "--sign-mode", SignModeDigest)
	require.NoError(t, err)
	assert.Contains(t, out, "nothing to digest")
}

func TestGroupHex(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "48ed9b0f c9e890e6", groupHex("48ed9b0fc9e890e6"))
	assert.Equal(t, "48ed9b0f c9e8", groupHex("48ed9b0fc9e8"), "a short tail is kept")
	assert.Empty(t, groupHex(""))
}

// WHERE THE REMEDY IS PHRASED. The device layer classifies a refusal and names
// no flag, because gnoclient reaches the same code with no flags to offer. So
// gnokey, which does have one, is the place that turns the classification into
// an instruction -- and must not hand it to a signer who is already using it.
func TestExplainLedgerSignError(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		err      error
		mode     string
		wants    string
		wantsNot string
	}{
		{
			name:  "payload too large, full mode",
			err:   fmt.Errorf("%w: 20452 bytes", ledger.ErrPayloadTooLarge),
			mode:  SignModeFull,
			wants: "-sign-mode digest",
		},
		{
			name:  "large payload refused, full mode",
			err:   fmt.Errorf("%w: 8840 bytes", ledger.ErrLargePayloadRefused),
			mode:  SignModeFull,
			wants: "-sign-mode digest",
		},
		{
			name:  "too many JSON values, full mode",
			err:   fmt.Errorf("%w", ledger.ErrTooManyJSONValues),
			mode:  SignModeFull,
			wants: "split the transaction",
		},
		{
			// Already the smallest rendering gnokey can produce. Sending them
			// round the same loop is worse than saying nothing.
			name:     "payload too large, already in digest mode",
			err:      fmt.Errorf("%w: 20452 bytes", ledger.ErrPayloadTooLarge),
			mode:     SignModeDigest,
			wantsNot: "-sign-mode digest",
		},
		{
			name:     "a rejected prompt collects no advice",
			err:      errors.New("[APDU_CODE_COMMAND_NOT_ALLOWED] User Rejected"),
			mode:     SignModeFull,
			wantsNot: "sign-mode",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := explainLedgerSignError(tc.err, tc.mode)

			// Whatever is added, the original survives for errors.Is and for
			// the signer reading it.
			require.ErrorIs(t, got, tc.err)
			if tc.wants != "" {
				assert.ErrorContains(t, got, tc.wants)
			}
			if tc.wantsNot != "" {
				assert.NotContains(t, got.Error(), tc.wantsNot)
			}
		})
	}
}
