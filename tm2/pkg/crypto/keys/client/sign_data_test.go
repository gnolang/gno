package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signDataOutput is what sign-data writes, in the shape of Adena's
// SignArbitrary response.
type signDataOutput struct {
	Signed    json.RawMessage `json:"signed"`
	Signature struct {
		PubKey struct {
			TypeURL string `json:"typeUrl"`
			Value   string `json:"value"`
		} `json:"pubKey"`
		Signature string `json:"signature"`
	} `json:"signature"`
}

// verifySignDataOutput checks out with std.VerifySignData and returns what the
// document says.
func verifySignDataOutput(t *testing.T, out []byte) std.SignedData {
	t.Helper()

	var o signDataOutput
	require.NoError(t, json.Unmarshal(out, &o))
	assert.Equal(t, "/tm.PubKeySecp256k1", o.Signature.PubKey.TypeURL)

	raw, err := base64.StdEncoding.DecodeString(o.Signature.PubKey.Value)
	require.NoError(t, err)
	require.Len(t, raw, secp256k1.PubKeySecp256k1Size)

	var pub secp256k1.PubKeySecp256k1
	copy(pub[:], raw)

	sig, err := base64.StdEncoding.DecodeString(o.Signature.Signature)
	require.NoError(t, err)

	signed, err := std.VerifySignData(o.Signed, pub, sig)
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

// run runs sign-data with args (the key name last, if any) and returns its
// standard output.
func (st signDataTest) run(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(st.password + "\n"))
	io.SetOut(commands.WriteNopCloser(&out))

	cmd := NewRootCmdWithBaseConfig(io, BaseOptions{InsecurePasswordStdin: true, Home: st.kbHome})

	err := cmd.ParseAndRun(ctx, append([]string{"sign-data", "--insecure-password-stdin", "--home", st.kbHome}, args...))

	return out.Bytes(), err
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

		signed := verifySignDataOutput(t, out)
		assert.Equal(t, "gnoland-1", signed.ChainID)
		assert.Equal(t, st.address, signed.Signer)
		assert.Equal(t, "prove this wallet is yours", signed.Data)
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
