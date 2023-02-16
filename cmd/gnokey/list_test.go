package main

import (
	"testing"

	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/stretchr/testify/assert"
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
			cfg := &baseCfg{
				BaseOptions: client.BaseOptions{
					Home: tt.kbDir,
				},
			}

			args := tt.args
			if err := execList(cfg, args); (err != nil) != tt.wantErr {
				t.Errorf("listApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
