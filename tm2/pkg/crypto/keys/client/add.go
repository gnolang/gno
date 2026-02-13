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
		"derivation path for deriving and persisting key in the keybase",
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

	getMnemonic := func() (string, error) {
		switch {
		case cfg.Recover:
			bip39Message := "Enter your bip39 mnemonic"
			var mnemonic string
			var err error
			if cfg.Masked {
				mnemonic, err = io.GetPassword(bip39Message, false)
			} else {
				mnemonic, err = io.GetString(bip39Message)
			}
			if err != nil {
				return "", fmt.Errorf("unable to parse mnemonic, %w", err)
			}

			// Make sure it's valid
			if !bip39.IsMnemonicValid(mnemonic) {
				return "", errInvalidMnemonic
			}

			return mnemonic, nil
		case cfg.Entropy:
			// Generate mnemonic using custom entropy
			mnemonic, err := GenerateMnemonicWithCustomEntropy(io, cfg.Masked)
			if err != nil {
				return "", fmt.Errorf("unable to generate mnemonic with custom entropy, %w", err)
			}

			return mnemonic, nil
		default:
			// Generate mnemonic using computer PRNG
			mnemonic, err := GenerateMnemonic(mnemonicEntropySize)
			if err != nil {
				return "", fmt.Errorf("unable to generate mnemonic, %w", err)
			}

			return mnemonic, nil
		}
	}

	var (
		infos    []keys.Info
		mnemonic string
	)

	type deriveEntry struct {
		name   string
		params *hd.BIP44Params
	}

	confirmOverwrite := func(keyName string) error {
		exists, err := kb.HasByName(keyName)
		if err != nil {
			return fmt.Errorf("unable to fetch key, %w", err)
		}

		if exists {
			overwrite, err := io.GetConfirmation(fmt.Sprintf("Override the existing name %s", keyName))
			if err != nil {
				return fmt.Errorf("unable to get confirmation, %w", err)
			}

			if !overwrite {
				return errOverwriteAborted
			}
		}

		return nil
	}

	if len(cfg.DerivationPath) == 0 {
		// Normal derivation uses account/index flags.
		if err := confirmOverwrite(name); err != nil {
			return err
		}

		// Ask for a password when generating a local key
		pw, err := promptPassphrase(io, cfg.RootCfg.InsecurePasswordStdin)
		if err != nil {
			return err
		}

		mnemonic, err = getMnemonic()
		if err != nil {
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

		infos = []keys.Info{info}
	} else {
		// Derivation paths override account/index flags.
		if cfg.Account != 0 || cfg.Index != 0 {
			io.Println("WARNING: -account/-index are ignored when -derivation-path is provided.")
		}

		entries := make([]deriveEntry, 0, len(cfg.DerivationPath))

		for _, path := range cfg.DerivationPath {
			params, err := hd.NewParamsFromPath(path)
			if err != nil {
				return fmt.Errorf("unable to parse derivation path, %w", err)
			}

			derivedName := deriveKeyName(name, params, len(cfg.DerivationPath))
			if err := confirmOverwrite(derivedName); err != nil {
				return err
			}

			entries = append(entries, deriveEntry{
				name:   derivedName,
				params: params,
			})
		}

		mnemonic, err = getMnemonic()
		if err != nil {
			return err
		}

		infos = make([]keys.Info, 0, len(entries))
		passphrases := make([]string, len(entries))

		for i := range entries {
			// Ask for a password when generating a local key
			pw, err := promptPassphrase(io, cfg.RootCfg.InsecurePasswordStdin)
			if err != nil {
				return err
			}

			passphrases[i] = pw
		}

		for i, entry := range entries {
			info, err := kb.CreateAccountBip44(
				entry.name,
				mnemonic,
				"",
				passphrases[i],
				*entry.params,
			)
			if err != nil {
				return fmt.Errorf("unable to save account to keybase, %w", err)
			}

			infos = append(infos, info)
		}
	}

	// Print the derived address info
	printDerive(mnemonic, cfg.DerivationPath, io)

	// Recover key from seed passphrase
	if cfg.Recover {
		for _, info := range infos {
			printCreate(info, false, "", io)
		}

		return nil
	}

	// Print the key create info (mnemonic only once)
	for i, info := range infos {
		printCreate(info, !cfg.NoBackup && i == 0, mnemonic, io)
	}

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

func deriveKeyName(base string, params *hd.BIP44Params, totalPaths int) string {
	if totalPaths == 1 {
		return base
	}

	return fmt.Sprintf("%s-a%di%d", base, params.Account, params.AddressIndex)
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
