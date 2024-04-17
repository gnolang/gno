import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

var (
	mytoken *grc20.AdminToken
	admin   std.Address
)

// init is called once at time of deployment
func init() {
	// Set deployer of Realm to admin
	admin = std.PrevRealm().Addr()

	// Set token name, symbol and number of decimals
	mytoken = grc20.NewAdminToken("My Token", "TKN", 4)

	// Mint 1 million tokens to admin
	mytoken.Mint(admin, 1000000*10000)
}

