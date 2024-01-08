package std

func AssertOriginCall()                    // injected
func IsOriginCall() bool                   // injected
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
