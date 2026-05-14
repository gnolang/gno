package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"

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
	Force    bool

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

	fs.BoolVar(
		&c.Force,
		"force",
		false,
		"override any existing key without interactive prompts",
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

	// If not forcing, check for collisions with existing keys
	if !cfg.Force {
		// Derive the address to check for collision
		seed := bip39.NewSeed(mnemonic, "")
		hdPath := hd.NewFundraiserParams(uint32(cfg.Account), crypto.CoinType, uint32(cfg.Index))
		key := generateKeyFromSeed(seed, hdPath.String())
		newAddress := key.PubKey().Address()

		// Handle address / name collision if any
		handled, err := handleCollision(kb, name, newAddress, keys.TypeLocal, io)
		if err != nil {
			return err
		}
		// If a collision was found and handled, we can skip saving the new key
		if handled {
			return nil
		}
	}

	// Ask for passphrase only when proceeding with key creation
	pw, err := promptPassphrase(io, cfg.RootCfg.InsecurePasswordStdin)
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

// hasSigningCapability returns true for key types that can sign transactions.
func hasSigningCapability(t keys.KeyType) bool {
	return t == keys.TypeLocal || t == keys.TypeLedger
}

// useColor returns whether ANSI color codes should be emitted to the given output.
// Returns false when NO_COLOR is set or the output is not a terminal.
func useColor(cio commands.IO) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if f, ok := cio.Out().(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// boldForTerminal returns the input string wrapped in bold ANSI escape codes
// when the output supports color, otherwise returns the string unchanged.
func boldForTerminal(s string, cio commands.IO) string {
	if !useColor(cio) {
		return s
	}
	const (
		ansiBold  = "\033[1m"
		ansiReset = "\033[0m"
	)
	return ansiBold + s + ansiReset
}

// printCollisionDiff prints a diff-style comparison between an existing key and the new key,
// highlighting fields that differ using ANSI bold.
func printCollisionDiff(
	existingName string, existingAddress crypto.Address, existingType keys.KeyType,
	newName string, newAddress crypto.Address, newType keys.KeyType,
	io commands.IO,
) {
	printKeyDetail := func(name string, address crypto.Address, keyType keys.KeyType) {
		// Append signing capability info to the type string
		typeStr := keyType.String()
		if hasSigningCapability(keyType) {
			typeStr += " (signing)"
		} else {
			typeStr += " (public-key-only)"
		}

		// Represent zero addresses with a placeholder for better readability
		addrStr := address.String()
		if keyType == keys.TypeLedger && address.IsZero() {
			addrStr = "(unknown - stored on ledger)"
		} else if address.IsZero() {
			addrStr = "(none)"
		}

		// Highlight differences in bold
		if existingName != newName {
			name = boldForTerminal(name, io)
		}
		if existingAddress != newAddress {
			addrStr = boldForTerminal(addrStr, io)
		}
		if existingType != newType {
			typeStr = boldForTerminal(typeStr, io)
		}

		// Print the key details
		io.Printfln("  Name:     %s", name)
		io.Printfln("  Address:  %s", addrStr)
		io.Printfln("  Type:     %s", typeStr)
		io.Println("")
	}

	io.Println("Key collision detected:\n")

	io.Println("Existing key:")
	printKeyDetail(existingName, existingAddress, existingType)

	io.Println("New key:")
	printKeyDetail(newName, newAddress, newType)
}

// handleCollision checks for existing keys that collide with the new key by name or address,
// and prompts the user to resolve the collision if any is found. It returns a boolean indicating
// whether the collision was handled (e.g. by renaming) and an error if any occurred during handling.
func handleCollision(
	kb keys.Keybase,
	newName string, newAddress crypto.Address, newType keys.KeyType,
	io commands.IO,
) (bool, error) {
	// Look for existing key by name
	existingByName, err := kb.GetByName(newName)
	if err != nil && !keyerror.IsErrKeyNotFound(err) {
		return false, fmt.Errorf("unable to fetch key by name: %w", err)
	}

	// Look for existing key by address (if address is non-zero)
	var existingByAddr keys.Info
	if !newAddress.IsZero() {
		existingByAddr, err = kb.GetByAddress(newAddress)
		if err != nil && !keyerror.IsErrKeyNotFound(err) {
			return false, fmt.Errorf("unable to fetch key by address: %w", err)
		}
	}

	// Detect double-collision: name matches one key, address matches a different key.
	// This requires manual resolution since automatic handling could silently delete data.
	if existingByName != nil && existingByAddr != nil &&
		existingByName.GetName() != existingByAddr.GetName() {
		return false, fmt.Errorf(
			"double collision detected:\n"+
				"  - Name %q is already used by a key with address %s\n"+
				"  - Address %s is already used by key %q\n"+
				"Resolve manually by deleting or renaming one of the conflicting keys",
			newName, existingByName.GetAddress(),
			newAddress, existingByAddr.GetName(),
		)
	}

	// Use whichever collision was found (they point to the same key, or only one exists)
	existing := existingByName
	if existing == nil {
		existing = existingByAddr
	}

	// No collision found
	if existing == nil {
		return false, nil
	}

	// Collision found - resolve it
	var (
		existingName    = existing.GetName()
		existingAddress = existing.GetAddress()
		existingType    = existing.GetType()

		sameName    = existingName == newName
		sameAddress = existingAddress == newAddress
		sameType    = existingType == newType
	)

	// Print the diff
	printCollisionDiff(existingName, existingAddress, existingType, newName, newAddress, newType, io)

	// Case 1: Exactly the same key (same name, address, type) -> skip
	if sameName && sameAddress && sameType {
		io.Println("Key is identical. Skipping.")
		return true, nil
	}

	// Case 2: Only the name differs -> prompt rename
	if !sameName && sameAddress && sameType {
		rename, err := io.GetConfirmation(fmt.Sprintf("Rename the existing key %q", existingName))
		if err != nil {
			return false, fmt.Errorf("unable to get confirmation, %w", err)
		}
		if !rename {
			return false, errOverwriteAborted
		}
		if err := kb.Rename(existingName, newName); err != nil {
			return false, err
		}
		return true, nil
	}

	// Case 3: Name differs, same address, and signing capability would be lost -> prompt rename vs override
	if !sameName && sameAddress && hasSigningCapability(existingType) && !hasSigningCapability(newType) {
		choice, err := io.GetString(fmt.Sprintf(`
  You are about to overwrite a key with signing capability (%q) with a key that does not have signing capability (%q).

  Options:
    (R)ename: rename existing key %q to %q, keeping signing capability.
    (o)verride: replace with the new key. %s
    (c)ancel: abort.

  Choose an action [R/o/c]: `,
			existingName, newName, existingName, newName, boldForTerminal("⚠ This will lose signing capability.", io),
		))
		if err != nil {
			return false, fmt.Errorf("unable to get choice, %w", err)
		}

		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "rename", "r", "":
			if err := kb.Rename(existingName, newName); err != nil {
				return false, err
			}
			return true, nil
		case "override", "o":
			return false, nil
		default:
			return false, errOverwriteAborted
		}
	}

	// Other cases -> prompt override
	if hasSigningCapability(existingType) && !hasSigningCapability(newType) {
		io.Printfln("\n  %s", boldForTerminal(
			"⚠ Warning: this will replace a key with signing capability with one that cannot sign transactions.", io))
	}
	overwrite, err := io.GetConfirmation(fmt.Sprintf("Override the existing key %q", existingName))
	if err != nil {
		return false, fmt.Errorf("unable to get confirmation, %w", err)
	}
	if !overwrite {
		return false, errOverwriteAborted
	}
	return false, nil
}
