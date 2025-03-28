package keyscli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeAddPkgCfg struct {
	RootCfg *client.MakeTxCfg

	PkgPath string
	PkgDir  string
	Deposit string
	Meta    commands.StringArr
}

func NewMakeAddPkgCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &MakeAddPkgCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "addpkg",
			ShortUsage: "addpkg [flags] <key-name>",
			ShortHelp:  "uploads a new package",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeAddPkg(cfg, args, io)
		},
	)
}

func (c *MakeAddPkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PkgPath,
		"pkgpath",
		"",
		"package path (required)",
	)

	fs.StringVar(
		&c.PkgDir,
		"pkgdir",
		"",
		"path to package files (required)",
	)

	fs.StringVar(
		&c.Deposit,
		"deposit",
		"",
		"deposit coins",
	)

	fs.Var(
		&c.Meta,
		"meta",
		"metadata fields (format: field=value)",
	)
}

func execMakeAddPkg(cfg *MakeAddPkgCfg, args []string, io commands.IO) error {
	if cfg.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if cfg.PkgDir == "" {
		return errors.New("pkgdir not specified")
	}
	if cfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	if len(args) != 1 {
		return flag.ErrHelp
	}

	// parse metadata fields
	metadata, err := parseMetadataFields(cfg.Meta)
	if err != nil {
		return err
	}

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	creator := info.GetAddress()
	// info.GetPubKey()

	// parse deposit.
	deposit, err := std.ParseCoins(cfg.Deposit)
	if err != nil {
		panic(err)
	}

	// open files in directory as MemPackage.
	memPkg := gno.MustReadMemPackage(cfg.PkgDir, cfg.PkgPath)
	if memPkg.IsEmpty() {
		panic(fmt.Sprintf("found an empty package %q", cfg.PkgPath))
	}

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
	if err != nil {
		panic(err)
	}
	// construct msg & tx and marshal.
	msg := vm.MsgAddPackage{
		Creator:  creator,
		Package:  memPkg,
		Deposit:  deposit,
		Metadata: metadata,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		err := client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

func parseMetadataFields(meta commands.StringArr) ([]*vm.MetaField, error) {
	if len(meta) == 0 {
		return nil, nil
	}

	metadata := make([]*vm.MetaField, len(meta))
	for i, v := range meta {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid metadata field format, expected field=value")
		}

		name := strings.TrimSpace(parts[0])
		if name == "" {
			return nil, errors.New("empty metadata field name")
		}

		var value []byte
		if strings.TrimSpace(parts[1]) != "" {
			value = []byte(parts[1])
		}

		metadata[i] = &vm.MetaField{
			Name:  name,
			Value: value,
		}
	}
	return metadata, nil
}
