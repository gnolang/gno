package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/internal/ledger"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The service and the use the test proofs are for, which runKeysCmd passes
// unless a test overrides them.
const (
	testURL     = "https://gno.land"
	testPurpose = "auth/SignIn"
)

var proofScopeArgs = []string{"--url", testURL, "--purpose", testPurpose}

// verifySignDataOutput checks out with std.VerifySignDataProof and returns what
// the document says.
func verifySignDataOutput(t *testing.T, out []byte) std.SignedData {
	t.Helper()

	signed, err := std.VerifySignDataProof(out)
	require.NoError(t, err)

	return signed
}

type signDataTest struct {
	kbHome   string
	keyName  string
	password string
	address  crypto.Address
}

func newSignDataTest(t *testing.T) signDataTest {
	t.Helper()

	st := signDataTest{kbHome: t.TempDir(), keyName: "generated-key", password: "encrypt"}

	kb, err := keys.NewKeyBaseFromDir(st.kbHome)
	require.NoError(t, err)

	info, err := kb.CreateAccount(st.keyName, generateTestMnemonic(t), "", st.password, 0, 0)
	require.NoError(t, err)
	st.address = info.GetAddress()

	return st
}

// runKeysCmd runs the subcommand args[0] with proofScopeArgs and the rest of
// args, on the keybase in home and with stdin as its input, and returns its
// standard output.
func runKeysCmd(t *testing.T, home, stdin string, args ...string) ([]byte, error) {
	t.Helper()

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(stdin))
	io.SetOut(commands.WriteNopCloser(&out))

	cmd := NewRootCmdWithBaseConfig(io, BaseOptions{Home: home})

	cmdArgs := append([]string{args[0], "--home", home}, proofScopeArgs...)
	err := cmd.ParseAndRun(ctx, append(cmdArgs, args[1:]...))

	return out.Bytes(), err
}

// run runs sign-data with args (the key name last, if any) and returns its
// standard output.
func (st signDataTest) run(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()

	return runKeysCmd(t, st.kbHome, st.password+"\n", append([]string{"sign-data", "--insecure-password-stdin"}, args...)...)
}

func TestSignData(t *testing.T) {
	t.Parallel()

	t.Run("no key provided", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)

		_, err := st.run(t, "--chainid", "dev", "--data", "hello")
		assert.ErrorIs(t, err, flag.ErrHelp)
	})

	t.Run("non-existing key", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)

		_, err := st.run(t, "--chainid", "dev", "--data", "hello", "TotallyExistingKey")
		assert.True(t, keyerror.IsErrKeyNotFound(err))
	})

	// A proof bound to a network nobody chose is refused by every verifier on
	// another one, so the network is never assumed.
	t.Run("no chain id", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)

		_, err := st.run(t, "--data", "hello", st.keyName)
		assert.ErrorContains(t, err, "chainid")
	})

	// Without them a proof made for one service, or one use, passes at another.
	t.Run("no url", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)

		_, err := st.run(t, "--chainid", "dev", "--url", "", "--data", "hello", st.keyName)
		assert.ErrorContains(t, err, "-url")
	})

	t.Run("no purpose", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)

		_, err := st.run(t, "--chainid", "dev", "--purpose", "", "--data", "hello", st.keyName)
		assert.ErrorContains(t, err, "-purpose")
	})

	t.Run("no data", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)

		_, err := st.run(t, "--chainid", "dev", st.keyName)
		assert.ErrorContains(t, err, "data")
	})

	t.Run("both data and data-path", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)
		path := filepath.Join(t.TempDir(), "data.txt")
		require.NoError(t, os.WriteFile(path, []byte("hello"), 0o644))

		_, err := st.run(t, "--chainid", "dev", "--data", "hello", "--data-path", path, st.keyName)
		assert.ErrorContains(t, err, "data")
	})

	t.Run("data that is not UTF-8", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)
		path := filepath.Join(t.TempDir(), "data.bin")
		require.NoError(t, os.WriteFile(path, []byte{0xff, 0xfe}, 0o644))

		_, err := st.run(t, "--chainid", "dev", "--data-path", path, st.keyName)
		assert.ErrorContains(t, err, "UTF-8")
	})

	t.Run("signs data", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)

		out, err := st.run(t, "--chainid", "gnoland-1", "--data", "prove this wallet is yours", st.keyName)
		require.NoError(t, err)

		assert.Equal(t, std.SignedData{
			ChainID: "gnoland-1",
			URL:     testURL,
			Purpose: testPurpose,
			Signer:  st.address,
			Data:    "prove this wallet is yours",
		}, verifySignDataOutput(t, out))
	})

	// The file is signed as it is, a final newline included: a verifier
	// compares the data with the challenge it handed out, byte for byte.
	t.Run("signs a file as it is", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)
		path := filepath.Join(t.TempDir(), "data.txt")
		require.NoError(t, os.WriteFile(path, []byte("line one\nline two\n"), 0o644))

		out, err := st.run(t, "--chainid", "dev", "--data-path", path, st.keyName)
		require.NoError(t, err)

		assert.Equal(t, "line one\nline two\n", verifySignDataOutput(t, out).Data)
	})

	t.Run("writes the output document", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)
		path := filepath.Join(t.TempDir(), "proof.json")

		out, err := st.run(t, "--chainid", "dev", "--data", "hello", "--output-document", path, st.keyName)
		require.NoError(t, err)
		assert.NotContains(t, string(out), `"signed"`)

		doc, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "hello", verifySignDataOutput(t, doc).Data)
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)
		st.password = "not the password"

		_, err := st.run(t, "--chainid", "dev", "--data", "hello", st.keyName)
		assert.Error(t, err)
	})
}

// ledgerDevice is a mock Ledger holding priv.
type ledgerDevice struct {
	ledger.MockLedger
	priv secp256k1.PrivKeySecp256k1
}

func (d ledgerDevice) GetPublicKeySECP256K1([]uint32) ([]byte, error) {
	_, pub := btcec.PrivKeyFromBytes(d.priv[:])
	return pub.SerializeUncompressed(), nil
}

func (d ledgerDevice) GetAddressPubKeySECP256K1([]uint32, string) ([]byte, string, error) {
	_, pub := btcec.PrivKeyFromBytes(d.priv[:])
	return pub.SerializeCompressed(), d.priv.PubKey().Address().String(), nil
}

func (d ledgerDevice) SignSECP256K1(_ []uint32, msg []byte, _ byte) ([]byte, error) {
	priv, _ := btcec.PrivKeyFromBytes(d.priv[:])
	hash := sha256.Sum256(msg)
	return ecdsa.Sign(priv, hash[:]).Serialize(), nil
}

// A Ledger signs with the device plugged in, so a proof by another device must
// not come out as the key's.
func TestSignData_AnotherLedger(t *testing.T) {
	plugged := ledgerDevice{priv: secp256k1.GenPrivKey()}
	ledger.Discover = func() (ledger.SECP256K1, error) { return plugged, nil }
	t.Cleanup(func() { ledger.Discover = ledger.DiscoverDefault })

	kbHome := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)
	info, err := kb.CreateLedger("ledger-key", keys.Secp256k1, crypto.Bech32AddrPrefix(), 0, 0)
	require.NoError(t, err)

	out, err := runKeysCmd(t, kbHome, "", "sign-data", "--chainid", "dev", "--data", "hello", "ledger-key")
	require.NoError(t, err)
	assert.Equal(t, info.GetAddress(), verifySignDataOutput(t, out).Signer)

	plugged = ledgerDevice{priv: secp256k1.GenPrivKey()}

	out, err = runKeysCmd(t, kbHome, "", "sign-data", "--chainid", "dev", "--data", "hello", "ledger-key")
	assert.ErrorContains(t, err, "connect the device that holds it")
	assert.NotContains(t, string(out), `"signed"`)
}
