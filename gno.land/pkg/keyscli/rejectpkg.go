package keyscli

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeRejectPkgCfg struct {
	RootCfg *client.MakeTxCfg
	PkgPath string
}

func NewMakeRejectPkgCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &MakeRejectPkgCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "rejectpkg",
			ShortUsage: "rejectpkg [flags] <key-name>",
			ShortHelp:  "removes a package awaiting approval",
			LongHelp: `Removes a package that is parked awaiting approval, whether because
an approver is declining it or because its creator is withdrawing it. Either
may send this; nobody else.

No content hash, unlike enablepkg: deleting the wrong bytes is not the hazard
that activating the wrong bytes is.

The submission charge is not refunded.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeRejectPkg(cfg, args, io)
		},
	)
}

func (c *MakeRejectPkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PkgPath,
		"pkgpath",
		"",
		"package path to remove from the queue (required)",
	)
}

func execMakeRejectPkg(cfg *MakeRejectPkgCfg, args []string, io commands.IO) error {
	if cfg.PkgPath == "" {
		return errors.New("pkgpath not specified")
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

	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}

	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee")
	}

	tx := std.Tx{
		Msgs: []std.Msg{vm.MsgRejectPackage{
			Sender:  info.GetAddress(),
			PkgPath: cfg.PkgPath,
		}},
		Fee:        std.NewFee(cfg.RootCfg.GasWanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		cfg.RootCfg.RootCfg.OnTxSuccess = PrintTxSuccess
		return client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
	}
	io.Println(string(amino.MustMarshalJSON(tx)))
	return nil
}
