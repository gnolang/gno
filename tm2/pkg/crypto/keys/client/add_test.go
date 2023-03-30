package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_execAddBasic(t *testing.T) {
	t.Parallel()

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	cfg := &addCfg{
		rootCfg: &baseCfg{
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

	cfg := &addCfg{
		rootCfg: &baseCfg{
			BaseOptions: BaseOptions{
				Home: kbHome,
			},
		},
		publicKey: test2PubkeyBech32, // test2 account
	}

	if err := execAdd(cfg, []string{"test2"}, nil); err != nil {
		t.Fatalf("unable to execute add cmd, %v", err)
	}
}

func Test_execAddRecover(t *testing.T) {
	t.Parallel()

	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	cfg := &addCfg{
		rootCfg: &baseCfg{
			BaseOptions: BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			},
		},
		recover: true, // init test2 account
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
