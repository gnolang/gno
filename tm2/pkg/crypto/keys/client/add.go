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
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

var (
	errOverwriteAborted      = errors.New("overwrite aborted")
	errInvalidMnemonic       = errors.New("invalid bip39 mnemonic")
	errInvalidDerivationPath = errors.New("invalid derivation path")
	errPassphraseMismatch    = errors.New("passphrases don't match")
)

var reDerivationPath = regexp.MustCompile(`^44'\/118'\/\d+'\/0\/\d+$`)

type AddCfg struct {
	RootCfg *BaseCfg

	Recover  bool
	NoBackup bool
	Account  uint64
	Index    uint64
	Entropy  bool
	Masked   bool

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

	fs.BoolVar(
		&c.Entropy,
		"entropy",
		false,
		"supply custom entropy for key generation instead of using computer's PRNG",
	)

	fs.BoolVar(
		&c.Masked,
		"masked",
		false,
		"mask input characters (use with --entropy or --recover)",
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

	// Check if the key name already exists
	confirmedKey, err := checkNameCollision(kb, name, io)
	if err != nil {
		return err
	}

	// Ask for a password when generating a local key
	pw, err := promptPassphrase(io, cfg.RootCfg.InsecurePasswordStdin)
	if err != nil {
		return err
	}

	var mnemonic string

	switch {
	case cfg.Recover:
		bip39Message := "Enter your bip39 mnemonic"
		if cfg.Masked {
			mnemonic, err = io.GetPassword(bip39Message, false)
		} else {
			mnemonic, err = io.GetString(bip39Message)
		}
		if err != nil {
			return fmt.Errorf("unable to parse mnemonic, %w", err)
		}

		// Make sure it's valid
		if !bip39.IsMnemonicValid(mnemonic) {
			return errInvalidMnemonic
		}
	case cfg.Entropy:
		// Generate mnemonic using custom entropy
		mnemonic, err = GenerateMnemonicWithCustomEntropy(io, cfg.Masked)
		if err != nil {
			return fmt.Errorf("unable to generate mnemonic with custom entropy, %w", err)
		}
	default:
		// Generate mnemonic using computer PRNG
		mnemonic, err = GenerateMnemonic(mnemonicEntropySize)
		if err != nil {
			return fmt.Errorf("unable to generate mnemonic, %w", err)
		}
	}

	// Derive the address early to check for address collision
	seed := bip39.NewSeed(mnemonic, "")
	hdPath := hd.NewFundraiserParams(uint32(cfg.Account), crypto.CoinType, uint32(cfg.Index))
	key := generateKeyFromSeed(seed, hdPath.String())
	newAddress := key.PubKey().Address()

	if err := checkAddressCollision(kb, newAddress, confirmedKey, io); err != nil {
		return err
	}

	// Save the account
	info, err := kb.CreateAccount(
		name,
		mnemonic,
		"",
		pw,
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

// promptPassphrase prompts for a password, with confirmation.
func promptPassphrase(io commands.IO, insecurePasswordStdin bool) (string, error) {
	pw, err := io.GetPassword("Enter a passphrase to encrypt your private key on disk: ", insecurePasswordStdin)
	if err != nil {
		return "", fmt.Errorf("unable to get provided passphrase, %w", err)
	}

	// If empty, just print the warning
	if pw == "" {
		io.Println("WARNING: a key with no passphrase will be stored UNENCRYPTED.\n" +
			"This is unsafe for any key used on-chain.")

		return "", nil
	}

	pw2, err := io.GetPassword("Repeat the passphrase: ", insecurePasswordStdin)
	if err != nil {
		return "", fmt.Errorf("unable to get provided passphrase, %w", err)
	}

	if pw != pw2 {
		return "", errPassphraseMismatch
	}

	return pw, nil
}

func printCreate(info keys.Info, showMnemonic bool, mnemonic string, io commands.IO) {
	io.Println("")
	printNewInfo(info, io)

	// print mnemonic unless requested not to.
	if showMnemonic {
		io.Printfln(`
**IMPORTANT** write this mnemonic phrase in a safe place.
It is the only way to recover your account if you ever forget your passphrase.
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

// checkNameCollision checks if a key with the given name already exists in the keybase.
// If it exists, it prints the existing key details and prompts for overwrite confirmation.
// Returns the existing key info if the user confirmed the overwrite, nil if no collision.
// Returns errOverwriteAborted if the user declines.
func checkNameCollision(kb keys.Keybase, name string, io commands.IO) (keys.Info, error) {
	exists, err := kb.HasByName(name)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch key, %w", err)
	}

	if !exists {
		return nil, nil
	}

	existingKey, err := kb.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch key, %w", err)
	}

	io.Println("Key already exists:")
	printNewInfo(existingKey, io)

	overwrite, err := io.GetConfirmation(fmt.Sprintf("Override the existing name %s", name))
	if err != nil {
		return nil, fmt.Errorf("unable to get confirmation, %w", err)
	}

	if !overwrite {
		return nil, errOverwriteAborted
	}

	return existingKey, nil
}

// checkAddressCollision checks if a key with the given address already exists in the keybase.
// If confirmedOverwrite is not nil, the check is skipped when the found key has the same name
// (meaning the user already confirmed the overwrite via name collision).
func checkAddressCollision(kb keys.Keybase, address crypto.Address, confirmedOverwrite keys.Info, io commands.IO) error {
	existingKey, err := kb.GetByAddress(address)
	if err != nil {
		if keyerror.IsErrKeyNotFound(err) {
			return nil
		}

		return fmt.Errorf("unable to fetch key by address, %w", err)
	}

	// Skip if this is the same key already confirmed via name collision
	if confirmedOverwrite != nil && existingKey.GetName() == confirmedOverwrite.GetName() {
		return nil
	}

	io.Println("An existing key already uses this address:")
	printNewInfo(existingKey, io)

	overwrite, err := io.GetConfirmation(fmt.Sprintf("Override the existing key %s", existingKey.GetName()))
	if err != nil {
		return fmt.Errorf("unable to get confirmation, %w", err)
	}

	if !overwrite {
		return errOverwriteAborted
	}

	return nil
}
