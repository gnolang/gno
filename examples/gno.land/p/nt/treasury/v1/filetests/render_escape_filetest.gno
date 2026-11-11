// PKGPATH: gno.land/r/treasury/main

// Token keys (lister map keys) and payment strings are rendered by the
// treasury pages; markdown in them must stay inert.
package main

import (
	"chain"
	"strings"

	"gno.land/p/nt/grc20/v1"
	"gno.land/p/nt/treasury/v1"
)

const evilKey = "[Claim](https://evil.example)"

func main(cur realm) {
	owner := chain.PackageAddress("gno.land/r/treasury/main")
	dest := chain.PackageAddress("gno.land/r/dest/main")

	token, ledger := grc20.NewToken("Tok", "TOK", 0, 0, cur)
	ledger.Mint(owner, 10)
	lister := func() map[string]*grc20.Token { return map[string]*grc20.Token{evilKey: token} }
	gb, err := treasury.NewGRC20BankerWithOwner(owner, lister)
	if err != nil {
		panic(err)
	}
	t, err := treasury.New([]treasury.Banker{gb}, "gno.land/r/treasury/main")
	if err != nil {
		panic(err)
	}
	if err := t.Send(0, cur, treasury.NewGRC20Payment(evilKey, 1, dest)); err != nil {
		panic(err)
	}

	for _, path := range []string{"GRC20", "GRC20/history"} {
		out := t.Render(path)
		println(path, "live link:", strings.Contains(out, evilKey))
		println(path, "escaped:", strings.Contains(out, `\[Claim\]`))
	}
}

// Output:
// GRC20 live link: false
// GRC20 escaped: true
// GRC20/history live link: false
// GRC20/history escaped: true
