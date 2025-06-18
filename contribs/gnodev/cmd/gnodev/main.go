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
			LongHelp: `The gnodev command starts an in-memory node and a gno.land
web interface, primarily for realm package development.

Currently gnodev comes with two mode <local> and <staging>, those command mostly
differ by there default values, while gnodev local as default for working
locally, satging default are oriented to be use on server.

gnodev uses its own package loader and resolver system to support multiple
scenarios and use cases. It currently supports three types of resolvers, each
taking a location as an argument.
- root: This resolver takes a <dir> as its location. It attempts to resolve
  packages based on your file system structure and the package path.
  For example, if 'root=/user/gnome/myproject' and you try to resolve
  'gno.land/r/bar/buzz' as a package, the <root> resolver will attempt to
  resolve it to /user/gnome/myproject/gno.land/r/bar/buzz.
- local: This resolver also takes a <dir> as its location. It is designed to
  load a single package, using the module name from 'gnomod.toml' within this
  package to resolve the package.
- remote: This resolver takes a <remote> RPC address as its location. It is
  meant to use a remote node as a resolver, primarily for testing a local
  package against a remote node.

Resolvers can be chained, and gnodev will attempt to use them in the order they
are declared.

For example:
    gnodev -resolver root=/user/gnome/myproject -resolver remote=https://rpc.gno.lands

If no resolvers can resolve a given package path, the loader will return a
"package not found" error.

If no command is provided, gnodev will automatically start in <local> mode.

For more information and flags usage description, use 'gnodev local -h'.`,
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
