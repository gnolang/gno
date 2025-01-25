package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type envCfg struct {
	json bool
}

func newEnvCmd(io commands.IO) *commands.Command {
	c := &envCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "env",
			ShortUsage: "env [flags] <pkgsym>",
			ShortHelp:  "print gno environment information",
		},
		c,
		func(_ context.Context, args []string) error {
			return execEnv(c, args, io)
		},
	)
}

func (c *envCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.json,
		"json",
		false,
		"prints the environment in JSON format instead of as a shell script.",
	)
}

type envVar struct {
	Key   string
	Value string
}

func findEnv(env []envVar, name string) string {
	for _, e := range env {
		if e.Key == name {
			return e.Value
		}
	}
	return ""
}

type envPrinter func(vars []envVar, io commands.IO)

func execEnv(cfg *envCfg, args []string, io commands.IO) error {
	envs := []envVar{
		// GNOROOT Should point to the local location of the GNO repository.
		// It serves as the gno equivalent of `GOROOT`.
		{Key: "GNOROOT", Value: gnoenv.RootDir()},
		// GNOHOME Should point to the user local configuration.
		// The most common place for this should be $HOME/gno.
		{Key: "GNOHOME", Value: gnoenv.HomeDir()},
	}

	// Setup filters
	filters := make([]envVar, len(args))
	for i, arg := range args {
		filters[i] = envVar{Key: arg, Value: findEnv(envs, arg)}
	}

	// Setup printer
	var printerEnv envPrinter
	if cfg.json {
		printerEnv = printJSON
	} else {
		printerEnv = getPrinterShell(len(args) == 0)
	}

	// Print environements
	if len(filters) > 0 {
		printerEnv(filters, io)
	} else {
		printerEnv(envs, io)
	}

	return nil
}

func getPrinterShell(printkeys bool) envPrinter {
	return func(vars []envVar, io commands.IO) {
		for _, env := range vars {
			if printkeys {
				io.Printf("%s=%q\n", env.Key, env.Value)
			} else {
				io.Printf("%s\n", env.Value)
			}
		}
	}
}

func printJSON(vars []envVar, io commands.IO) {
	io.Println("{")
	for i, env := range vars {
		io.Printf("\t%q: %q", env.Key, env.Value)
		if i != len(vars)-1 {
			io.Printf(",")
		}

		// Jump to next line
		io.Printf("\n")
	}
	io.Println("}")
}
