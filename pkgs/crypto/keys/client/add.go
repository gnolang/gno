package client

import (
	"fmt"
	"sort"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/bip39"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/multisig"
	"github.com/gnolang/gno/pkgs/errors"
)

type AddOptions struct {
	BaseOptions
	Multisig          []string `flag:"multisig" help:"Construct and store a multisig public key (implies --pubkey)"`
	MultisigThreshold int      `flag:"threshold" help:"K out of N required signatures. For use in conjunction with --multisig"`
	NoSort            bool     `flag:"nosort" help:"Keys passed to --multisig are taken in the order they're supplied"`
	PublicKey         string   `flag:"pubkey" help:"Parse a public key in bech32 format and save it to disk"`
	UseLedger         bool     `flag:"ledger" help:"Store a local reference to a private key on a Ledger device"`
	Recover           bool     `flag:"recover" help:"Provide seed phrase to recover existing key instead of creating"`
	NoBackup          bool     `flag:"nobackup" help:"Don't print out seed phrase (if others are watching the terminal)"`
	DryRun            bool     `flag:"dryrun" help:"Perform action, but don't add key to local keystore"`
	Account           uint32   `flag:"account" help:"Account number for HD derivation"`
	Index             uint32   `flag:"index" description:"Address index number for HD derivation"`
}

var DefaultAddOptions = AddOptions{
	BaseOptions:       DefaultBaseOptions,
	MultisigThreshold: 1,
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
func addApp(cmd *command.Command, args []string, iopts interface{}) error {
	var kb keys.Keybase
	var err error
	var encryptPassword string
	var opts AddOptions = iopts.(AddOptions)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: add <keyname>")
		return errors.New("invalid args")
	}

	name := args[0]
	showMnemonic := !opts.NoBackup

	if opts.DryRun {
		// we throw this away, so don't enforce args,
		// we want to get a new random seed phrase quickly
		kb = keys.NewInMemory()
		encryptPassword = DryRunKeyPass
	} else {
		kb, err = keys.NewKeyBaseFromDir(opts.Home)
		if err != nil {
			return err
		}

		_, err = kb.GetByName(name)
		if err == nil {
			// account exists, ask for user confirmation
			response, err2 := cmd.GetConfirmation(fmt.Sprintf("Override the existing name %s", name))
			if err2 != nil {
				return err2
			}
			if !response {
				return errors.New("aborted")
			}
		}

		multisigKeys := opts.Multisig
		if len(multisigKeys) != 0 {
			var pks []crypto.PubKey

			multisigThreshold := opts.MultisigThreshold
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
			if !opts.NoSort {
				sort.Slice(pks, func(i, j int) bool {
					return pks[i].Address().Compare(pks[j].Address()) < 0
				})
			}

			pk := multisig.NewPubKeyMultisigThreshold(multisigThreshold, pks)
			if _, err := kb.CreateMulti(name, pk); err != nil {
				return err
			}

			cmd.Printfln("Key %q saved to disk.\n", name)
			return nil
		}

		// ask for a password when generating a local key
		if opts.PublicKey == "" && !opts.UseLedger {
			encryptPassword, err = cmd.GetCheckPassword(
				"Enter a passphrase to encrypt your key to disk:",
				"Repeat the passphrase:")
			if err != nil {
				return err
			}
		}
	}

	if opts.PublicKey != "" {
		pk, err := crypto.PubKeyFromBech32(opts.PublicKey)
		if err != nil {
			return err
		}
		_, err = kb.CreateOffline(name, pk)
		if err != nil {
			return err
		}
		return nil
	}

	account := uint32(opts.Account)
	index := uint32(opts.Index)

	// If we're using ledger, only thing we need is the path and the bech32 prefix.
	if opts.UseLedger {
		bech32PrefixAddr := crypto.Bech32AddrPrefix
		info, err := kb.CreateLedger(name, keys.Secp256k1, bech32PrefixAddr, account, index)
		if err != nil {
			return err
		}

		return printCreate(cmd, info, false, "")
	}

	// Get bip39 mnemonic
	var mnemonic string
	const bip39Passphrase string = "" // XXX research.

	if opts.Recover {
		bip39Message := "Enter your bip39 mnemonic"
		mnemonic, err = cmd.GetString(bip39Message)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	}

	if len(mnemonic) == 0 {
		mnemonic, err = generateMnemonic(mnemonicEntropySize)
		if err != nil {
			return err
		}
	}

	info, err := kb.CreateAccount(name, mnemonic, bip39Passphrase, encryptPassword, account, index)
	if err != nil {
		return err
	}

	// Recover key from seed passphrase
	if opts.Recover {
		// Hide mnemonic from output
		showMnemonic = false
		mnemonic = ""
	}

	return printCreate(cmd, info, showMnemonic, mnemonic)
}

func printCreate(cmd *command.Command, info keys.Info, showMnemonic bool, mnemonic string) error {
	cmd.Println("")
	printNewInfo(cmd, info)

	// print mnemonic unless requested not to.
	if showMnemonic {
		cmd.Printfln(`
**IMPORTANT** write this mnemonic phrase in a safe place.
It is the only way to recover your account if you ever forget your password.

%v
`, mnemonic)
	}

	return nil
}

func printNewInfo(cmd *command.Command, info keys.Info) {
	keyname := info.GetName()
	keytype := info.GetType()
	keypub := info.GetPubKey()
	keyaddr := info.GetAddress()
	keypath, _ := info.GetPath()
	cmd.Printfln("* %s (%s) - addr: %v pub: %v, path: %v",
		keyname, keytype, keyaddr, keypub, keypath)
}
