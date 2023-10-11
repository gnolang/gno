package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

func main() {
	baseCfg := client.BaseOptions{
		Home:                  gnoenv.HomeDir(),
		Remote:                "127.0.0.1:26657",
		Quiet:                 false,
		InsecurePasswordStdin: false,
		Config:                "",
	}

	cmd := client.NewRootCmdWithBaseConfig(commands.NewDefaultIO(), baseCfg)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)

		os.Exit(1)
	}
}
