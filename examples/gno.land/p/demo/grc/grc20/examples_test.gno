package grc20

/*
 Users
 - token creator (root)
 - token holder / enduser / basic realm (GRC20)
 - dao/intermediary-realm (vault) contracts
 Some challenges I've in mind:
 - making a `GRC20` interface that is an interface. -> CHECK
 - making a `grc20.XXX` structure which provides the best suggested implementation.
 - encouraging realms to store an admin Banker while exposing other things (TBD).
 - user banker. (prevrealm)
 - realm banker.
 - compatible with govdao. (and async closures).
 - working from top-level variable.
 - working from from proxified.
 - compatible with custom implementation/rules.
 - compatible with custom getCaller as a callback?
 - support addresseable OR sub treasuries.
 - single bnaker for multiple grc20 tokens; like std.Coin{}
   - consider returning `std.Coin`
 - grc20factory.gno // var banks avl.Tree // key -> Bank
    bank.LocalBankerWithSuffix(key).Transfer()
 - boundingtoken.gno // tokenA, boundA
 - think about the r/grc20factory special contract
   - create new user token
   - what about being seen by another contract as the prevRealm (dex)
 - wugnot
 - GOOD NAMES!
 - GOOD NAMES!
 - GOOD NAMES!
 - GOOD NAMES!
*/
