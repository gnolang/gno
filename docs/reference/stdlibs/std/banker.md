---
id: banker
---

# Banker
View concept page [here](../../../concepts/stdlibs/banker.md).

```go
type BankerType uint8

const (
    BankerTypeReadonly BankerType = iota
    BankerTypeOrigSend
    BankerTypeRealmSend
    BankerTypeRealmIssue
)

type Banker interface {
    GetCoins(addr Address) (dst Coins)
    SendCoins(from, to Address, coins Coins)
    IssueCoin(addr Address, denom string, amount int64)
    RemoveCoin(addr Address, denom string, amount int64)
}
```

## GetBanker
Returns `Banker` of the specified type.

#### Parameters
- `BankerType` - type of Banker to get:
    - `BankerTypeReadonly` - read-only access to coin balances
    - `BankerTypeOrigSend` - full access to coins sent with the transaction that calls the banker
    - `BankerTypeRealmSend` - full access to coins that the realm itself owns, including the ones sent with the transaction
    - `BankerTypeRealmIssue` - able to issue new coins

#### Usage

```go
banker := std.GetBanker(std.<BankerType>)
```
---

## GetCoins
Returns `Coins` owned by `Address`.

#### Parameters
- `addr` **Address** to fetch balances for

#### Usage

```go
coins := banker.GetCoins(addr)
```
---

## SendCoins
Sends `coins` from address `from` to address `to`. `coins` needs to be a well-defined
`Coins` slice.

#### Parameters
- `from` **Address** to send from
- `to` **Address** to send to
- `coins` **Coins** to send

#### Usage
```go
banker.SendCoins(from, to, coins)
```
---

## IssueCoin
Issues `amount` of coin with a denomination `denom` to address `addr`.

#### Parameters
- `addr` **Address** to issue coins to
- `denom` **string** denomination of coin to issue
- `amount` **int64** amount of coin to issue

#### Usage
```go
banker.IssueCoin(addr, denom, amount)
```
---

## RemoveCoin
Removes (burns) `amount` of coin with a denomination `denom` from address `addr`.

#### Parameters
- `addr` **Address** to remove coins from
- `denom` **string** denomination of coin to remove
- `amount` **int64** amount of coin to remove

#### Usage
```go
banker.RemoveCoin(addr, denom, amount)
```
