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

func newEnvCmd(io *commands.IO) *commands.Command {
	c := &envCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "env",
			ShortUsage: "env [flags] <pkgsym>",
			ShortHelp:  "`env` prints Gno environment information",
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
		"Prints the environment in JSON format instead of as a shell script.",
	)
}

type envVar struct {
	Key   string
	Value string
}

type envVars []envVar

func (vars *envVars) Set(key, value string) {
	for _, env := range *vars {
		if env.Key == key {
			env.Value = value
		}
	}
	*vars = append(*vars, envVar{Key: key, Value: value})
}

func (vars envVars) Get(key string) string {
	for _, env := range vars {
		if env.Key == key {
			return env.Value
		}
	}
	return ""
}

type envPrinter func(vars envVars, io *commands.IO)

func execEnv(cfg *envCfg, args []string, io *commands.IO) error {
	envs := envVars{}

	// GNOROOT:
	// Should point to the local location of the GNO repository.
	// It serves as the gno equivalent of `GOROOT`.
	envs.Set("GNOROOT", gnoenv.MustGuessGnoRootDir())

	// GNOHOME:
	// Should point
	envs.Set("GNOHOME", gnoenv.HomeDir())

	// Setup filters
	filters := envVars{}
	for _, arg := range args {
		filters.Set(arg, envs.Get(arg))
	}

	// Setup printer
	var printerEnv envPrinter
	{
		if cfg.json {
			printerEnv = printJSON
		} else {
			printerEnv = getPrinterShell(len(args) == 0)
		}
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
	return func(vars envVars, io *commands.IO) {
		for _, env := range vars {
			if printkeys {
				io.Printf("%s=%q\n", env.Key, env.Value)
			} else {
				io.Printf("%s\n", env.Value)
			}

		}
	}
}

func printJSON(vars envVars, io *commands.IO) {
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
