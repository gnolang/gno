package client

import (
	"io/ioutil"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/std"
)

type SignOptions struct {
	BaseOptions           // home,...
	TxPath        string  `flag:"txpath" help:"path to file of tx to sign"`
	ChainID       string  `flag:"chainid" help:"chainid to sign for"`
	AccountNumber *uint64 `flag:"number" help:"account number to sign with (required)"`
	Sequence      *uint64 `flag:"sequence" help:"sequence to sign with (required)"`
}

var DefaultSignOptions = SignOptions{
	BaseOptions: DefaultBaseOptions,
	TxPath:      "-", // read from stdin.
}

func signApp(cmd *command.Command, args []string, iopts interface{}) error {
	var kb keys.Keybase
	var err error
	var opts SignOptions = iopts.(SignOptions)

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
	kb, err = keys.NewKeyBaseFromDir(opts.Home)
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

	// fill tx signatures.
	signers := tx.GetSigners()
	if tx.Signatures == nil {
		for _, _ = range signers {
			tx.Signatures = append(tx.Signatures, std.Signature{
				PubKey:    nil, // zero signature
				Signature: nil, // zero signature
			})
		}
	}

	// validate document to sign.
	err = tx.ValidateBasic()
	if err != nil {
		return err
	}

	// derive sign doc bytes.
	chainID := opts.ChainID
	accountNumber := *opts.AccountNumber
	sequence := *opts.Sequence
	signbz := tx.GetSignBytes(chainID, accountNumber, sequence)

	pass, err := "", error(nil)
	if opts.Quiet {
		pass, err = cmd.GetPassword("")
	} else {
		pass, err = cmd.GetPassword("Enter password.")
	}
	if err != nil {
		return err
	}
	sig, pub, err := kb.Sign(name, pass, signbz)
	if err != nil {
		return err
	}
	addr := pub.Address()
	found := false
	for i, _ := range tx.Signatures {
		// override signature for matching slot.
		if signers[i] == addr {
			found = true
			tx.Signatures[i] = std.Signature{
				PubKey:    pub,
				Signature: sig,
			}
		}
	}
	if !found {
		return errors.New("addr %v (%s) not in signer set",
			addr, name)
	}
	txjson2, err := amino.MarshalJSON(tx)
	if err != nil {
		return err
	}
	cmd.Printfln(string(txjson2))
	return nil
}
