package client

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_execVerify(t *testing.T) {
	t.Parallel()

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// initialize test options
	cfg := &VerifyCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
		DocPath: "",
	}

	io := commands.NewTestIO()

	fakeKeyName1 := "verifyApp_Key1"
	// encPassword := "12345678"
	encPassword := ""
	testMsg := "some message"

	// add test account to keybase.
	kb, err := keys.NewKeyBaseFromDir(kbHome)
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
	args := []string{fakeKeyName1, testSigHex}
	io.SetIn(
		strings.NewReader(
			fmt.Sprintf("%s\n", testMsg),
		),
	)
	err = execVerify(cfg, args, io)
	assert.NoError(t, err)

	// mutated bad signature fails test.
	testBadSig := testutils.MutateByteSlice(testSig)
	testBadSigHex := hex.EncodeToString(testBadSig)
	args = []string{fakeKeyName1, testBadSigHex}
	io.SetIn(
		strings.NewReader(
			fmt.Sprintf("%s\n", testMsg),
		),
	)
	err = execVerify(cfg, args, io)
	assert.Error(t, err)
}
