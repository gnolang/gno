package std

// AssertOriginCall panics is [IsOriginCall] return false.
func AssertOriginCall() // injected

// IsOriginCall return true only if the calling method is invoked via a direct
// MsgCall. It returns false for all other cases, like if the calling method
// is invoked by an other method (even from the same realm or package).
// It also returns false every time when the transaction is broadcasted via
// MsgRun.
func IsOriginCall() bool // injected

func CurrentRealmPath() string             // injected
func GetChainID() string                   // injected
func GetHeight() int64                     // injected
func GetOrigSend() Coins                   // injected
func GetOrigCaller() Address               // injected
func CurrentRealm() Realm                  // injected
func PrevRealm() Realm                     // injected
func GetOrigPkgAddr() Address              // injected
func GetCallerAt(n int) Address            // injected
func GetBanker(bt BankerType) Banker       // injected
func DerivePkgAddr(pkgPath string) Address // injected

func EncodeBech32(prefix string, bz [20]byte) Address                 // injected
func DecodeBech32(addr Address) (prefix string, bz [20]byte, ok bool) // injected
