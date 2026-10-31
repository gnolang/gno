package client

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_execList(t *testing.T) {
	// Prepare some keybases
	kbHome1, cleanUp1 := testutils.NewTestCaseDir(t)
	kbHome2, cleanUp2 := testutils.NewTestCaseDir(t)
	defer cleanUp1()
	defer cleanUp2()
	// leave home1 and home2 empty

	// initialize test keybase.
	kb, err := keys.NewKeyBaseFromDir(kbHome2)
	assert.NoError(t, err)
	_, err = kb.CreateAccount("something", testMnemonic, "", "", 0, 0)
	assert.NoError(t, err)

	testData := []struct {
		name    string
		kbDir   string
		args    []string
		wantErr bool
	}{
		{"invalid keybase", "/dev/null", []string{}, true},
		{"keybase: empty", kbHome1, []string{}, false},
		{"keybase: w/key", kbHome2, []string{}, false},
	}
	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			rootCfg := &BaseCfg{
				BaseOptions: BaseOptions{
					Home: tt.kbDir,
				},
			}
			cfg := &ListCfg{
				RootCfg:         rootCfg,
				MultisigMembers: multisigMembersNone,
			}

			args := tt.args
			if err := execList(cfg, args, commands.NewTestIO()); (err != nil) != tt.wantErr {
				t.Errorf("listApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newMultisigListTestSetup(t *testing.T, numKeys int) (string, []secp256k1.PrivKeySecp256k1, multisig.PubKeyMultisigThreshold) {
	t.Helper()

	kbHome := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	privKeys := make([]secp256k1.PrivKeySecp256k1, numKeys)
	pubKeys := make([]crypto.PubKey, numKeys)
	for i := range privKeys {
		privKeys[i] = secp256k1.GenPrivKey()
		pubKeys[i] = privKeys[i].PubKey()
		require.NoError(t, kb.ImportPrivKey(t.Name()+"k"+string(rune('0'+i)), privKeys[i], ""))
	}

	msPub := multisig.NewPubKeyMultisigThreshold(1, pubKeys).(multisig.PubKeyMultisigThreshold)
	_, err = kb.CreateMulti("ms", msPub)
	require.NoError(t, err)

	return kbHome, privKeys, msPub
}

func Test_execList_MultisigNone(t *testing.T) {
	t.Parallel()

	kbHome, _, msPub := newMultisigListTestSetup(t, 3)

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&out))

	cfg := &ListCfg{
		RootCfg:         &BaseCfg{BaseOptions: BaseOptions{Home: kbHome}},
		MultisigMembers: multisigMembersNone,
	}

	require.NoError(t, execList(cfg, nil, io))

	output := out.String()
	assert.Contains(t, output, "pub: "+crypto.PubKeyToBech32(msPub))
	for _, pk := range msPub.PubKeys {
		assert.NotContains(t, output, "\n  "+pk.String())
	}
}

func Test_execList_MultisigShort(t *testing.T) {
	t.Parallel()

	const numKeys = 5
	kbHome, privKeys, msPub := newMultisigListTestSetup(t, numKeys)

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&out))

	cfg := &ListCfg{
		RootCfg:         &BaseCfg{BaseOptions: BaseOptions{Home: kbHome}},
		MultisigMembers: multisigMembersShort,
	}

	require.NoError(t, execList(cfg, nil, io))

	output := out.String()
	assert.Contains(t, output, "pub: "+crypto.PubKeyToBech32(msPub))

	// Count how many member pubkeys appear in the output.
	shown := 0
	for _, pk := range privKeys {
		if bytes.Contains([]byte(output), []byte("  "+pk.PubKey().String())) {
			shown++
		}
	}
	assert.Equal(t, multisigMembersShortLimit, shown)
	assert.Contains(t, output, "... and 2 more (use -multisig-members=full to see all)")
}

func Test_execList_MultisigFull(t *testing.T) {
	t.Parallel()

	kbHome, _, msPub := newMultisigListTestSetup(t, 3)

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&out))

	cfg := &ListCfg{
		RootCfg:         &BaseCfg{BaseOptions: BaseOptions{Home: kbHome}},
		MultisigMembers: multisigMembersFull,
	}

	require.NoError(t, execList(cfg, nil, io))

	output := out.String()
	assert.Contains(t, output, "pub: "+crypto.PubKeyToBech32(msPub))
	for _, pk := range msPub.PubKeys {
		assert.Contains(t, output, "\n  "+pk.String())
	}
	assert.NotContains(t, output, "... and")
}
