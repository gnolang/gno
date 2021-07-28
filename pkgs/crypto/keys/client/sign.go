package client

import (
	"io/ioutil"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
)

type SignOptions struct {
	BaseOptions        // home,...
	DocPath     string `flag:"docpath", help:"path of document file to sign"`
}

var DefaultSignOptions = SignOptions{
	DocPath: "", // read from stdin.
}

func signApp(cmd *command.Command, args []string, iopts interface{}) error {
	var kb keys.Keybase
	var err error
	var opts SignOptions = iopts.(SignOptions)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: sign <keyname>")
		return errors.New("invalid args")
	}

	name := args[0]
	docpath := opts.DocPath
	kb, err = keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	msg := []byte{}

	// read document to sign
	if docpath == "" { // from stdin.
		msgstr, err := cmd.GetString("Enter document text to sign, terminated by a newline.")
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

	pass, err := cmd.GetPassword("Enter password.")
	if err != nil {
		return err
	}
	sig, pub, err := kb.Sign(name, pass, msg)
	if err != nil {
		return err
	}

	cmd.Printfln("Signature: %X\nPub: %v", sig, pub)
	return nil
}
