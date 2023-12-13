package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	baseCfg := keyscli.BaseOptions{
		Home:   gnoenv.HomeDir(),
		Remote: "127.0.0.1:26657",
	}

	cmd := keyscli.NewRootCmdWithBaseConfig(commands.NewDefaultIO(), baseCfg)
	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)

		os.Exit(1)
	}
}
