package gnoclient_test

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

func Example_withDisk() {
	kb, _ := keys.NewKeyBaseFromDir("/path/to/dir")
	signer := gnoclient.SignerFromKeybase{
		Keybase:  kb,
		Account:  "mykey",
		Password: "secure",
	}
	client := gnoclient.Client{
		Signer: signer,
	}
	_ = client
}

func Example_withInMemCrypto() {
	// create inmem keybase from bip39
	mnemo := "index brass unknown lecture autumn provide royal shrimp elegant wink now zebra discover swarm act ill you bullet entire outdoor tilt usage gap multiply"
	bip39Passphrase := ""
	account := uint32(0)
	index := uint32(0)
	signer, _ := gnoclient.SignerFromBip39(mnemo, bip39Passphrase, account, index)
	client := gnoclient.Client{
		Signer: signer,
	}
	_ = client
	fmt.Println("Hello")
	// Output:
	// Hello
}

func Example_readOnly() {
	// read-only client, can only query.
	client := gnoclient.Client{}
	_ = client
	fmt.Println("Hello")
	// Output:
	// Hello
}
