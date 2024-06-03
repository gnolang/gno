---
id: testing
---

# Testing

```go
func TestCurrentRealm() string
func TestSkipHeights(count int64)
func TestSetOrigCaller(addr Address)
func TestSetOrigPkgAddr(addr Address)
func TestSetOrigSend(sent, spent Coins)
func TestIssueCoins(addr Address, coins Coins)
```

## TestCurrentRealm
```go
func TestCurrentRealm() string
```
Returns the current realm path.

#### Usage
```go
currentRealmPath := std.TestCurrentRealm()
```
---

## TestSkipHeights
```go
func TestSkipHeights(count int64)
```
Modifies the block height variable by skipping **count** blocks.

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
std.TestSetOrigCaller("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
```
---

## TestSetOrigPkgAddr
```go
func TestSetOrigPkgAddr(addr Address)
```
Sets the current realm/package address to **addr**.

#### Usage
```go
std.TestSetOrigPkgAddr("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec")
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
addr := "g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec"
std.TestIssueCoins(addr, issue)
```





