package std

// AssertOriginCall panics if [IsOriginCall] returns false.
func AssertOriginCall() // injected

// IsOriginCall returns true only if the calling method is invoked via a direct
// MsgCall. It returns false for all other cases, like if the calling method
// is invoked by another method (even from the same realm or package).
// It also returns false every time when the transaction is broadcasted via
// MsgRun.
func IsOriginCall() bool // injected

func GetChainID() string     // injected
func GetChainDomain() string // injected
func GetHeight() int64       // injected

func GetOrigSend() Coins {
	den, amt := origSend()
	coins := make(Coins, len(den))
	for i := range coins {
		coins[i] = Coin{Denom: den[i], Amount: amt[i]}
	}
	return coins
}

func GetOrigCaller() Address {
	return Address(origCaller())
}

// CurrentRealm returns the [Realm] in which the caller is being executed.
// The value of CurrentRealm remains the same if called by a function within the
// same realm, or a function in a pure package. It will change if called by a
// realm with a different pkgpath.
//
// As an example, here is a sequence of function calls and the result of
// CurrentRealm() in each. In this example, main() is called in the context of a
// MsgRun transaction.
//
//	Function           | CurrentRealm result
//	main.main()        | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
//	-> main.helper()   | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
//	-> realm1.Fn()     | Realm{addr: "g1em9rtwnzspuwpqdsxk9ldrn6z7r60vzj2l8xuw", pkgPath: "gno.land/r/demo/realm1"}
//	-> realm1.helper() | Realm{addr: "g1em9rtwnzspuwpqdsxk9ldrn6z7r60vzj2l8xuw", pkgPath: "gno.land/r/demo/realm1"}
//	-> realm2.Fn()     | Realm{addr: "g1x5cn0ef9mtwfed7yfp0t0jqwc5zlhzqpnd9mpd", pkgPath: "gno.land/r/demo/realm2"}
func CurrentRealm() Realm {
	addr, path := getRealm(0)
	return Realm{Address(addr), path}
}

// PrevRealm returns the [Realm] which called the code being executed.
// The value of PrevRealm remains the same if called by a function within the
// same realm, or a function in a pure package. It will change if called by a
// realm with a different pkgpath.
//
// As an example, here is a sequence of function calls and the result of
// PrevRealm() in each. In this example, main() is called in the context of a
// MsgRun transaction.
//
//	Function           | PrevRealm result
//	main.main()        | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
//	-> main.helper()   | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
//	-> realm1.Fn()     | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
//	-> realm1.helper() | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
//	-> realm2.Fn()     | Realm{addr: "g1em9rtwnzspuwpqdsxk9ldrn6z7r60vzj2l8xuw", pkgPath: "gno.land/r/demo/realm1"}
func PrevRealm() Realm {
	addr, path := getRealm(1)
	return Realm{Address(addr), path}
}

func GetOrigPkgAddr() Address {
	return Address(origPkgAddr())
}

func GetCallerAt(n int) Address {
	return Address(callerAt(n))
}

// Variations which don't use named types.
func origSend() (denoms []string, amounts []int64)
func origCaller() string
func origPkgAddr() string
func callerAt(n int) string
func getRealm(height int) (address string, pkgPath string)
func assertCallerIsRealm()
