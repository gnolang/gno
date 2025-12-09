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

MODES:
  local    Development mode with interactive features and lazy loading (default)
  staging  Server mode with JSON logging and all examples preloaded

PACKAGE DISCOVERY:
gnodev automatically discovers and loads packages based on your project structure:

  - gnomod.toml: If present in a directory, gnodev recognizes it as a Gno package
    and uses the module path defined within to load it.

  - gnowork.toml: If present, gnodev treats the directory as a workspace and
    discovers all packages within it.

When running 'gnodev local' from a directory with gnomod.toml, your package is
automatically detected and loaded. Additional directories can be passed as arguments.

EXAMPLES:
  gnodev                           Start in current directory (auto-detects gnomod.toml)
  gnodev ./myrealm                 Load package from ./myrealm directory
  gnodev -paths "gno.land/r/**"   Preload matching packages by path pattern

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
