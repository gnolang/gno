package client

import (
	"errors"
	"fmt"
	"sort"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/bip39"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/multisig"
)

type BaseOptions struct {
	Home string
}

type AddOptions struct {
	BaseOptions                // home,...
	Multisig          []string // construct and store a multisig public key (implies --pubkey)
	MultisigThreshold int      // K out of N required signatures. For use in conjunction with --multisig
	NoSort            bool     // Keys passed to --multisig are taken in the order they're supplied
	PublicKey         string   // Parse a public key in bech32 format and save it to disk
	Interactive       bool     // Interactively prompt user for BIP39 passphrase and mnemonic
	UseLedger         bool     // Store a local reference to a private key on a Ledger device
	Recover           bool     // Provide seed phrase to recover existing key instead of creating
	NoBackup          bool     // Don't print out seed phrase (if others are watching the terminal)
	DryRun            bool     // Perform action, but don't add key to local keystore
	Account           uint32   // Account number for HD derivation
	Index             uint32   // Address index number for HD derivation
}

var DefaultAddOptions = AddOptions{
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
func runAddCmd(cmd *command.Command) error {
	var kb keys.Keybase
	var err error
	var encryptPassword string
	var opts AddOptions = cmd.Options.(AddOptions)
	var args = cmd.Args

	name := args[0]

	interactive := opts.Interactive
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

		_, err = kb.Get(name)
		if err == nil {
			// account exists, ask for user confirmation
			response, err2 := cmd.GetConfirmation(fmt.Sprintf("override the existing name %s", name))
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
				k, err := kb.Get(keyname)
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
	var bip39Passphrase string

	if interactive || opts.Recover {
		bip39Message := "Enter your bip39 mnemonic"
		if !opts.Recover {
			bip39Message = "Enter your bip39 mnemonic, or hit enter to generate one."
		}

		mnemonic, err = cmd.GetString(bip39Message)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	}

	if len(mnemonic) == 0 {
		// read entropy seed straight from crypto.Rand and convert to mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}

		mnemonic, err = bip39.NewMnemonic(entropySeed[:])
		if err != nil {
			return err
		}
	}

	// override bip39 passphrase
	if interactive {
		bip39Passphrase, err = cmd.GetString(
			"Enter your bip39 passphrase. This is combined with the mnemonic to derive the seed. " +
				"Most users should just hit enter to use the default, \"\"")
		if err != nil {
			return err
		}

		// if they use one, make them re-enter it
		if len(bip39Passphrase) != 0 {
			p2, err := cmd.GetString("Repeat the passphrase:")
			if err != nil {
				return err
			}

			if bip39Passphrase != p2 {
				return errors.New("passphrases don't match")
			}
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
	cmd.Printfln("\n%#v", info)

	// print mnemonic unless requested not to.
	if showMnemonic {
		cmd.Printfln(
			`**Important** write this mnemonic phrase in a safe place.
				It is the only way to recover your account if you ever forget your password.
				
				%v
				`, mnemonic)
	}

	return nil
}
