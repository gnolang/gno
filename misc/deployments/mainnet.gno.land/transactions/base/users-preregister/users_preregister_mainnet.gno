// users_preregister_mainnet.gno runs as a MsgRun at genesis to register the
// initial mainnet namespaces in r/sys/users, alongside the universal system
// names that r/sys/namereg/v1's preregister seed adds on every chain.
//
// The r/sys/users controller gate is skipped at ChainHeight()==0 and
// r/sys/users/init.RegisterUser is genesis-only by construction (it panics
// after height 0), so this grants NO standing authority: nothing is
// whitelisted, and nothing survives genesis except the registrations
// themselves. GovDAO keeps full control over every name afterwards
// (rename/delete via r/sys/namereg/v1 proposals).
//
// Registration uses the bypass path (RegisterUserIgnoreCanonical, later-wins
// on canonical collisions), so this curated list must not confusable-collide
// with the namereg seed (archive, demo, gno, gnoland, gnolang, gov, nt, sys,
// x, test1) or with itself — checked by eye, and a mistake surfaces as a
// silently re-pointed canonical stem, not a build failure.
//
// Once r/sys/names.Enable runs (transactions/migration/names-enable/), each
// name's address also owns the matching deploy namespace (gno.land/r/<name>/...)
// from block 1.
package main

import (
	usersinit "gno.land/r/sys/users/init"
)

func main(cur realm) {
	usersinit.RegisterUser(cross(cur), "gnoswap", address("g1q5tjjv3kkslzwf2na9y9wk7e5a8ugph8xfsk04"))
	usersinit.RegisterUser(cross(cur), "onbloc", address("g18839tls6c4w9kkxtzthjst4qdcm4lgyntvtqvs"))
	usersinit.RegisterUser(cross(cur), "moul", address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"))
	usersinit.RegisterUser(cross(cur), "aeddi", address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m"))
	usersinit.RegisterUser(cross(cur), "aib", address("g1gkqe9c90tfuk2a7f07ygs8t826aff03vxasjsl"))
	usersinit.RegisterUser(cross(cur), "samcrew", address("g136j0m08pkm2lwwde9dmlx8uee26llent9s5cpf"))
	usersinit.RegisterUser(cross(cur), "howl", address("g1m0rgan0rla00ygmdmp55f5m0unvsvknluyg2a4"))
}
