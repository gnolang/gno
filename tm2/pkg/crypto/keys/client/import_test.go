package client

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
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
	cfg := &importCfg{
		rootCfg: &baseCfg{
			BaseOptions: BaseOptions{
				Home:                  importOpts.kbHome,
				InsecurePasswordStdin: true,
			},
		},
		keyName:   importOpts.keyName,
		armorPath: importOpts.armorPath,
		unsafe:    importOpts.unsafe,
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
		name     string
		baseOpts testCmdKeyOptsBase
		input    io.Reader
	}{
		{
			"encrypted private key",
			testCmdKeyOptsBase{
				unsafe: false, // explicit
			},
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
			testCmdKeyOptsBase{
				unsafe: true,
			},
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n",
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
			info, err := addRandomKeyToKeybase(kb, keyName, password)
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
						unsafe:  testCase.baseOpts.unsafe,
					},
					outputPath: outputFile.Name(),
				},
				strings.NewReader(
					fmt.Sprintf(
						"%s\n%s\n%s\n",
						password,
						password,
						password,
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
						unsafe:  testCase.baseOpts.unsafe,
					},
					armorPath: outputFile.Name(),
				},
				testCase.input,
			); err != nil {
				t.Fatalf("unable to import private key armor, %v", err)
			}

			// Make sure the key-base has the new key imported
			info, err = kb.GetByName(importKeyName)

			assert.NotNil(t, info)
			assert.NoError(t, err)
		})
	}
}
