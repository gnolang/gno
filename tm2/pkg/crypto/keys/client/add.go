package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"regexp"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

var (
	errInvalidMnemonic       = errors.New("invalid bip39 mnemonic")
	errInvalidDerivationPath = errors.New("invalid derivation path")
)

var reDerivationPath = regexp.MustCompile(`^44'\/118'\/\d+'\/0\/\d+$`)

type AddCfg struct {
	RootCfg *BaseCfg

	Recover  bool
	NoBackup bool
	Account  uint64
	Index    uint64

	DerivationPath commands.StringArr
}

func NewAddCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &AddCfg{
		RootCfg: rootCfg,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "add [flags] <key-name>",
			ShortHelp:  "adds key to the keybase",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAdd(cfg, args, io)
		},
	)

	cmd.AddSubCommands(
		NewAddMultisigCmd(cfg, io),
		NewAddLedgerCmd(cfg, io),
		NewAddBech32Cmd(cfg, io),
	)

	return cmd
}

func (c *AddCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.Recover,
		"recover",
		false,
		"provide seed phrase to recover existing key instead of creating",
	)

	fs.BoolVar(
		&c.NoBackup,
		"nobackup",
		false,
		"don't print out seed phrase (if others are watching the terminal)",
	)

	fs.Uint64Var(
		&c.Account,
		"account",
		0,
		"account number for HD derivation",
	)

	fs.Uint64Var(
		&c.Index,
		"index",
		0,
		"address index number for HD derivation",
	)

	fs.Var(
		&c.DerivationPath,
		"derivation-path",
		"derivation path for deriving the address",
	)
}

func execAdd(cfg *AddCfg, args []string, io commands.IO) error {
	// Check if the key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	// Validate the derivation paths are correct
	for _, path := range cfg.DerivationPath {
		// Make sure the path is valid
		if _, err := hd.NewParamsFromPath(path); err != nil {
			return fmt.Errorf(
				"%w, %w",
				errInvalidDerivationPath,
				err,
			)
		}

		// Make sure the path conforms to the Gno derivation path
		if !reDerivationPath.MatchString(path) {
			return errInvalidDerivationPath
		}
	}

	name := args[0]

	// Read the keybase from the home directory
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("unable to read keybase, %w", err)
	}

	// Check if the key exists
	exists, err := kb.HasByName(name)
	if err != nil {
		return fmt.Errorf("unable to fetch key, %w", err)
	}

	// Get overwrite confirmation, if any
	if exists {
		overwrite, err := io.GetConfirmation(fmt.Sprintf("Override the existing name %s", name))
		if err != nil {
			return fmt.Errorf("unable to get confirmation, %w", err)
		}

		if !overwrite {
			return errOverwriteAborted
		}
	}

	// Ask for a password when generating a local key
	encryptPassword, err := io.GetCheckPassword(
		[2]string{
			"Enter a passphrase to encrypt your key to disk:",
			"Repeat the passphrase:",
		},
		cfg.RootCfg.InsecurePasswordStdin,
	)
	if err != nil {
		return fmt.Errorf("unable to parse provided password, %w", err)
	}

	// Get bip39 mnemonic
	mnemonic, err := GenerateMnemonic(mnemonicEntropySize)
	if err != nil {
		return fmt.Errorf("unable to generate mnemonic, %w", err)
	}

	if cfg.Recover {
		bip39Message := "Enter your bip39 mnemonic"
		mnemonic, err = io.GetString(bip39Message)
		if err != nil {
			return fmt.Errorf("unable to parse mnemonic, %w", err)
		}

		// Make sure it's valid
		if !bip39.IsMnemonicValid(mnemonic) {
			return errInvalidMnemonic
		}
	}

	// Save the account
	info, err := kb.CreateAccount(
		name,
		mnemonic,
		"",
		encryptPassword,
		uint32(cfg.Account),
		uint32(cfg.Index),
	)
	if err != nil {
		return fmt.Errorf("unable to save account to keybase, %w", err)
	}

	// Print the derived address info
	printDerive(mnemonic, cfg.DerivationPath, io)

	// Recover key from seed passphrase
	if cfg.Recover {
		printCreate(info, false, "", io)

		return nil
	}

	// Print the key create info
	printCreate(info, !cfg.NoBackup, mnemonic, io)

	return nil
}

func printCreate(info keys.Info, showMnemonic bool, mnemonic string, io commands.IO) {
	io.Println("")
	printNewInfo(info, io)

	// print mnemonic unless requested not to.
	if showMnemonic {
		io.Printfln(`
**IMPORTANT** write this mnemonic phrase in a safe place.
It is the only way to recover your account if you ever forget your password.
%v
`, mnemonic)
	}
}

func printNewInfo(info keys.Info, io commands.IO) {
	keyname := info.GetName()
	keytype := info.GetType()
	keypub := info.GetPubKey()
	keyaddr := info.GetAddress()
	keypath, _ := info.GetPath()

	io.Printfln("* %s (%s) - addr: %v pub: %v, path: %v",
		keyname, keytype, keyaddr, keypub, keypath)
}

// printDerive prints the derived accounts, if any
func printDerive(
	mnemonic string,
	paths []string,
	io commands.IO,
) {
	if len(paths) == 0 {
		// No accounts to print
		return
	}

	// Generate the accounts
	accounts := generateAccounts(
		mnemonic,
		paths,
	)

	io.Printf("[Derived Accounts]\n\n")

	// Print them out
	for index, path := range paths {
		io.Printfln(
			"%d. %s: %s",
			index,
			path,
			accounts[index].String(),
		)
	}
}

// generateAccounts the accounts using the provided mnemonics
func generateAccounts(mnemonic string, paths []string) []crypto.Address {
	addresses := make([]crypto.Address, len(paths))

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")

	for index, path := range paths {
		key := generateKeyFromSeed(seed, path)
		address := key.PubKey().Address()

		addresses[index] = address
	}

	return addresses
}

// generateKeyFromSeed generates a private key from
// the provided seed and path
func generateKeyFromSeed(seed []byte, path string) crypto.PrivKey {
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, path)

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}
