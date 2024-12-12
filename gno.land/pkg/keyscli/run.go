package keyscli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

type MakeRunCfg struct {
	RootCfg *client.MakeTxCfg
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

func (c *MakeRunCfg) RegisterFlags(fs *flag.FlagSet) {}

func execMakeRun(cfg *MakeRunCfg, args []string, cmdio commands.IO) error {
	if len(args) != 2 {
		return flag.ErrHelp
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

	memPkg := &gnovm.MemPackage{}
	if sourcePath == "-" { // stdin
		data, err := io.ReadAll(cmdio.In())
		if err != nil {
			return fmt.Errorf("could not read stdin: %w", err)
		}
		memPkg.Files = []*gnovm.MemFile{
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
			memPkg = gno.MustReadMemPackage(sourcePath, "")
		} else { // is file
			b, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("could not read %q: %w", sourcePath, err)
			}
			memPkg.Files = []*gnovm.MemFile{
				{
					Name: info.Name(),
					Body: string(b),
				},
			}
		}
	}
	if memPkg.IsEmpty() {
		panic(fmt.Sprintf("found an empty package %q", memPkg.Path))
	}

	memPkg.Name = "main"
	// Set to empty; this will be automatically set by the VM keeper.
	memPkg.Path = ""

	// construct msg & tx and marshal.
	return client.MakeTransaction(vm.MsgRun{
		Caller:  caller,
		Package: memPkg,
	}, cfg.RootCfg, args, cmdio)
}
