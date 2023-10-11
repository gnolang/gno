package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gnovm/pkg/gnoroot"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
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
			ShortHelp:  "Env prints Gno environment information",
		},
		c,
		func(_ context.Context, args []string) error {
			return execEnv(c, args, io)
		},
	)
}

func (c *envCfg) RegisterFlags(fs *flag.FlagSet) {
	// XXX: add flags
	fs.BoolVar(
		&c.json,
		"json",
		false,
		"prints the environment in JSON format instead of as a shell script.",
	)

	// XXX: does those go orignal flags make sense here ?
	// fs.BoolVar(
	// 	&c.unset,
	// 	"u",
	// 	false,
	// 	"unsets the default setting for the named environment variables",
	// )

	// fs.BoolVar(
	// 	&c.warp,
	// 	"w",
	// 	false,
	// 	"changes the default settings of the named environment variables to the given values",
	// )
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
	gnorootVar, _ := gnoroot.GuessGnoRootDir()

	envs := envVars{}

	// GNOROOT:
	// Should point to the local location of the GNO repository.
	// It serves as the gno equivalent of `GOROOT`.
	envs.Set("GNOROOT", gnorootVar)

	// GNOHOME:
	envs.Set("GNOHOME", client.HomeDir())

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
