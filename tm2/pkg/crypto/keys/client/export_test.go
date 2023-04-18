package client

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// newTestKeybase generates a new test key-base
// Returns the temporary key-base, and its path
func newTestKeybase(t *testing.T) (keys.Keybase, string) {
	t.Helper()

	// Generate a temporary key-base directory
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)

	t.Cleanup(func() {
		kbCleanUp()
	})

	kb, err := keys.NewKeyBaseFromDir(kbHome)
	if err != nil {
		t.Fatalf(
			"unable to create a key base in directory %s, %v",
			kbHome,
			err,
		)
	}

	return kb, kbHome
}

// addRandomKeyToKeybase adds a random key to the key-base
func addRandomKeyToKeybase(
	kb keys.Keybase,
	keyName,
	encryptPassword string,
) (keys.Info, error) {
	// Generate a random mnemonic
	mnemonic, err := generateMnemonic(mnemonicEntropySize)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to generate a mnemonic phrase, %w",
			err,
		)
	}

	// Add the key to the key base
	return kb.CreateAccount(
		keyName,
		mnemonic,
		"",
		encryptPassword,
		0,
		0,
	)
}

type testCmdKeyOptsBase struct {
	kbHome          string
	keyName         string
	decryptPassword string
	encryptPassword string
	unsafe          bool
}

type testExportKeyOpts struct {
	testCmdKeyOptsBase

	outputPath string
}

// exportKey runs the private key export command
// using the provided options
func exportKey(
	exportOpts testExportKeyOpts,
	input io.Reader,
) error {
	cfg := &exportCfg{
		rootCfg: &baseCfg{
			BaseOptions: BaseOptions{
				Home:                  exportOpts.kbHome,
				InsecurePasswordStdin: true,
			},
		},
		nameOrBech32: exportOpts.keyName,
		outputPath:   exportOpts.outputPath,
		unsafe:       exportOpts.unsafe,
	}

	cmdIO := commands.NewTestIO()
	cmdIO.SetIn(input)

	return execExport(cfg, cmdIO)
}

// TestExport_ExportKey makes sure the key can be exported correctly
func TestExport_ExportKey(t *testing.T) {
	t.Parallel()

	// numLines returns the number of new lines
	// in a string
	numLines := func(s string) int {
		n := strings.Count(s, "\n")
		if len(s) > 0 && !strings.HasSuffix(s, "\n") {
			n++
		}

		return n
	}

	const (
		keyName  = "key name"
		password = "password"
	)

	testTable := []struct {
		name     string
		baseOpts testCmdKeyOptsBase
		input    io.Reader
	}{
		{
			"encrypted key export",
			testCmdKeyOptsBase{},
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n%s\n",
					password, // decrypt
					password, // encrypt
					password, // encrypt confirm
				),
			),
		},
		{
			"unencrypted key export",
			testCmdKeyOptsBase{
				unsafe: true,
			},
			strings.NewReader(
				fmt.Sprintf(
					"%s\n",
					password, // decrypt
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
			t.Cleanup(func() {
				outputCleanupFn()
			})

			// Make sure the command executes correctly
			assert.NoError(
				t,
				exportKey(
					testExportKeyOpts{
						testCmdKeyOptsBase: testCmdKeyOptsBase{
							kbHome:  kbHome,
							keyName: info.GetName(),
							unsafe:  testCase.baseOpts.unsafe,
						},
						outputPath: outputFile.Name(),
					},
					testCase.input,
				),
			)

			// Make sure the encrypted armor has been written to disk
			buff, err := os.ReadFile(outputFile.Name())
			if err != nil {
				t.Fatalf(
					"unable to read temporary file from disk, %v",
					err,
				)
			}

			assert.Greater(t, numLines(string(buff)), 1)
		})
	}
}
