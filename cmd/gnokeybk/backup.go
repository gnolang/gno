package main

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/bip39"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	"github.com/gnolang/gno/pkgs/errors"
)

// It finds the address to the key name and ask user to generate a new  priviate key with the same  nemonic
// and sign the relation between the  new  backup public key and  current pubkey.
// If the name is not found, it asks user to add new key, which automatically genereate back up key.
// TODO  Add customized entropy to generate Mnemonic.
const mnemonicEntropySize = 256

func backupKeyApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(client.BaseOptions)

	if len(args) != 1 {

		cmd.ErrPrintln("Usage: gnokeybk bkkey <keyname>")
		return errors.New("invalid args")
	}

	// read primary key's public info
	name := args[0]
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {

		return err
	}

	info, err := kb.Get(name)
	if err != nil {
		//TODO: call addApp to generate mnemonic and the primary key
		return fmt.Errorf("%s does not exist. please create a primary key first", name)
	}
	//TODO: add switch to support ledger info

	if info.GetType() != keys.TypeLocal {
		return errors.New("backup key only work for local private key")
	}

	addr := info.GetAddress()
	cmd.Printfln("This is your primary wallet address: %s\n", addr)
	cmd.Printfln("Please input corresponding mnemonic to generate back up key.")

	// import mnemonic and add bkkey in backup key store
	// TODO: take care  of multisig case and ledger case as in addApp()

	// you can have one single seed with multiple passphrases to create multiple different wallets.
	// Each wallet would be designated by a different passphrase. seed = "mnemonic"+phassphrase?
	const bip39Passphrase string = ""
	// TODO: should user enter bip39 passphrase? Maybe not. User has a lot burn already.
	// TODO: should we add  bip39 passphrase to backup key generation automatically?
	// Maybe not, backup key already creates an  layer of security and extra burdon to the user.
	// Plus, backward compatible  maintenaince will be a nightmare

	passphrase, err := cmd.GetPassword("Enter the passphrase to unlock the key store")

	var priv crypto.PrivKey
	priv, err = kb.ExportPrivateKeyObject(name, passphrase)

	if err != nil {

		return fmt.Errorf("Please check the pass phrase for %s, it can not unlock the keybase.", name)
	}

	kbBK, err := keys.NewBkKeyBaseFromDir(opts.Home)

	if err != nil {
		return err
	}

	//TODO: Do we allow people create multiple backup key?
	// It could be a nightmare for an end user track multiple backup key.

	i, err := kbBK.Get(name)
	if i != nil {

		return fmt.Errorf("backup key already generated for %s", name)
	}

	bip39Message := "Enter your backup bip39 mnemonic, which should be different from you primary mnemonic, or hit enter to generate a new one"
	mnemonic, err := cmd.GetString(bip39Message)

	if err != nil {

		return err
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

		cmd.Printfln(`
**IMPORTANT** write this mnemonic phrase in a safe place.
It is the only way to recover your back up account if you ever forget your password.
%v
`, mnemonic)
	}

	if !bip39.IsMnemonicValid(mnemonic) {

		return errors.New("invalid mnemonic")

	}
	// the bip39 passphrase is appendixed to the mnemonic to generate new account
	//TODO: take care multi derived accounts from the same mnemonic
	account := uint32(0)
	index := uint32(0)

	infobk, err := keys.BackupAccount(priv, kbBK, name, mnemonic, bip39Passphrase, passphrase, account, index)

	addrbk := infobk.GetAddress()
	// verify if mnemonic generate the same address
	/*
		if addr.Compare(addrbk) != 0 {
			mnemonicMsg := "The imput mnemonic is not correct.\n %s \n"
			addrMsg := "It does not match the address.\n %s \n"
			return fmt.Errorf(
				mnemonicMsg, mnemonic, addrMsg, mnemonic, addr.String())

		}
	*/

	cmd.Printfln("\nBackup key  is created for primary key address\n%s", addr)
	cmd.Printfln("\nBackup key's multisig address is \n%s", addrbk.String())

	return nil
}
