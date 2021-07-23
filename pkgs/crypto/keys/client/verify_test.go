package client

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/jaekwon/testify/assert"
)

func Test_runVerifyCmdBasic(t *testing.T) {
	cmd := command.NewMockCommand()
	assert.NotNil(t, cmd)

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// initialize test options
	opts := VerifyOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
		DocPath: "",
	}
	cmd.Options = opts

	fakeKeyName1 := "runVerifyCmd_Key1"
	// encPassword := "12345678"
	encPassword := ""
	testMsg := "some message"

	// add test account to keybase.
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
	assert.NoError(t, err)

	// sign test message.
	priv, err := kb.ExportPrivateKeyObject(fakeKeyName1, encPassword)
	assert.NoError(t, err)
	testSig, err := priv.Sign([]byte(testMsg))
	assert.NoError(t, err)
	testSigHex := hex.EncodeToString(testSig)

	// good signature passes test.
	cmd.SetIn(strings.NewReader(fmt.Sprintf(
		"%s\n", testMsg)))
	cmd.Args = []string{fakeKeyName1, testSigHex}
	err = runVerifyCmd(cmd)
	assert.NoError(t, err)

	// mutated bad signature fails test.
	testBadSig := testutils.MutateByteSlice(testSig)
	testBadSigHex := hex.EncodeToString(testBadSig)
	cmd.SetIn(strings.NewReader(fmt.Sprintf(
		"%s\n", testMsg)))
	cmd.Args = []string{fakeKeyName1, testBadSigHex}
	err = runVerifyCmd(cmd)
	assert.Error(t, err)
}
