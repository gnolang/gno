package gnoclient_test

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

func Example_withDisk() {
	home := "/path/to/dir"
	account := "mykey"
	passwd := "secure"

	kb, _ := keys.NewKeyBaseFromDir(home)
	signer := gnoclient.Signer(kb, account, passwd)
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
	kb, _ := gnoclient.InmemKeybaseFromBip39(mnemo, bip39Passphrase, account, index)
	client := gnoclient.Client{
		Keybase: kb,
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
