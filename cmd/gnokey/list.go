package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

func newListCmd(rootCfg *baseCfg) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "list [flags]",
			ShortHelp:  "Lists all keys in the keybase",
		},
		nil,
		func(_ context.Context, args []string) error {
			return execList(rootCfg, args)
		},
	)
}

func execList(cfg *baseCfg, args []string) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.Home)
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(infos)
	}

	return err
}

func printInfos(infos []keys.Info) {
	for i, info := range infos {
		keyname := info.GetName()
		keytype := info.GetType()
		keypub := info.GetPubKey()
		keyaddr := info.GetAddress()
		keypath, _ := info.GetPath()
		fmt.Printf("%d. %s (%s) - addr: %v pub: %v, path: %v\n",
			i, keyname, keytype, keyaddr, keypub, keypath)
	}
}
