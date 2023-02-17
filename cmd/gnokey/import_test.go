package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/cmd/common"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/stretchr/testify/assert"
)

type testImportKeyOpts struct {
	testCmdKeyOptsBase

	armorPath string
}

// importKey runs the import private key command (from armor)
func importKey(importOpts testImportKeyOpts) error {
	var (
		cfg = &importCfg{
			rootCfg: &baseCfg{
				BaseOptions: common.BaseOptions{
					Home:                  importOpts.kbHome,
					InsecurePasswordStdin: true,
				},
			},
			keyName:   importOpts.keyName,
			armorPath: importOpts.armorPath,
		}
	)

	input := bufio.NewReader(
		strings.NewReader(
			fmt.Sprintf(
				"%s\n%s\n%s\n",
				importOpts.decryptPassword,
				importOpts.encryptPassword,
				importOpts.encryptPassword,
			),
		),
	)

	return execImport(cfg, input)
}

// TestImport_ImportKey makes sure the key can be imported correctly
func TestImport_ImportKey(t *testing.T) {
	t.Parallel()

	var (
		keyName       = "key name"
		importKeyName = "import key name"
		password      = "password"
	)

	// Generate a temporary key-base directory
	kb, kbHome := newTestKeybase(t)

	// Add an initial key to the key base
	info, err := addRandomKeyToKeybase(kb, keyName, password)
	if err != nil {
		t.Fatalf(
			"unable to create a key base account, %v",
			err,
		)
	}

	outputFile, outputCleanupFn := testutils.NewTestFile(t)
	defer outputCleanupFn()

	baseOpts := testCmdKeyOptsBase{
		kbHome:          kbHome,
		keyName:         info.GetName(),
		decryptPassword: password,
		encryptPassword: password,
	}

	// Make sure the export command executes correctly
	if err := exportKey(testExportKeyOpts{
		testCmdKeyOptsBase: baseOpts,
		outputPath:         outputFile.Name(),
	}); err != nil {
		t.Fatalf("unable to export private key, %v", err)
	}

	// Change the import key name so the existing one (in the key-base)
	// doesn't get overwritten
	baseOpts.keyName = importKeyName

	// Make sure the encrypted armor can be imported correctly
	if err := importKey(testImportKeyOpts{
		testCmdKeyOptsBase: baseOpts,
		armorPath:          outputFile.Name(),
	}); err != nil {
		t.Fatalf("unable to import private key armor, %v", err)
	}

	// Make sure the key-base has the new key imported
	info, err = kb.GetByName(importKeyName)

	assert.NotNil(t, info)
	assert.NoError(t, err)
}
