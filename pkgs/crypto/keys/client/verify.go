package client

import (
	"encoding/hex"
	"os"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
)

type VerifyOptions struct {
	BaseOptions
	DocPath string `flag:"docpath" help:"path of document file to verify"`
}

var DefaultVerifyOptions = VerifyOptions{
	BaseOptions: DefaultBaseOptions,
	DocPath:     "", // read from stdin.
}

func verifyApp(cmd *command.Command, args []string, iopts interface{}) error {
	var kb keys.Keybase
	var err error
	var opts VerifyOptions = iopts.(VerifyOptions)

	if len(args) != 2 {
		cmd.ErrPrintfln("Usage: verify <keyname> <signature>")
		return errors.New("invalid args")
	}

	name := args[0]
	sig, err := parseSignature(args[1])
	if err != nil {
		return err
	}
	docpath := opts.DocPath
	kb, err = keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	msg := []byte{}

	// read document to sign
	if docpath == "" { // from stdin.
		msgstr, err := cmd.GetString("Enter document to sign.")
		if err != nil {
			return err
		}
		msg = []byte(msgstr)
	} else { // from file
		msg, err = os.ReadFile(docpath)
		if err != nil {
			return err
		}
	}

	// validate document to sign.
	// XXX

	// verify signature.
	err = kb.Verify(name, msg, sig)
	if err == nil {
		cmd.Println("Valid signature!")
	}
	return err
}

func parseSignature(sigstr string) ([]byte, error) {
	sig, err := hex.DecodeString(sigstr)
	if err != nil {
		return nil, err
	}
	return sig, nil
}
