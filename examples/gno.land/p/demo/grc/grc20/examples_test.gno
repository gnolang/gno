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

/*
Foo20 := NewBank("foo20", "FOO", 5)



Foo20.PrevRealmBanker().Transfer()

local := Foo20.PrevRealmBanker()
local.Transfer()

func (b Bank) RealmIssueBanker() IssueBanker {
	if caller != creator { panic}
}
*/

/*
type Bank interface { // XXX Ledger?
	// GRC20

	PrevRea0lmBanker() Banker                // most standard banker -> proxifed 1-1 on foo20.gno;  exposable object, and also perfect for dexes
	PrevRealmNamedBanker(name string) Banker // most standard banker -> proxifed 1-1 on foo20.gno;  exposable object, and also perfect for dexes
	LocalBanker() Banker                     // when a realm wants to make advanced workflows, such as a dao that will "give" delegation of their account to a specificed callback
	LocalNamedBanker(name string) Banker     // when a realm wants to make advanced workflows, such as a dao that will "give" delegation of their account to a specificed callback

	RealmIssueBanker() IssueBanker // panic if called from non-owning realm
}
*/

// Spender???

/*
func ReadonlyWrapper(Banker) Banker {}
func Mockify(Banker) MockBanker     { /* readonly + changes  }

type Bank interface { // XXX Ledger?
	PrevRealmBanker() BankerWithOptionToBecomeLocal // most standard banker -> proxifed 1-1 on foo20.gno;  exposable object, and also perfect for dexes
	LocalBanker() Banker                            // when a realm wants to make advanced workflows, such as a dao that will "give" delegation of their account to a specificed callback
	ReadonlyBanker() ReadonlyBanker                 // XXX: ReadOnlyBanker?
	RealmIssueBanker() IssueBanker                  // panic if called from non-owning realm
}

type PrevRealmBanker interface {
	GRC20
	LocalBanker() Banker
}

type IssueBanker interface {
	Mint()
	Burn()
	Transfer(from, to std.Address, amount uint64)
}

type GRC20 interface{}

func ExampleBucket() {
	foo20 := grc20.NewBuilder("Foo20", "FOO", 5)
	foo20.Mint(addr1, 42)

	grc20 := foo20.TokenWithBucketName("mysecondary")
	grc20.Transfer(to, amount)
}

func ExampleUserBanker() {
	//
}

func ExampleRealmBanker() {
	//
}

func ExampleDex() {
	//
}

func ScopedBanker()
*/
/*
// --- grc20 interface v1
type Banker interface{} //?

type GRC20 interface {
	Transfer()
}
type SuperTokenManager interface {
	LocalBanker() GRC20
	UserBanker() GRC20                // PrevRealm
	DerivedAccountBanker(path string) // banker
}

type UserTokenManager interface {
	GRC20
	LocalBanker() LocalBanker
}

//

// mydao/dao.gno
var proposals []paymentProposal
var paymentProposal struct{to std.Address, amount uint64, denom string}
func ProposeTransfer(to std.Address, amount uint64, denom string){
	proposals = append(proposals, paymentProposal{to, amount, denom})
}

func ProposeTransfer(executeFn func(token grc20)) {
	proposals = append(proposals, executefn)
	//proposals = append(proposals, paymentProposal{to, amount, denom})
}


func ExecuteProposals() {
	for _, proposal := range proposals {
		localBanker := grc20.CreateLocalBanker(receivedToken)
		//token := grc20reg.Get(denom)
		//token.Transfer(proposal.To, proposal.Amount)
		proposal.executor(localBanker)
	}
}

//---

// next/foo20.gno
var (
	banker = grc20.NewRootBanker("params") // Transfer(from, to, amount)
	Token  = banker.TokenWithCoolOptions() CoolPublicToken           // Tranfer(to, amount) // this is safe to be exposed, because it checks PrevRealm()
)

//
import foo20

foo20.Transfer()
bk := foo20.LocalBanker()

// opt1: use the root API: banker.Transfer(from, to, amount) // super power
// opt2: create a grc20 compatible wrapper: MyPubBanker := banker.PrevRealmSafeBanker() // + consider exposing it with proxy functions

// next/manfred.gno
var localBanker = foo20.LocalBanker() // i can use this object myself ()

// next/grc20factory.gno

var tokens []grc20.MegaBanker

// grc20reg/registry.gno
grc20reg.Register(foo20)

// dex/dex.gno
import "grc20reg"

func SwapByImpls(a, b grc20.Token) {}

func SwapByName(a, b string) {
	ta := grc20reg.GetByName(a)
}

// cur/foo20.gno
var (
	banker = grc20.NewBanker("params")
	Token  = banker.Token() // this is safe to be exposed, because it checks PrevRealm()
)

// gnokey friendly
func TransferOpt1(to std.Address, amount unint64) { Token.Transfer(to, amount) }

// cur/bar20.gno
var Token = banker.Token() // this is safe to be exposed, because it checks PrevRealm()
func init() {
	banker := grc20.NewBanker("params")
	banker.Mint(adminAddr, 100000)
	Token = banker.Token()
	// don't keep a reference to banker
}

// dex/dex.gno

func Swap(coinA, coinB grc20.Token) {}

func Example() {
	Swap(foo20.Token, bar20.Token)
}

// govdao/prop_v1.gno
// import "/foo20"
func ProposePaying(to std.Address, amount uint64) {
	foo20.Transfer(to, amount)
}

// govdao/grc20_prop_wrapper.gno
var localBanker = foo20.LocalBanker()

func WrapProposal(fn func(localBanker)) {
	fn(localBanker)
}

// manfred/daoprop.gno
func Example() {
	govdao.WarpProposal(func(banker) {
		banker.Transfer(fooAddr, 42)
	})
}

//---

type PermissionedBanker struct{}

   func (mb *PermissionedBanker) Transfer(from, to std.Address, amount uint64) {}
   func (mb *PermissinlessBanker) Transfer(to std.Address, amount uint64) {
   from := mb.resolveFn()
   }

func (mb *PermissionedBanker) CustomBanker(callerFn func() std.Address) GRC20 {
	// wrap
}

func PrevRealmBanker(mb *PermissionedBanker) GRC20 {
        resolveFn := func() std.Address {
		return std.PrevRealm().Addr()
	}
        return mb.CustomBanker(resolveFn)
}

   func LocalRealmBanker(mb *PermissionedBanker) GRC20 {
	rlm := std.PrevRealm().Addr()
        resolveFn := func() std.Address {
		return rlm
	}
        return mb.CustomBanker(resolveFn)
   }

   func LocalRealmBankerWithSuffix(mb *PermissionedBanker, suffix string) GRC20 {
	rlm := std.PrevRealm().Addr()
	return mb.CustomBanker(func() string { return rlm + "/"+suffix })
        }

func MockBanker(mb *PermissionedBanker) GRC20 {
	return mb.CustomBanker(func() std.Address { })
}


///
*/
