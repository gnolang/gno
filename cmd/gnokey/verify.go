package main

import (
	"context"
	"encoding/hex"
	"flag"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

type verifyCfg struct {
	rootCfg *baseCfg

	docPath string
}

func newVerifyCmd(rootCfg *baseCfg) *commands.Command {
	cfg := &verifyCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "verify [flags] <key-name> <signature>",
			ShortHelp:  "Verifies the document signature",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execVerify(cfg, args, commands.NewDefaultIO())
		},
	)
}

func (c *verifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.docPath,
		"docpath",
		"",
		"path of document file to verify",
	)
}

func execVerify(cfg *verifyCfg, args []string, io *commands.IO) error {
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
	docpath := cfg.docPath
	kb, err = keys.NewKeyBaseFromDir(cfg.rootCfg.Home)
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

	// verify signature.
	err = kb.Verify(name, msg, sig)
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
