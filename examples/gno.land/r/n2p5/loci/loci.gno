package loci

import (
	"chain"
	"chain/runtime"
	"encoding/base64"

	"gno.land/p/n2p5/loci"
	"gno.land/p/nt/ufmt"
)

var store *loci.LociStore

func init() {
	store = loci.New()
}

// Set takes a base64 encoded string and stores it in the Loci store.
// Keyed by the address of the caller. It also emits a "set" event with
// the address of the caller.
func Set(_ realm, value string) {
	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}
	store.Set(b)
	chain.Emit("SetValue", "ForAddr", string(runtime.PreviousRealm().Address()))
}

// Get retrieves the value stored at the provided address and
// returns it as a base64 encoded string.
func Get(_ realm, addr address) string {
	return base64.StdEncoding.EncodeToString(store.Get(addr))
}

func Render(path string) string {
	if path == "" {
		return `
# Welcome to Loci

Loci is a simple key-value store keyed by the caller's gno.land address. 
Only the caller can set the value for their address, but anyone can 
retrieve the value for any address. There are only two functions: Set and Get.
If you'd like to set a value, simply base64 encode any message you'd like and
it will be stored in in Loci. If you'd like to retrieve a value, simply provide 
the address of the value you'd like to retrieve.

For convenience, you can also use gnoweb to view the value for a given address,
if one exists. For instance append :g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t to
this URL to view the value stored at that address.
`
	}
	return renderGet(cross, address(path))
}

func renderGet(cur realm, addr address) string {
	value := "```\n" + Get(cur, addr) + "\n```"

	return ufmt.Sprintf(`
# Loci Value Viewer

**Address:** %s

%s

`, addr, value)
}
