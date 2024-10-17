package grc20

// XXX: write Examples

func ExampleInit()                            {}
func ExampleExposeBankForMaketxRunOrImports() {}
func ExampleCustomGRC20Impl()                 {}
func ExampleAllowance()                       {}
func ExampleRealmBanker()                     {}
func ExamplePrevRealmBanker()                 {}
func ExampleAccountBanker()                   {}
func ExampleTransfer()                        {}
func ExampleApprove()                         {}
func ExampleTransferFrom()                    {}
func ExampleMint()                            {}
func ExampleBurn()                            {}

/*
 Users
 - token creator (root)
 - token holder / enduser / basic realm (GRC20)
 - dao/intermediary-realm (vault) contracts
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
