# PACKAGE DOCUMENTATION

## package chain

Package **chain** provides core chain primitives, events, chain metadata, and
realm helpers.

```
import "chain"
```

### Index

* [func ChainDomain() string](#chaindomain)
* [func ChainID() string](#chainid)
* [func ChainHeight() int64](#chainheight)
* [func Emit(typ string, attrs ...string)](#emit)
* [func CoinDenom(pkgPath, coinName string) string](#coindenom)
* [type Address](#address)

  * [func (Address) IsValid() bool](#address-isvalid)
  * [func (Address) String() string](#address-string)
* [type Coin](#coin)

  * [func NewCoin(denom string, amount int64) Coin](#newcoin)
  * [func (Coin) String() string](#coin-string)
  * [func (Coin) IsGTE(other Coin) bool](#coin-isgte)
  * [func (Coin) IsLT(other Coin) bool](#coin-islt)
  * [func (Coin) IsEqual(other Coin) bool](#coin-isequal)
  * [func (Coin) Add(other Coin) Coin](#coin-add)
  * [func (Coin) Sub(other Coin) Coin](#coin-sub)
  * [func (Coin) IsPositive() bool](#coin-ispositive)
  * [func (Coin) IsNegative() bool](#coin-isnegative)
  * [func (Coin) IsZero() bool](#coin-iszero)
* [type Coins](#coins)

  * [func NewCoins(cs ...Coin) Coins](#newcoins)
  * [func (Coins) String() string](#coins-string)
  * [func (Coins) AmountOf(denom string) int64](#coins-amountof)
  * [func (Coins) Add(other Coins) Coins](#coins-add)
* [type Realm](#realm)

  * [func (Realm) Address() Address](#realm-address)
  * [func (Realm) PkgPath() string](#realm-pkgpath)
  * [func (Realm) IsUser() bool](#realm-isuser)
  * [func (Realm) IsUserRun() bool](#realm-isuserrun)
  * [func (Realm) IsUserCall() bool](#realm-isusercall)
  * [func (Realm) CoinDenom(coinName string) string](#realm-coindenom)

---

### func ChainDomain() string {#chaindomain}

ChainDomain returns the chain domain (e.g., "gno.land").

### func ChainID() string {#chainid}

ChainID returns the current chain identifier.

### func ChainHeight() int64 {#chainheight}

ChainHeight returns the current block height.

### func Emit(typ string, attrs ...string) {#emit}

Emit records an event of type typ with alternating key, value attribute pairs.
Attributes must be supplied as an even-length list.

### func CoinDenom(pkgPath, coinName string) string {#coindenom}

CoinDenom composes a qualified denomination using pkgPath and coinName.
coinName must begin with a lowercase letter and contain 3–16 lowercase letters
or digits.

---

### type Address {#address}

```
type Address string
```

#### func (Address) IsValid() bool {#address-isvalid}

IsValid reports whether the address is well-formed (bech32).

#### func (Address) String() string {#address-string}

String returns the string form of the address.

---

### type Coin {#coin}

```
type Coin struct {
    Denom  string `json:"denom"`
    Amount int64  `json:"amount"`
}
```

#### func NewCoin(denom string, amount int64) Coin {#newcoin}

NewCoin returns a Coin of denom and amount.

#### Methods

* String formats the Coin as "<amount><denom>".
* IsGTE, IsLT, IsEqual compare amounts; the denoms must match.
* Add, Sub require matching denoms; they panic on overflow/underflow.
* IsPositive, IsNegative, IsZero report the sign of Amount.

---

### type Coins {#coins}

```
type Coins []Coin
```

Coins represents a set of Coin, at most one per denomination.

#### func NewCoins(cs ...Coin) Coins {#newcoins}

NewCoins constructs a set from cs, consolidating duplicate denoms.

#### Methods

* String formats the set as a comma-separated list of coins.
* AmountOf returns the amount for denom, or 0 if absent.
* Add merges other into the receiver, summing matching denoms.

---

### type Realm {#realm}

```
type Realm struct {
    addr    Address
    pkgPath string
}
```

#### Methods

* Address returns the realm address.
* PkgPath returns the package path.
* IsUser reports whether the realm is a user realm.
* IsUserRun reports whether the call originated from MsgRun.
* IsUserCall reports whether the call originated from MsgCall.
* CoinDenom qualifies a coin name using the realm's package path.

---

## package banker

Package **banker** provides controlled access to balances for native coins.

```
import "gno.land/.../chain/banker"
```

### Index

* [type Type](#banker-type)
* [consts](#banker-consts)
* [type Banker](#banker-interface)
* [func New(t Type) Banker](#banker-new)

### type Type {#banker-type}

```
type Type uint8
```

### Constants {#banker-consts}

```
const (
    TypeReadonly Type = iota   // read-only balances
    TypeOriginSend             // access to coins sent with the tx
    TypeRealmSend              // access to realm-owned coins (incl. sent)
    TypeRealmIssue             // may issue new coins
)
```

### type Banker {#banker-interface}

```
type Banker interface {
    GetCoins(addr chain.Address) (dst chain.Coins)
    SendCoins(from, to chain.Address, coins chain.Coins)
    IssueCoin(addr chain.Address, denom string, amount int64)
    RemoveCoin(addr chain.Address, denom string, amount int64)
}
```

### func New(t Type) Banker {#banker-new}

New returns a Banker with the capabilities implied by t.

---

## package runtime

Package **runtime** exposes call-context information and realm stack utilities.

```
import "gno.land/.../chain/runtime"
```

### Index

* [func AssertOriginCall()](#runtime-assertorigincall)
* [func OriginSend() chain.Coins](#runtime-originsend)
* [func OriginCaller() chain.Address](#runtime-origincaller)
* [func CurrentRealm() chain.Realm](#runtime-currentrealm)
* [func PreviousRealm() chain.Realm](#runtime-previousrealm)
* [func CallerAt(n int) chain.Address](#runtime-callerat)

### func AssertOriginCall() {#runtime-assertorigincall}

AssertOriginCall panics unless the caller is an EOA (MsgCall). MsgRun callers
are rejected.

### func OriginSend() chain.Coins {#runtime-originsend}

OriginSend returns the Coins attached to the current transaction.

### func OriginCaller() chain.Address {#runtime-origincaller}

OriginCaller returns the original signer of the transaction.

### func CurrentRealm() chain.Realm {#runtime-currentrealm}

CurrentRealm returns the current realm.

### func PreviousRealm() chain.Realm {#runtime-previousrealm}

PreviousRealm returns the previous caller realm; for user callers, pkgPath may
be empty.

### func CallerAt(n int) chain.Address {#runtime-callerat}

CallerAt returns the n-th caller in the stack trace (n > 0).

---

## package testing

Package **testing** provides helpers for deterministic testing of chain logic.

```
import "gno.land/.../testing" // path subject to final layout
```

### Index

* [func SkipHeights(count int64)](#testing-skipheights)
* [func SetOriginCaller(orig chain.Address)](#testing-setorigincaller)
* [func SetOriginSend(sent chain.Coins)](#testing-setoriginsend)
* [func IssueCoins(addr chain.Address, coins chain.Coins)](#testing-issuecoins)
* [func SetRealm(realm chain.Realm)](#testing-setrealm)
* [func NewUserRealm(address chain.Address) chain.Realm](#testing-newuserrealm)
* [func NewCodeRealm(pkgPath string) chain.Realm](#testing-newcoderealm)

### func SkipHeights(count int64) {#testing-skipheights}

SkipHeights advances the block height by count and the timestamp by 5 seconds
per block.

### func SetOriginCaller(orig chain.Address) {#testing-setorigincaller}

SetOriginCaller sets the current origin caller.

### func SetOriginSend(sent chain.Coins) {#testing-setoriginsend}

SetOriginSend sets sent/spent coins for the current context.

### func IssueCoins(addr chain.Address, coins chain.Coins) {#testing-issuecoins}

IssueCoins credits coins to addr in the testing context.

### func SetRealm(realm chain.Realm) {#testing-setrealm}

SetRealm sets the realm for the current frame. Subsequent calls to
runtime.CurrentRealm in the same test return realm.

### func NewUserRealm(address chain.Address) chain.Realm {#testing-newuserrealm}

NewUserRealm constructs a user realm for tests.

### func NewCodeRealm(pkgPath string) chain.Realm {#testing-newcoderealm}

NewCodeRealm constructs a code realm for tests.
