package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

func main() {
	baseCfg := client.DefaultBaseOptions
	baseCfg.Home = gnoenv.HomeDir()

	cmd := keyscli.NewRootCmd(commands.NewDefaultIO(), baseCfg)
	cmd.Execute(context.Background(), os.Args[1:])
}
