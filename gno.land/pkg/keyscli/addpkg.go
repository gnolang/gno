package keyscli

import (
	"context"
	"flag"
	"fmt"

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
}

func execMakeAddPkg(cfg *MakeAddPkgCfg, args []string, io commands.IO) error {
	if cfg.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if cfg.PkgDir == "" {
		return errors.New("pkgdir not specified")
	}

	if len(args) != 1 {
		return flag.ErrHelp
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
	memPkg := gno.ReadMemPackage(cfg.PkgDir, cfg.PkgPath)
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
		Creator: creator,
		Package: memPkg,
		Deposit: deposit,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		if cfg.RootCfg.Client == nil {
			return errors.New("rpcClient hasn't been initialized")
		}

		err := client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
