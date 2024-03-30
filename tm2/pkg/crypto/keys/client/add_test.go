package client

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestMnemonic generates a random mnemonic
func generateTestMnemonic(t *testing.T) string {
	t.Helper()

	entropy, entropyErr := bip39.NewEntropy(256)
	require.NoError(t, entropyErr)

	mnemonic, mnemonicErr := bip39.NewMnemonic(entropy)
	require.NoError(t, mnemonicErr)

	return mnemonic
}

func TestAdd_Base_Add(t *testing.T) {
	t.Parallel()

	t.Run("valid key addition, generated mnemonic", func(t *testing.T) {
		t.Parallel()

		// TODO
	})

	t.Run("valid key addition, provided mnemonic", func(t *testing.T) {
		t.Parallel()

		// TODO
	})

	t.Run("no overwrite permission", func(t *testing.T) {
		t.Parallel()

		// TODO
	})
}

func generateDerivationPaths(count int) []string {
	paths := make([]string, count)

	for i := 0; i < count; i++ {
		paths[i] = fmt.Sprintf("44'/118'/0'/0/%d", i)
	}

	return paths
}

func TestAdd_Derive(t *testing.T) {
	t.Parallel()

	t.Run("valid address derivation", func(t *testing.T) {
		t.Parallel()

		var (
			mnemonic = generateTestMnemonic(t)
			paths    = generateDerivationPaths(10)

			dummyPass = "dummy-pass"
		)

		kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
		require.NotNil(t, kbHome)
		t.Cleanup(kbCleanUp)

		cfg := &AddCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					InsecurePasswordStdin: true,
					Home:                  kbHome,
				},
			},
			Recover:        true,
			DerivationPath: paths,
		}

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(dummyPass + "\n" + dummyPass + "\n" + mnemonic + "\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		require.NoError(t,
			execAdd(
				cfg,
				[]string{
					"example-key",
				},
				io,
			),
		)

		// Verify the addresses are derived correctly
		expectedAccounts := generateAccounts(
			mnemonic,
			paths,
		)

		// Grab the output
		deriveOutput := mockOut.String()

		for _, expectedAccount := range expectedAccounts {
			assert.Contains(t, deriveOutput, expectedAccount.String())
		}
	})

	t.Run("malformed derivation path", func(t *testing.T) {
		t.Parallel()

		var (
			mnemonic  = generateTestMnemonic(t)
			dummyPass = "dummy-pass"
		)

		kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
		require.NotNil(t, kbHome)
		t.Cleanup(kbCleanUp)

		cfg := &AddCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					InsecurePasswordStdin: true,
					Home:                  kbHome,
				},
			},
			Recover: true,
			DerivationPath: []string{
				"malformed path",
			},
		}

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(dummyPass + "\n" + dummyPass + "\n" + mnemonic + "\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		require.ErrorIs(
			t, execAdd(
				cfg,
				[]string{
					"example-key",
				},
				io,
			),
			errInvalidDerivationPath,
		)
	})

	t.Run("invalid derivation path", func(t *testing.T) {
		t.Parallel()

		var (
			mnemonic  = generateTestMnemonic(t)
			dummyPass = "dummy-pass"
		)

		kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
		require.NotNil(t, kbHome)
		t.Cleanup(kbCleanUp)

		cfg := &AddCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					InsecurePasswordStdin: true,
					Home:                  kbHome,
				},
			},
			Recover: true,
			DerivationPath: []string{
				"44'/500'/0'/0/0", // invalid coin type
			},
		}

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(dummyPass + "\n" + dummyPass + "\n" + mnemonic + "\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		require.ErrorIs(
			t, execAdd(
				cfg,
				[]string{
					"example-key",
				},
				io,
			),
			errInvalidDerivationPath,
		)
	})
}

func Test_execAddBasic(t *testing.T) {
	t.Parallel()

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	cfg := &AddCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			},
		},
	}

	keyName := "keyname1"

	io := commands.NewTestIO()
	io.SetIn(strings.NewReader("test1234\ntest1234\n"))

	// Create a new key
	if err := execAdd(cfg, []string{keyName}, io); err != nil {
		t.Fatalf("unable to execute add cmd, %v", err)
	}

	io.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))

	// Confirm overwrite
	if err := execAdd(cfg, []string{keyName}, io); err != nil {
		t.Fatalf("unable to execute add cmd, %v", err)
	}
}

var (
	test2Mnemonic     = "hair stove window more scrap patient endorse left early pear lawn school loud divide vibrant family still bulk lyrics firm plate media critic dove"
	test2PubkeyBech32 = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqg5y7u93gpzug38k2p8s8322zpdm96t0ch87ax88sre4vnclz2jcy8uyhst"
)

func Test_execAddPublicKey(t *testing.T) {
	t.Parallel()

	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	cfg := &AddBech32Cfg{
		RootCfg: &AddCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home: kbHome,
				},
			},
		},
		PublicKey: test2PubkeyBech32, // test2 account
	}

	if err := execAddBech32(cfg, []string{"test2"}, nil); err != nil {
		t.Fatalf("unable to execute add cmd, %v", err)
	}
}

func Test_execAddRecover(t *testing.T) {
	t.Parallel()

	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	cfg := &AddCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			},
		},
		Recover: true, // init test2 account
	}

	test2Name := "test2"
	test2Passphrase := "gn0rocks!"

	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(test2Passphrase + "\n" + test2Passphrase + "\n" + test2Mnemonic + "\n"))

	if err := execAdd(cfg, []string{test2Name}, io); err != nil {
		t.Fatalf("unable to execute add cmd, %v", err)
	}

	kb, err2 := keys.NewKeyBaseFromDir(kbHome)
	assert.NoError(t, err2)

	infos, err3 := kb.List()
	assert.NoError(t, err3)

	info := infos[0]

	keypub := info.GetPubKey()
	keypub = keypub.(secp256k1.PubKeySecp256k1)

	s := fmt.Sprintf("%s", keypub)
	assert.Equal(t, s, test2PubkeyBech32)
}
