// PKGPATH: gno.land/r/treasury/canonicaltest

package canonicaltest

import (
	"chain"
	"chain/banker"

	"gno.land/p/nt/treasury/v0"
)

// evilBanker embeds *treasury.CoinsBanker so it inherits ID/Send/Balances/
// Address via promotion. If treasury used a sealed-interface marker, this
// type would satisfy the interface and bypass the gate. The canonical-impl
// allowlist (IsCanonicalBanker / treasury.New's type switch) rejects it
// because type assertions are nominal: *evilBanker is not *CoinsBanker.
type evilBanker struct {
	*treasury.CoinsBanker
}

func main(cur realm) {
	ownerAddr := chain.PackageAddress("gno.land/r/treasury/canonicaltest")

	inner := banker.NewBanker(banker.BankerTypeRealmSend, cur)
	legit, err := treasury.NewCoinsBankerWithOwner(ownerAddr, inner)
	if err != nil {
		panic("failed to construct canonical banker: " + err.Error())
	}

	// Verify the helper accepts the canonical impl.
	if !treasury.IsCanonicalBanker(legit) {
		panic("canonical *CoinsBanker must pass IsCanonicalBanker")
	}

	// Verify the helper rejects an embedded-impl bypass attempt.
	evil := &evilBanker{CoinsBanker: legit}
	if treasury.IsCanonicalBanker(evil) {
		panic("embedded-impl wrapper must NOT pass IsCanonicalBanker")
	}

	// Verify treasury.New rejects the same bypass attempt.
	_, err = treasury.New([]treasury.Banker{evil})
	if err != treasury.ErrNonCanonicalBankerImpl {
		panic("expected ErrNonCanonicalBankerImpl from treasury.New; got: " + err.Error())
	}

	// And confirms the canonical banker still works.
	_, err = treasury.New([]treasury.Banker{legit})
	if err != nil {
		panic("canonical banker must be accepted: " + err.Error())
	}

	println("canonical allowlist OK")
}

// Output:
// canonical allowlist OK
