package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/secp256k1"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/jaekwon/testify/assert"
)

func Test_addAppBasic(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// initialize test options
	opts := AddOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
	}

	cmd.SetIn(strings.NewReader("test1234\ntest1234\n"))
	err := addApp(cmd, []string{"keyname1"}, opts)
	assert.NoError(t, err)

	cmd.SetIn(strings.NewReader("test1234\ntest1234\n"))
	err = addApp(cmd, []string{"keyname1"}, opts)
	assert.Error(t, err)

	cmd.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))
	err = addApp(cmd, []string{"keyname1"}, opts)
	assert.NoError(t, err)
}

var test2_address = "g1fupfatmln5844rjafzp6d2vc825vav2x2kzaac"
var test2_mnemonic = "hair stove window more scrap patient endorse left early pear lawn school loud divide vibrant family still bulk lyrics firm plate media critic dove"
var test2_pubkey_bech32 = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqg5y7u93gpzug38k2p8s8322zpdm96t0ch87ax88sre4vnclz2jcy8uyhst"

func Test_addPublicKey(t *testing.T) {

	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	opts := AddOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},

		PublicKey: test2_pubkey_bech32, // test2 account
	}
	err := addApp(cmd, []string{"test2"}, opts)
	assert.NoError(t, err)

}

func Test_addAppRecover(t *testing.T) {

	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	opts := AddOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},

		Recover: true, // init test2 account
	}

	test2_name := "test2"
	test2_passphrase := "gn0rocks!"

	cmd.SetIn(strings.NewReader(test2_passphrase + "\n" + test2_passphrase + "\n" + test2_mnemonic + "\n"))

	err := addApp(cmd, []string{test2_name}, opts)
	assert.NoError(t, err)

	kb, err2 := keys.NewKeyBaseFromDir(opts.Home)
	assert.NoError(t, err2)

	infos, err3 := kb.List()
	assert.NoError(t, err3)

	info := infos[0]

	keypub := info.GetPubKey()
	keypub = keypub.(secp256k1.PubKeySecp256k1)

	s := fmt.Sprintf("%s", keypub)
	assert.Equal(t, s, test2_pubkey_bech32)
}
