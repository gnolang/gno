package main

import (
	"io/ioutil"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/std"
)

// replace the message sender/creater/signer to back key address if the message's sender address matches
// the backup key's primary address.

func signBkApp(cmd *command.Command, args []string, iopts interface{}) error {
	var kbPrimary keys.Keybase
	var kbBackup keys.Keybase
	var err error
	var opts client.SignOptions = iopts.(client.SignOptions)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: sign <keyname>")
		return errors.New("invalid args")
	}
	if opts.AccountNumber == nil {
		return errors.New("invalid account number")
	}
	if opts.Sequence == nil {
		return errors.New("invalid sequence")
	}

	name := args[0]
	txpath := opts.TxPath
	kbPrimary, err = keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}

	kbBackup, err = keys.NewBkKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}

	// read tx to sign
	var tx std.Tx
	var txjson []byte
	if txpath == "-" { // from stdin.
		txjsonstr, err := cmd.GetString("Enter tx to sign, terminated by a newline.")
		if err != nil {
			return err
		}
		txjson = []byte(txjsonstr)
	} else { // from file
		txjson, err = ioutil.ReadFile(txpath)
		if err != nil {
			return err
		}
	}
	err = amino.UnmarshalJSON(txjson, &tx)
	if err != nil {
		return err
	}

	// retrieve password
	pass, err := "", error(nil)
	if opts.Quiet {
		pass, err = cmd.GetPassword("")
	} else {
		pass, err = cmd.GetPassword("Enter password.")
	}
	if err != nil {
		return err
	}

	s := keys.SignerInfo{
		ChainId:       opts.ChainID,
		AccountNumber: *opts.AccountNumber, // back info's mutlsig account number
		Sequence:      *opts.Sequence,      // back info's multisig sequence number

	}

	// sign tx

	signedTx, err := keys.SignTx(kbPrimary, kbBackup, name, pass, tx, s)

	if err != nil {

		return err
	}
	// print tx
	txjson2, err := amino.MarshalJSON(signedTx)
	if err != nil {
		return err
	}
	cmd.Printfln(string(txjson2))

	return nil

}
