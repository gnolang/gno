package keyscli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeRunCfg struct {
	RootCfg    *client.MakeTxCfg
	Send       string
	MaxDeposit string
}

func NewMakeRunCmd(rootCfg *client.MakeTxCfg, cmdio commands.IO) *commands.Command {
	cfg := &MakeRunCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] <key-name or address> <file or - or dir>",
			ShortHelp:  "runs Gno code by invoking main() in a package",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeRun(cfg, args, cmdio)
		},
	)
}

func (c *MakeRunCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.Send,
		"send",
		"",
		"send amount",
	)
	fs.StringVar(
		&c.MaxDeposit,
		"max-deposit",
		"",
		"max storage deposit",
	)
}

func execMakeRun(cfg *MakeRunCfg, args []string, cmdio commands.IO) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	if cfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	nameOrBech32 := args[0]
	sourcePath := args[1] // can be a file path, a dir path, or '-' for stdin

	// read account pubkey.
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	caller := info.GetAddress()

	// Parse send amount.
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// Parse deposit amount
	deposit, err := std.ParseCoins(cfg.MaxDeposit)
	if err != nil {
		return errors.Wrap(err, "parsing storage deposit coins")
	}

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	memPkg := &std.MemPackage{}
	if sourcePath == "-" { // stdin
		data, err := io.ReadAll(cmdio.In())
		if err != nil {
			return fmt.Errorf("could not read stdin: %w", err)
		}
		memPkg.Files = []*std.MemFile{
			{
				Name: "stdin.gno",
				Body: string(data),
			},
		}
	} else {
		info, err := os.Stat(sourcePath)
		if err != nil {
			return fmt.Errorf("could not read source path: %q, %w", sourcePath, err)
		}
		if info.IsDir() {
			memPkg = gno.MustReadMemPackage(sourcePath, "", gno.MPUserProd)
		} else { // is file
			b, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("could not read %q: %w", sourcePath, err)
			}
			memPkg.Files = []*std.MemFile{
				{
					Name: info.Name(),
					Body: string(b),
				},
			}
		}
	}

	memPkg.Name = "main"
	if memPkg.IsEmpty() {
		panic(fmt.Sprintf("found an empty package %q", memPkg.Path))
	}

	// Set to empty; this will be automatically set by the VM keeper.
	memPkg.Path = ""

	// construct msg & tx and marshal.
	msg := vm.MsgRun{
		Caller:     caller,
		Package:    memPkg,
		Send:       send,
		MaxDeposit: deposit,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		cfg.RootCfg.RootCfg.OnTxSuccess = func(tx std.Tx, res *ctypes.ResultBroadcastTxCommit) {
			PrintTxInfo(tx, res, cmdio)
		}
		err := client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, cmdio)
		if err != nil {
			return err
		}
	} else {
		cmdio.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
