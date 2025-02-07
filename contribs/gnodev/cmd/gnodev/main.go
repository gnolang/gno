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
			LongHelp: `The gnodev command starts an in-memory node and a gno.land web interface, primarily for realm package development.

If no command is provided, gnodev will automatically start in <local> mode.
For more information and flags usage description, use 'gnodev local -h'.`,
		},
		nil,
		func(ctx context.Context, _ []string) error { return nil },
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
