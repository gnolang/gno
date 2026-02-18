package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	stdio := commands.NewDefaultIO()

	localcmd := NewLocalCmd(stdio) // default

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnodev",
			ShortUsage: "gnodev <cmd> [flags] ",
			ShortHelp:  "Runs an in-memory node and gno.land web server for development purposes.",
			LongHelp: `The gnodev command starts an in-memory node and a gno.land web interface
for realm package development.

LOAD MODES (-load flag):
  auto   Pre-load current workspace/package only (default for local)
  lazy   Load all packages on-demand
  full   Pre-load all discovered packages (default for staging)

PACKAGE DISCOVERY:
gnodev automatically discovers packages based on your project structure:

  - gnomod.toml: Marks a directory as a Gno package
  - gnowork.toml: Marks a directory as a workspace containing multiple packages

EXAMPLES:
  gnodev                     Start with auto-detection (pre-loads current package)
  gnodev -load=lazy          Start with on-demand loading only
  gnodev -load=full          Pre-load all discovered packages
  gnodev ./myrealm           Load package from ./myrealm directory
  gnodev -paths "gno.land/r/demo/**"   Pre-load additional packages

For detailed flags, use 'gnodev local -h' or 'gnodev staging -h'.`,
		},
		nil,
		func(ctx context.Context, _ []string) error {
			localcmd.Execute(ctx, os.Args[1:])
			return nil
		},
	)

	cmd.AddSubCommands(localcmd)
	cmd.AddSubCommands(NewStagingCmd(stdio))

	// XXX: This part is a bit hacky; it mostly configures the command to
	// use the local command as default, but still falls back on gnodev root
	// help if asked.
	var buff bytes.Buffer
	cmd.SetOutput(&buff)
	if err := cmd.Parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			localcmd.Execute(context.Background(), os.Args[1:])
			return
		}

		if buff.Len() > 0 {
			fmt.Fprint(stdio.Err(), buff.String())
		}

		return
	}

	if err := cmd.Run(context.Background()); err != nil {
		stdio.ErrPrintfln(err.Error())
	}
}
