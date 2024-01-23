package client

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type SignCfg struct {
	RootCfg *BaseCfg

	TxPath        string
	ChainID       string
	AccountNumber uint64
	Sequence      uint64
	ShowSignBytes bool
	NameOrBech32  string
	TxJSON        []byte
	Pass          string
}

func NewSignCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &SignCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "sign",
			ShortUsage: "sign [flags] <key-name or address>",
			ShortHelp:  "Signs the document",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSign(cfg, args, io)
		},
	)
}

func (c *SignCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.TxPath,
		"txpath",
		"-",
		"path to file of tx to sign",
	)

	fs.StringVar(
		&c.ChainID,
		"chainid",
		"dev",
		"chainid to sign for",
	)

	fs.Uint64Var(
		&c.AccountNumber,
		"number",
		0,
		"account number to sign with (required)",
	)

	fs.Uint64Var(
		&c.Sequence,
		"sequence",
		0,
		"sequence to sign with (required)",
	)

	fs.BoolVar(
		&c.ShowSignBytes,
		"show-signbytes",
		false,
		"show sign bytes and quit",
	)
}

func execSign(cfg *SignCfg, args []string, io commands.IO) error {
	var err error

	if len(args) != 1 {
		return flag.ErrHelp
	}

	cfg.NameOrBech32 = args[0]

	// read tx to sign
	txpath := cfg.TxPath
	if txpath == "-" { // from stdin.
		txjsonstr, err := io.GetString(
			"Enter tx to sign, terminated by a newline.",
		)
		if err != nil {
			return err
		}
		cfg.TxJSON = []byte(txjsonstr)
	} else { // from file
		cfg.TxJSON, err = os.ReadFile(txpath)
		if err != nil {
			return err
		}
	}

	if cfg.RootCfg.Quiet {
		cfg.Pass, err = io.GetPassword(
			"",
			cfg.RootCfg.InsecurePasswordStdin,
		)
	} else {
		cfg.Pass, err = io.GetPassword(
			"Enter password.",
			cfg.RootCfg.InsecurePasswordStdin,
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

func SignHandler(cfg *SignCfg) (*std.Tx, error) {
	var err error
	var tx std.Tx

	if cfg.TxJSON == nil {
		return nil, errors.New("invalid tx content")
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return nil, err
	}

	err = amino.UnmarshalJSON(cfg.TxJSON, &tx)
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
	chainID := cfg.ChainID
	accountNumber := cfg.AccountNumber
	sequence := cfg.Sequence
	signbz := tx.GetSignBytes(chainID, accountNumber, sequence)
	if cfg.ShowSignBytes {
		fmt.Printf("sign bytes: %X\n", signbz)
		return nil, nil
	}

	sig, pub, err := kb.Sign(cfg.NameOrBech32, cfg.Pass, signbz)
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
			fmt.Sprintf("addr %v (%s) not in signer set", addr, cfg.NameOrBech32),
		)
	}

	return &tx, nil
}
