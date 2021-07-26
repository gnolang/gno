package client

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

type ListOptions struct {
	BaseOptions // home, ...
}

var DefaultListOptions = ListOptions{}

func listApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(ListOptions)
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
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
		fmt.Println(">>> [XXX TODO IMPROVE LISTING]", i, info)
	}
}
