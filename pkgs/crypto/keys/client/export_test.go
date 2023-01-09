package client

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/stretchr/testify/assert"
)

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

	var (
		keyName         = "key name"
		encryptPassword = "encrypt password"

		cmd = command.NewMockCommand()
	)

	// Generate a temporary key-base directory
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	defer kbCleanUp()

	kb, err := keys.NewKeyBaseFromDir(kbHome)
	if err != nil {
		t.Fatalf(
			"unable to create a key base in directory %s, %v",
			kbHome,
			err,
		)
	}

	// Generate a random mnemonic
	mnemonic, err := generateMnemonic(mnemonicEntropySize)
	if err != nil {
		t.Fatalf(
			"unable to generate a mnemonic phrase, %v",
			err,
		)
	}

	// Add an initial key to the key base
	info, err := kb.CreateAccount(
		keyName,
		mnemonic,
		"",
		encryptPassword,
		0,
		0,
	)
	if err != nil {
		t.Fatalf(
			"unable to create a key base account, %v",
			err,
		)
	}

	outputFile, outputCleanupFn := testutils.NewTestFile(t)
	defer outputCleanupFn()

	opts := ExportOptions{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
		NameOrBech32: info.GetName(),
		OutputPath:   outputFile.Name(),
	}

	// Prepend standard input
	cmd.SetIn(
		strings.NewReader(
			fmt.Sprintf(
				"%s\n%s\n%s\n",
				encryptPassword,
				encryptPassword,
				encryptPassword,
			),
		),
	)

	// Make sure the command executes correctly
	assert.NoError(
		t,
		exportApp(cmd, nil, opts),
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
}
