package main

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	"github.com/gnolang/gno/pkgs/crypto/multisig"
	"github.com/gnolang/gno/pkgs/errors"
)

/*
type ListOptions struct {
	client.BaseOptions // home, ...
}

var DefaultListOptions = ListOptions{
	BaseOptions: client.DefaultBaseOptions,
}
*/

func listBkApp(cmd *command.Command, args []string, iopts interface{}) error {

	if len(args) != 0 {
		cmd.ErrPrintfln("Usage: list (no args)")
		return errors.New("invalid args")
	}

	opts := iopts.(client.ListOptions)
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}

	bkKeyBase, err := keys.NewBkKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err != nil {

		return err

	}

	printInfos(cmd, infos, "primary")

	cmd.Println("\n---------------------------")

	infos, err = bkKeyBase.List()
	if err != nil {

		return err

	}
	printInfos(cmd, infos, "backup")

	return nil
}

func printInfos(cmd *command.Command, infos []keys.Info, keybaseName string) {

	cmd.Printfln("Keybase %s", keybaseName)
	var keypubString string

	for i, info := range infos {
		keyname := info.GetName()
		keytype := info.GetType()
		keypub := info.GetPubKey()
		keyaddr := info.GetAddress()
		keypath, _ := info.GetPath()
		keypubString = ""

		if mPub, ok := keypub.(multisig.PubKeyMultisigThreshold); ok {

			for _, pub := range mPub.PubKeys {

				keypubString = keypubString + pub.String() + " | "
			}

		} else {

			keypubString = keypub.String()
		}

		cmd.Printfln("%d. %s (%s) - addr: %v pub: %v, path: %v\n",
			i, keyname, keytype, keyaddr, keypubString, keypath)

		//TODO: implement PubKeyMultisigThreshold.String()

	}
}
