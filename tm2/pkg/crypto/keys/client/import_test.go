package client

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testImportKeyOpts struct {
	testCmdKeyOptsBase

	armorPath string
}

// importKey runs the import private key command (from armor)
func importKey(
	importOpts testImportKeyOpts,
	input io.Reader,
) error {
	cfg := &ImportCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  importOpts.kbHome,
				InsecurePasswordStdin: true,
			},
		},
		KeyName:   importOpts.keyName,
		ArmorPath: importOpts.armorPath,
	}

	cmdIO := commands.NewTestIO()
	cmdIO.SetIn(input)

	return execImport(cfg, cmdIO)
}

// TestImport_ImportKey makes sure the key can be imported correctly
func TestImport_ImportKey(t *testing.T) {
	t.Parallel()

	const (
		keyName       = "key name"
		importKeyName = "import key name"
		password      = "password"
	)

	testTable := []struct {
		name        string
		baseOpts    testCmdKeyOptsBase
		encryptPass string
		input       io.Reader
	}{
		{
			"encrypted private key",
			testCmdKeyOptsBase{},
			password,
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n%s\n",
					password, // decrypt
					password, // key-base encrypt
					password, // key-base encrypt confirm
				),
			),
		},
		{
			"unencrypted private key",
			testCmdKeyOptsBase{},
			"",
			strings.NewReader(
				fmt.Sprintf(
					"\n%s\n%s\n",
					password, // key-base encrypt
					password, // key-base encrypt confirm
				),
			),
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Generate a temporary key-base directory
			kb, kbHome := newTestKeybase(t)

			// Add an initial key to the key base
			_, err := addRandomKeyToKeybase(kb, keyName, testCase.encryptPass)
			if err != nil {
				t.Fatalf(
					"unable to create a key base account, %v",
					err,
				)
			}

			outputFile, outputCleanupFn := testutils.NewTestFile(t)
			defer outputCleanupFn()

			// Make sure the export command executes correctly
			if err := exportKey(
				testExportKeyOpts{
					testCmdKeyOptsBase: testCmdKeyOptsBase{
						kbHome:  kbHome,
						keyName: keyName,
					},
					outputPath: outputFile.Name(),
				},
				strings.NewReader(
					fmt.Sprintf(
						"%s\n%s\n%s\n",
						testCase.encryptPass,
						testCase.encryptPass,
						testCase.encryptPass,
					),
				),
			); err != nil {
				t.Fatalf("unable to export private key, %v", err)
			}

			// Make sure the encrypted armor can be imported correctly
			if err := importKey(
				testImportKeyOpts{
					testCmdKeyOptsBase: testCmdKeyOptsBase{
						kbHome: kbHome,
						// Change the import key name so the existing one (in the key-base)
						// doesn't get overwritten
						keyName: importKeyName,
					},
					armorPath: outputFile.Name(),
				},
				testCase.input,
			); err != nil {
				t.Fatalf("unable to import private key armor, %v", err)
			}

			// Make sure the key-base has the new key imported
			info, err := kb.GetByName(importKeyName)

			assert.NotNil(t, info)
			assert.NoError(t, err)
		})
	}
}

func TestImport_ImportKeyWithEmptyName(t *testing.T) {
	t.Parallel()

	// Generate a temporary key-base directory
	_, kbHome := newTestKeybase(t)
	err := importKey(
		testImportKeyOpts{
			testCmdKeyOptsBase: testCmdKeyOptsBase{
				kbHome:  kbHome,
				keyName: "",
			},
		},
		nil,
	)
	assert.Error(t, err)
	assert.EqualError(t, err, "name shouldn't be empty")
}

func TestImport_ImportKeyInvalidArmor(t *testing.T) {
	t.Parallel()

	_, kbHome := newTestKeybase(t)

	armorFile, err := os.CreateTemp("", "armor.key")
	require.NoError(t, err)

	defer os.Remove(armorFile.Name())

	// Write invalid armor
	_, err = armorFile.Write([]byte("totally valid tendermint armor"))
	require.NoError(t, err)

	err = importKey(
		testImportKeyOpts{
			testCmdKeyOptsBase: testCmdKeyOptsBase{
				kbHome:  kbHome,
				keyName: "key-name",
			},
			armorPath: armorFile.Name(),
		},
		strings.NewReader(
			fmt.Sprintf(
				"\n%s\n%s\n",
				"",
				"",
			),
		),
	)

	assert.ErrorContains(t, err, "unable to decrypt private key armor,")
}

func TestImport_ImportKeyInvalidPKArmor(t *testing.T) {
	t.Parallel()

	_, kbHome := newTestKeybase(t)

	armorFile, err := os.CreateTemp("", "armor.key")
	require.NoError(t, err)

	defer os.Remove(armorFile.Name())

	// Write invalid armor
	_, err = armorFile.Write([]byte("totally valid tendermint armor"))
	require.NoError(t, err)

	err = importKey(
		testImportKeyOpts{
			testCmdKeyOptsBase: testCmdKeyOptsBase{
				kbHome:  kbHome,
				keyName: "key-name",
			},
			armorPath: armorFile.Name(),
		},
		strings.NewReader(
			fmt.Sprintf(
				"%s\n%s\n%s\n",
				"",
				"",
				"",
			),
		),
	)

	assert.ErrorContains(t, err, "unable to decrypt private key armor")
}
