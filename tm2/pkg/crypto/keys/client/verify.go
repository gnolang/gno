package client

import (
	"context"
	"encoding/hex"
	"flag"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type VerifyCfg struct {
	RootCfg *BaseCfg

	DocPath       string
	ChainID       string
	AccountNumber uint64
	Sequence      uint64
}

func NewVerifyCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &VerifyCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "verify [flags] <key-name> <signature>",
			ShortHelp:  "verifies the document signature",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execVerify(cfg, args, io)
		},
	)
}

func (c *VerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.DocPath,
		"docpath",
		"",
		"path of document file to verify",
	)

	// info to recontruct the message
	fs.StringVar(
		&c.ChainID,
		"chainid",
		"dev",
		"the ID of the chain",
	)

	fs.Uint64Var(
		&c.AccountNumber,
		"account-number",
		0,
		"account number to verify with",
	)

	fs.Uint64Var(
		&c.Sequence,
		"account-sequence",
		0,
		"account sequence to verify with",
	)
}

func execVerify(cfg *VerifyCfg, args []string, io commands.IO) error {
	var (
		kb  keys.Keybase
		err error
	)

	if len(args) != 2 {
		return flag.ErrHelp
	}

	name := args[0]
	sig, err := parseSignature(args[1])
	if err != nil {
		return err
	}
	docpath := cfg.DocPath
	kb, err = keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return err
	}
	msg := []byte{}

	// read document to sign
	if docpath == "" { // from stdin.
		msgstr, err := io.GetString(
			"Enter document to sign.",
		)
		if err != nil {
			return err
		}
		msg = []byte(msgstr)
	} else { // from file
		msg, err = os.ReadFile(docpath)
		if err != nil {
			return err
		}
	}

	// validate document to sign.
	// XXX

	// reconstruct the message hash
	stdTx := std.Tx{}
	err = amino.UnmarshalJSON(msg, &stdTx)
	if err != nil {
		io.Println("error in validate document")
	}
	origMessageHash, err := stdTx.GetSignBytes(cfg.ChainID, cfg.AccountNumber, cfg.Sequence)
	if err != nil {
		io.Println("error in reconstruct message hash")
	}

	// verify signature.
	err = kb.Verify(name, origMessageHash, sig)
	if err == nil {
		io.Println("Valid signature!")
	}
	return err
}

func parseSignature(sigstr string) ([]byte, error) {
	sig, err := hex.DecodeString(sigstr)
	if err != nil {
		return nil, err
	}
	return sig, nil
}
