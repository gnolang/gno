import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

var (
	mytoken *grc20.Token // main token object
	o       *ownable.Ownable // ownable to store admin of token
)

func init() {
	// Instantiate ownable
	o = ownable.New()

	// Instantiate token
	mytoken = grc20.NewGRC20Token("My Token", "MYT", 4)

	// Mint 1M tokens to admin
	mytoken.Mint(o.Owner(), uint64(1_000_000*10000))
}
