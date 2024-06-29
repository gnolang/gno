---
id: testing
---

# Testing

```go
func TestSkipHeights(count int64)
func TestSetOrigCaller(addr Address)
func TestSetOrigPkgAddr(addr Address)
func TestSetOrigSend(sent, spent Coins)
func TestIssueCoins(addr Address, coins Coins)
func TestSetRealm(realm Realm)
func NewUserRealm(address Address) Realm
func NewCodeRealm(pkgPath string) Realm
```

---

## TestSkipHeights

```go
func TestSkipHeights(count int64)
```
Modifies the block height variable by skipping **count** blocks.

It also increases block timestamp by 5 seconds for every single count

#### Usage
```go
std.TestSkipHeights(100)
```
---

## TestSetOrigCaller

```go
func TestSetOrigCaller(addr Address)
```
Sets the current caller of the transaction to **addr**.

#### Usage
```go
std.TestSetOrigCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
```
---

## TestSetOrigPkgAddr

```go
func TestSetOrigPkgAddr(addr Address)
```
Sets the call entry realm address to **addr**.

#### Usage
```go
std.TestSetOrigPkgAddr(std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec"))
```

---

## TestSetOrigSend

```go
func TestSetOrigSend(sent, spent Coins)
```
Sets the sent & spent coins for the current context.

#### Usage
```go
std.TestSetOrigSend(sent, spent Coins)
```
---

## TestIssueCoins

```go
func TestIssueCoins(addr Address, coins Coins)
```

Issues testing context **coins** to **addr**.

#### Usage

```go
issue := std.Coins{{"coin1", 100}, {"coin2", 200}}
addr := std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec")
std.TestIssueCoins(addr, issue)
```

---

## TestSetRealm

```go
func TestSetRealm(rlm Realm)
```

Sets the realm for the current frame. After calling `TestSetRealm()`, calling 
[`CurrentRealm()`](chain.md#currentrealm) in the same test function will yield the value of `rlm`, and 
any `PrevRealm()` called from a function used after TestSetRealm will yield `rlm`.

Should be used in combination with [`NewUserRealm`](#newuserrealm) &
[`NewCodeRealm`](#newcoderealm).

#### Usage
```go
addr := std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec")
std.TestSetRealm(std.NewUserRealm(""))
// or 
std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
```

---

## NewUserRealm

```go
func NewUserRealm(address Address) Realm
```

Creates a new user realm for testing purposes.

#### Usage
```go
addr := std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec")
userRealm := std.NewUserRealm(addr)
```

---

## NewCodeRealm

```go
func NewCodeRealm(pkgPath string) Realm
```

Creates a new code realm for testing purposes.

#### Usage
```go
path := "gno.land/r/demo/boards"
codeRealm := std.NewCodeRealm(path)
```







