package client

import (
	"io/ioutil"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

type SignOptions struct {
	BaseOptions        // home,...
	DocPath     string // path of document file to sign
}

var DefaultSignOptions = SignOptions{
	DocPath: "", // read from stdin.
}

func runSignCmd(cmd *command.Command) error {
	var kb keys.Keybase
	var err error
	var opts SignOptions = cmd.Options.(SignOptions)
	var args = cmd.Args

	name := args[0]
	docpath := opts.DocPath
	kb, err = keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	msg := []byte{}

	// read document to sign
	if docpath == "" { // from stdin.
		msgstr, err := cmd.GetString("enter document to sign.")
		if err != nil {
			return err
		}
		msg = []byte(msgstr)
	} else { // from file
		msg, err = ioutil.ReadFile(docpath)
		if err != nil {
			return err
		}
	}

	// validate document to sign.
	// XXX

	pass, err := cmd.GetPassword("enter password.")
	if err != nil {
		return err
	}
	sig, pub, err := kb.Sign(name, pass, msg)
	if err != nil {
		return err
	}

	cmd.Printfln("signature: %v\npub: %v", sig, pub)
	return nil
}
