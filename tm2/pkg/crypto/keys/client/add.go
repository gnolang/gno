package client

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

var errInvalidMnemonic = errors.New("invalid bip39 mnemonic")

type AddCfg struct {
	RootCfg *BaseCfg

	Recover        bool
	NoBackup       bool
	Account        uint64
	Index          uint64
	DeriveAccounts uint64
}

type AddBaseCfg struct {
	RootCfg *BaseCfg
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

	fs.Uint64Var(
		&c.DeriveAccounts,
		"derive-accounts",
		0,
		"the number of accounts to derive from the mnemonic",
	)
}

/*
input
  - bip39 mnemonic
  - bip39 passphrase
  - bip44 path
  - local encryption password

output
  - armor encrypted private key (saved to file)
*/
func execAdd(cfg *AddCfg, args []string, io commands.IO) error {
	// Check if the key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
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
	printDerive(mnemonic, cfg.Index, cfg.DeriveAccounts, io)

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
	accountIndex,
	numAccounts uint64,
	io commands.IO,
) {
	if numAccounts == 0 {
		// No accounts to print
		return
	}

	// Generate the accounts
	accounts := generateAccounts(
		mnemonic,
		accountIndex,
		numAccounts,
	)

	io.Printf("[Derived Accounts]\n\n")
	io.Printf("Account Index: %d\n\n", accountIndex)

	// Print them out
	for index, account := range accounts {
		io.Printfln("%d. %s", index, account.String())
	}
}

// generateAccounts the accounts using the provided mnemonics
func generateAccounts(mnemonic string, accountIndex, numAccounts uint64) []crypto.Address {
	addresses := make([]crypto.Address, numAccounts)

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")

	for i := uint64(0); i < numAccounts; i++ {
		key := generateKeyFromSeed(seed, uint32(accountIndex), uint32(i))
		address := key.PubKey().Address()

		addresses[i] = address
	}

	return addresses
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, account, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(account, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}
