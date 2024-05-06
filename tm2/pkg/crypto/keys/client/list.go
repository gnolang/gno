package client

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

func NewListCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "list",
			ShortHelp:  "lists all keys in the keybase",
		},
		nil,
		func(_ context.Context, args []string) error {
			return execList(rootCfg, args, io)
		},
	)
}

func execList(cfg *BaseCfg, args []string, io commands.IO) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.Home)
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(infos, io)
	}

	return err
}

func printInfos(infos []keys.Info, io commands.IO) {
	for i, info := range infos {
		keyname := info.GetName()
		keytype := info.GetType()
		keypub := info.GetPubKey()
		keyaddr := info.GetAddress()
		keypath, _ := info.GetPath()
		io.Printfln("%d. %s (%s) - addr: %v pub: %v, path: %v",
			i, keyname, keytype, keyaddr, keypub, keypath)
	}
}
