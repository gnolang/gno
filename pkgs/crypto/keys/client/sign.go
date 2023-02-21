package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/std"
)

type signCfg struct {
	rootCfg *baseCfg

	txPath        string
	chainID       string
	accountNumber uint64
	sequence      uint64
	showSignBytes bool

	// internal flags, when called programmatically
	nameOrBech32 string
	txJSON       []byte
	pass         string
}

func newSignCmd(rootCfg *baseCfg) *commands.Command {
	cfg := &signCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "sign",
			ShortUsage: "sign [flags] <key-name or address>",
			ShortHelp:  "Signs the document",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSign(cfg, args, commands.NewDefaultIO())
		},
	)
}

func (c *signCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.txPath,
		"txpath",
		"-",
		"path to file of tx to sign",
	)

	fs.StringVar(
		&c.chainID,
		"chainid",
		"dev",
		"chainid to sign for",
	)

	fs.Uint64Var(
		&c.accountNumber,
		"number",
		0,
		"account number to sign with (required)",
	)

	fs.Uint64Var(
		&c.sequence,
		"sequence",
		0,
		"sequence to sign with (required)",
	)

	fs.BoolVar(
		&c.showSignBytes,
		"show-signbytes",
		false,
		"show sign bytes and quit",
	)
}

func execSign(cfg *signCfg, args []string, io *commands.IO) error {
	var err error

	if len(args) != 1 {
		return flag.ErrHelp
	}

	cfg.nameOrBech32 = args[0]

	// read tx to sign
	txpath := cfg.txPath
	if txpath == "-" { // from stdin.
		txjsonstr, err := io.GetString(
			"Enter tx to sign, terminated by a newline.",
		)
		if err != nil {
			return err
		}
		cfg.txJSON = []byte(txjsonstr)
	} else { // from file
		cfg.txJSON, err = os.ReadFile(txpath)
		if err != nil {
			return err
		}
	}

	if cfg.rootCfg.Quiet {
		cfg.pass, err = io.GetPassword(
			"",
			cfg.rootCfg.InsecurePasswordStdin,
		)
	} else {
		cfg.pass, err = io.GetPassword(
			"Enter password.",
			cfg.rootCfg.InsecurePasswordStdin,
		)
	}
	if err != nil {
		return err
	}

	signedTx, err := SignHandler(cfg)
	if err != nil {
		return err
	}

	signedJSON, err := amino.MarshalJSON(signedTx)
	if err != nil {
		return err
	}
	io.Println(string(signedJSON))

	return nil
}

func SignHandler(cfg *signCfg) (*std.Tx, error) {
	var err error
	var tx std.Tx

	if cfg.txJSON == nil {
		return nil, errors.New("invalid tx content")
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.Home)
	if err != nil {
		return nil, err
	}

	err = amino.UnmarshalJSON(cfg.txJSON, &tx)
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
	chainID := cfg.chainID
	accountNumber := cfg.accountNumber
	sequence := cfg.sequence
	signbz := tx.GetSignBytes(chainID, accountNumber, sequence)
	if cfg.showSignBytes {
		fmt.Printf("sign bytes: %X\n", signbz)
		return nil, nil
	}

	sig, pub, err := kb.Sign(cfg.nameOrBech32, cfg.pass, signbz)
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
		return nil, errors.New(
			fmt.Sprintf("addr %v (%s) not in signer set", addr, cfg.nameOrBech32),
		)
	}

	return &tx, nil
}
