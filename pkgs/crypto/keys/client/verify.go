package client

import (
	"encoding/hex"
	"io/ioutil"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

type VerifyOptions struct {
	BaseOptions        // home,...
	DocPath     string // path of document file to verify
}

var DefaultVerifyOptions = VerifyOptions{
	DocPath: "", // read from stdin.
}

func runVerifyCmd(cmd *command.Command) error {
	var kb keys.Keybase
	var err error
	var opts VerifyOptions = cmd.Options.(VerifyOptions)
	var args = cmd.Args

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

	// verify signature.
	err = kb.Verify(name, msg, sig)
	return err
}

func parseSignature(sigstr string) ([]byte, error) {
	sig, err := hex.DecodeString(sigstr)
	if err != nil {
		return nil, err
	}
	return sig, nil
}
