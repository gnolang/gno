package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"sort"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
)

type addCfg struct {
	rootCfg *baseCfg

	multisig          commands.StringArr
	multisigThreshold int
	noSort            bool
	publicKey         string
	useLedger         bool
	recover           bool
	noBackup          bool
	dryRun            bool
	account           uint64
	index             uint64
}

func newAddCmd(rootCfg *baseCfg, io *commands.IO) *commands.Command {
	cfg := &addCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "add [flags] <key-name>",
			ShortHelp:  "Adds key to the keybase",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAdd(cfg, args, io)
		},
	)
}

func (c *addCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.Var(
		&c.multisig,
		"multisig",
		"Construct and store a multisig public key (implies --pubkey)",
	)

	fs.IntVar(
		&c.multisigThreshold,
		"threshold",
		1,
		"K out of N required signatures. For use in conjunction with --multisig",
	)

	fs.BoolVar(
		&c.noSort,
		"nosort",
		false,
		"Keys passed to --multisig are taken in the order they're supplied",
	)

	fs.StringVar(
		&c.publicKey,
		"pubkey",
		"",
		"Parse a public key in bech32 format and save it to disk",
	)

	fs.BoolVar(
		&c.useLedger,
		"ledger",
		false,
		"Store a local reference to a private key on a Ledger device",
	)

	fs.BoolVar(
		&c.recover,
		"recover",
		false,
		"Provide seed phrase to recover existing key instead of creating",
	)

	fs.BoolVar(
		&c.noBackup,
		"nobackup",
		false,
		"Don't print out seed phrase (if others are watching the terminal)",
	)

	fs.BoolVar(
		&c.dryRun,
		"dryrun",
		false,
		"Perform action, but don't add key to local keystore",
	)

	fs.Uint64Var(
		&c.account,
		"account",
		0,
		"Account number for HD derivation",
	)

	fs.Uint64Var(
		&c.index,
		"index",
		0,
		"Address index number for HD derivation",
	)
}

// DryRunKeyPass contains the default key password for genesis transactions
const DryRunKeyPass = "12345678"

/*
input
  - bip39 mnemonic
  - bip39 passphrase
  - bip44 path
  - local encryption password

output
  - armor encrypted private key (saved to file)
*/
func execAdd(cfg *addCfg, args []string, io *commands.IO) error {
	var (
		kb              keys.Keybase
		err             error
		encryptPassword string
	)

	if len(args) != 1 {
		return flag.ErrHelp
	}

	name := args[0]
	showMnemonic := !cfg.noBackup

	if cfg.dryRun {
		// we throw this away, so don't enforce args,
		// we want to get a new random seed phrase quickly
		kb = keys.NewInMemory()
		encryptPassword = DryRunKeyPass
	} else {
		kb, err = keys.NewKeyBaseFromDir(cfg.rootCfg.Home)
		if err != nil {
			return err
		}

		if has, err := kb.HasByName(name); err == nil && has {
			// account exists, ask for user confirmation
			response, err2 := io.GetConfirmation(fmt.Sprintf("Override the existing name %s", name))
			if err2 != nil {
				return err2
			}
			if !response {
				return errors.New("aborted")
			}
		}

		multisigKeys := cfg.multisig
		if len(multisigKeys) != 0 {
			var pks []crypto.PubKey

			multisigThreshold := cfg.multisigThreshold
			if err := keys.ValidateMultisigThreshold(multisigThreshold, len(multisigKeys)); err != nil {
				return err
			}

			for _, keyname := range multisigKeys {
				k, err := kb.GetByName(keyname)
				if err != nil {
					return err
				}
				pks = append(pks, k.GetPubKey())
			}

			// Handle --nosort
			if !cfg.noSort {
				sort.Slice(pks, func(i, j int) bool {
					return pks[i].Address().Compare(pks[j].Address()) < 0
				})
			}

			pk := multisig.NewPubKeyMultisigThreshold(multisigThreshold, pks)
			if _, err := kb.CreateMulti(name, pk); err != nil {
				return err
			}

			io.Printfln("Key %q saved to disk.\n", name)
			return nil
		}

		// ask for a password when generating a local key
		if cfg.publicKey == "" && !cfg.useLedger {
			encryptPassword, err = io.GetCheckPassword(
				[2]string{
					"Enter a passphrase to encrypt your key to disk:",
					"Repeat the passphrase:",
				},
				cfg.rootCfg.InsecurePasswordStdin,
			)
			if err != nil {
				return err
			}
		}
	}

	if cfg.publicKey != "" {
		pk, err := crypto.PubKeyFromBech32(cfg.publicKey)
		if err != nil {
			return err
		}
		_, err = kb.CreateOffline(name, pk)
		if err != nil {
			return err
		}
		return nil
	}

	account := cfg.account
	index := cfg.index

	// If we're using ledger, only thing we need is the path and the bech32 prefix.
	if cfg.useLedger {
		bech32PrefixAddr := crypto.Bech32AddrPrefix
		info, err := kb.CreateLedger(name, keys.Secp256k1, bech32PrefixAddr, uint32(account), uint32(index))
		if err != nil {
			return err
		}

		return printCreate(info, false, "", io)
	}

	// Get bip39 mnemonic
	var mnemonic string
	const bip39Passphrase string = "" // XXX research.

	if cfg.recover {
		bip39Message := "Enter your bip39 mnemonic"
		mnemonic, err = io.GetString(bip39Message)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	}

	if len(mnemonic) == 0 {
		mnemonic, err = GenerateMnemonic(mnemonicEntropySize)
		if err != nil {
			return err
		}
	}

	info, err := kb.CreateAccount(name, mnemonic, bip39Passphrase, encryptPassword, uint32(account), uint32(index))
	if err != nil {
		return err
	}

	// Recover key from seed passphrase
	if cfg.recover {
		// Hide mnemonic from output
		showMnemonic = false
		mnemonic = ""
	}

	return printCreate(info, showMnemonic, mnemonic, io)
}

func printCreate(info keys.Info, showMnemonic bool, mnemonic string, io *commands.IO) error {
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

	return nil
}

func printNewInfo(info keys.Info, io *commands.IO) {
	keyname := info.GetName()
	keytype := info.GetType()
	keypub := info.GetPubKey()
	keyaddr := info.GetAddress()
	keypath, _ := info.GetPath()

	io.Printfln("* %s (%s) - addr: %v pub: %v, path: %v",
		keyname, keytype, keyaddr, keypub, keypath)
}
