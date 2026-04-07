package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type queryCfg struct {
	render bool
}

func (c *queryCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.render, "render", false, "call Render() instead of QEval")
}

func newQueryCmd(base *baseCfg, io commands.IO) *commands.Command {
	cfg := &queryCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "query",
			ShortUsage: "gnosh query [flags] <pkg-path> <expression>",
			ShortHelp:  "Read-only query (qeval or render).",
			LongHelp: `Execute a read-only query against a realm.

By default, uses QEval to evaluate an expression.
Use --render to call the Render() function instead.

Examples:
  gnosh query gno.land/r/demo/boards 'GetBoardIDFromName("testboard")'
  gnosh query --render gno.land/r/demo/boards ""
  gnosh query --render gno.land/r/demo/boards "testboard"`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execQuery(ctx, base, cfg, args, io)
		},
	)
}

func execQuery(_ context.Context, base *baseCfg, cfg *queryCfg, args []string, io commands.IO) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: gnosh query <pkg-path> <expression>")
	}

	pkgPath := args[0]
	expression := args[1]

	client, err := base.queryClient()
	if err != nil {
		return err
	}

	var result string
	if cfg.render {
		result, _, err = client.Render(pkgPath, expression)
	} else {
		result, _, err = client.QEval(pkgPath, expression)
	}
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	if base.json {
		return outputJSON(io, map[string]any{
			"pkg_path":   pkgPath,
			"expression": expression,
			"result":     result,
		})
	}

	io.Println(result)
	return nil
}
