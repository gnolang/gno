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
			// Set current home
			rootCfg := &BaseCfg{
				BaseOptions: BaseOptions{
					Home: tt.kbDir,
				},
			}
			cfg := &ListCfg{
				RootCfg: rootCfg,
			}

			args := tt.args
			if err := execList(cfg, args, commands.NewTestIO()); (err != nil) != tt.wantErr {
				t.Errorf("listApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_execList_MultisigDefaultDisplay(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	privKeys := []secp256k1.PrivKeySecp256k1{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	require.NoError(t, kb.ImportPrivKey("k0", privKeys[0], ""))
	require.NoError(t, kb.ImportPrivKey("k1", privKeys[1], ""))
	require.NoError(t, kb.ImportPrivKey("k2", privKeys[2], ""))

	msPub := multisig.NewPubKeyMultisigThreshold(
		2,
		[]crypto.PubKey{
			privKeys[0].PubKey(),
			privKeys[1].PubKey(),
			privKeys[2].PubKey(),
		},
	)

	_, err = kb.CreateMulti("multisig-012", msPub)
	require.NoError(t, err)

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&out))

	cfg := &ListCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home: kbHome,
			},
		},
	}

	require.NoError(t, execList(cfg, nil, io))

	output := out.String()
	assert.Contains(t, output, "pub: "+crypto.PubKeyToBech32(msPub))
	assert.NotContains(t, output, msPub.String())
}

func Test_execList_MultisigMembersDisplay(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	privKeys := []secp256k1.PrivKeySecp256k1{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	require.NoError(t, kb.ImportPrivKey("k0", privKeys[0], ""))
	require.NoError(t, kb.ImportPrivKey("k1", privKeys[1], ""))
	require.NoError(t, kb.ImportPrivKey("k2", privKeys[2], ""))

	msPub := multisig.NewPubKeyMultisigThreshold(
		2,
		[]crypto.PubKey{
			privKeys[0].PubKey(),
			privKeys[1].PubKey(),
			privKeys[2].PubKey(),
		},
	)

	_, err = kb.CreateMulti("multisig-012", msPub)
	require.NoError(t, err)

	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&out))

	cfg := &ListCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home: kbHome,
			},
		},
		ShowMultisigMembers: true,
	}

	require.NoError(t, execList(cfg, nil, io))

	output := out.String()
	for _, pk := range msPub.(multisig.PubKeyMultisigThreshold).PubKeys {
		assert.Contains(t, output, "\n  "+pk.String()+"\n")
	}
	assert.Contains(t, output, "pub:")
	assert.NotContains(t, output, crypto.PubKeyToBech32(msPub))
}
