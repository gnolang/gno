package client

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
)

type ListOptions struct {
	BaseOptions // home, ...
}

var DefaultListOptions = ListOptions{
	BaseOptions: DefaultBaseOptions,
}

func listApp(cmd *command.Command, args []string, iopts interface{}) error {
	if len(args) != 0 {
		cmd.ErrPrintfln("Usage: list (no args)")

		return errors.New("invalid args")
	}

	opts := iopts.(ListOptions)
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(cmd, infos)
	}
	return err
}

func printInfos(cmd *command.Command, infos []keys.Info) {
	for i, info := range infos {
		keyname := info.GetName()
		keytype := info.GetType()
		keypub := info.GetPubKey()
		keyaddr := info.GetAddress()
		keypath, _ := info.GetPath()
		cmd.Printfln("%d. %s (%s) - addr: %v pub: %v, path: %v",
			i, keyname, keytype, keyaddr, keypub, keypath)
	}
}
