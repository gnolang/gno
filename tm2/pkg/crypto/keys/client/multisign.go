package client

import (
	"context"
	"errors"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidMultisigKey = errors.New("provided key is not a multisig reference")

type MultisignCfg struct {
	RootCfg *BaseCfg

	TxPath     string
	Signatures commands.StringArr
}

func NewMultisignCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &MultisignCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "multisign",
			ShortUsage: "multisign [flags] <multisig key-name or address>",
			ShortHelp:  "aggregates the multiple signatures, signs the given tx document and saves it to disk",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMultisign(cfg, args, io)
		},
	)
}

func (c *MultisignCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.TxPath,
		"tx-path",
		"",
		"path to the Amino JSON-encoded tx (file) to sign",
	)

	fs.Var(
		&c.Signatures,
		"signature",
		"the path to the signature.json",
	)
}

func execMultisign(cfg *MultisignCfg, args []string, io commands.IO) error {
	return nil // TODO
}
