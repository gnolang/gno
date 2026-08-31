// PKGPATH: gno.land/r/demo/grc20subrealm

// A frame-relative teller must keep working when the token's OWN realm
// operates under an identity minted by its own cur.Sub(). The sub's synthesized
// pkgpath is "<host>#vault", which is not origRealm verbatim, so a raw string
// comparison would refuse the token's own realm — a false positive that would
// break the one property the home binding promises unconditionally.
package grc20subrealm

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
)

var (
	token  *grc20.Token
	ledger *grc20.PrivateLedger
	teller grc20.Teller
)

func init(cur realm) {
	token, ledger = grc20.NewToken("SubRealm", "SUBR", 4, cur)
	teller = ledger.CallerTeller()
}

func main(cur realm) {
	payer := cur.Previous().Address() // whom the frame-relative teller debits
	bob := chain.PackageAddress("bob")
	ledger.Mint(payer, 1_000)

	// Control: the primary identity moves the caller's funds.
	if err := teller.Transfer(0, cur, bob, 100); err != nil {
		panic("primary identity rejected: " + err.Error())
	}

	// Same realm, same teller, under this realm's own "vault" sub identity.
	if err := teller.Transfer(0, cur.Sub("vault"), bob, 100); err != nil {
		panic("own sub identity rejected: " + err.Error())
	}

	if got := token.BalanceOf(bob); got != 200 {
		panic("unexpected bob balance")
	}

	// The mirror image: a token CREATED from a sub frame must still be usable
	// from its host realm. origRealm drops the subpath at construction, so the
	// token belongs to the host rather than being stranded in the sub.
	subTok, subLedger := grc20.NewToken("FromSub", "FSUB", 4, cur.Sub("vault"))
	subTeller := subLedger.CallerTeller()
	subLedger.Mint(payer, 500)
	if err := subTeller.Transfer(0, cur, bob, 50); err != nil {
		panic("host realm rejected for a sub-created token: " + err.Error())
	}
	if got := subTok.BalanceOf(bob); got != 50 {
		panic("unexpected bob balance on the sub-created token")
	}

	println("ok")
}

// Output:
// ok
