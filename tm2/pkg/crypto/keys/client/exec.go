package client

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type execCfg struct {
	rootCfg *makeTxCfg
	send    string
}

func newExecCmd(rootCfg *makeTxCfg, io *commands.IO) *commands.Command {
	cfg := &execCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "exec",
			ShortUsage: "exec [flags] <key-name or address> <file or - or dir>",
			ShortHelp:  "Executes arbitrary Gno code",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execExec(cfg, args, io)
		},
	)
}

func (c *execCfg) RegisterFlags(fs *flag.FlagSet) {}

func execExec(cfg *execCfg, args []string, io *commands.IO) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	if cfg.rootCfg.gasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.rootCfg.gasFee == "" {
		return errors.New("gas-fee not specified")
	}

	nameOrBech32 := args[0]
	sourcePath := args[1] // can be a file path, a dir path, or '-' for stdin

	// read account pubkey.
	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.rootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	caller := info.GetAddress()

	// parse gas wanted & fee.
	gaswanted := cfg.rootCfg.gasWanted
	gasfee, err := std.ParseCoin(cfg.rootCfg.gasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	memPkg := &std.MemPackage{}
	if sourcePath == "-" { // stdin
		data, err := ioutil.ReadAll(io.In)
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
			memPkg = gno.ReadMemPackage(sourcePath, "")
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
	memPkg.Path = "gno.land/r/main"
	if memPkg.IsEmpty() {
		panic(fmt.Sprintf("found an empty package %q", memPkg.Path))
	}
	// precompile and validate syntax
	err = gno.PrecompileAndCheckMempkg(memPkg)
	if err != nil {
		panic(err)
	}

	// construct msg & tx and marshal.
	msg := vm.MsgExec{
		Caller:  caller,
		Package: memPkg,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.rootCfg.memo,
	}

	if cfg.rootCfg.broadcast {
		err := signAndBroadcast(cfg.rootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
