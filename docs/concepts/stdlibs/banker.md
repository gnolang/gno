---
id: banker
---

# Banker

The Banker's main purpose is to handle balance changes of [native coins](./coin.md) within Gno chains. This includes issuance, transfers, and burning of coins. 

The Banker module can be cast into 4 subtypes of bankers that expose different functionalities and safety features within your packages and realms.

### Banker Types

1. `BankerTypeReadonly` - read-only access to coin balances
2. `BankerTypeOrigSend` - full access to coins sent with the transaction that called the banker
3. `BankerTypeRealmSend` - full access to coins that the realm itself owns, including the ones sent with the transaction
4. `BankerTypeRealmIssue` - able to issue new coins

The Banker API can be found under the `std` package [reference](../../reference/stdlibs/std/banker.md).
