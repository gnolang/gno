package client

import (
	"fmt"
	"os"

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
	ShowSignBytes bool    `flag:"show-signbytes" help:"show sign bytes and quit"`

	// internal flags, when called programmatically
	NameOrBech32 string `flag:"-"`
	TxJson       []byte `flag:"-"`
	Pass         string `flag:"-"`
}

var DefaultSignOptions = SignOptions{
	BaseOptions: DefaultBaseOptions,
	TxPath:      "-", // read from stdin.
	ChainID:     "dev",
}

func signApp(cmd *command.Command, args []string, iopts interface{}) error {
	var opts SignOptions = iopts.(SignOptions)
	var err error

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: sign <keyname or address>")
		return errors.New("invalid args")
	}
	opts.NameOrBech32 = args[0]

	// read tx to sign
	txpath := opts.TxPath
	if txpath == "-" { // from stdin.
		txjsonstr, err := cmd.GetString("Enter tx to sign, terminated by a newline.")
		if err != nil {
			return err
		}
		opts.TxJson = []byte(txjsonstr)
	} else { // from file
		opts.TxJson, err = os.ReadFile(txpath)
		if err != nil {
			return err
		}
	}

	if opts.Quiet {
		opts.Pass, err = cmd.GetPassword("", opts.InsecurePasswordStdin)
	} else {
		opts.Pass, err = cmd.GetPassword("Enter password.", opts.InsecurePasswordStdin)
	}
	if err != nil {
		return err
	}

	signedTx, err := SignHandler(opts)
	if err != nil {
		return err
	}

	signedJson, err := amino.MarshalJSON(signedTx)
	if err != nil {
		return err
	}
	cmd.Println(string(signedJson))

	return nil
}

func SignHandler(opts SignOptions) (*std.Tx, error) {
	var err error
	var tx std.Tx

	if opts.AccountNumber == nil {
		return nil, errors.New("invalid account number")
	}
	if opts.Sequence == nil {
		return nil, errors.New("invalid sequence")
	}
	if opts.TxJson == nil {
		return nil, errors.New("invalid tx content")
	}

	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return nil, err
	}

	err = amino.UnmarshalJSON(opts.TxJson, &tx)
	if err != nil {
		return nil, err
	}

	// fill tx signatures.
	signers := tx.GetSigners()
	if tx.Signatures == nil {
		for range signers {
			tx.Signatures = append(tx.Signatures, std.Signature{
				PubKey:    nil, // zero signature
				Signature: nil, // zero signature
			})
		}
	}

	// validate document to sign.
	err = tx.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// derive sign doc bytes.
	chainID := opts.ChainID
	accountNumber := *opts.AccountNumber
	sequence := *opts.Sequence
	signbz := tx.GetSignBytes(chainID, accountNumber, sequence)
	if opts.ShowSignBytes {
		fmt.Printf("sign bytes: %X\n", signbz)
		return nil, nil
	}

	sig, pub, err := kb.Sign(opts.NameOrBech32, opts.Pass, signbz)
	if err != nil {
		return nil, err
	}
	addr := pub.Address()
	found := false
	for i := range tx.Signatures {
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
		return nil, errors.New("addr %v (%s) not in signer set",
			addr, opts.NameOrBech32)
	}

	return &tx, nil
}
