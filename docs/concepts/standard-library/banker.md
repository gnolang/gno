---
id: banker
---

# Banker

The Banker's main purpose is to handle balance changes of native coins (link native coin) within Gno chains. This includes issuance, transfers, and burning of [Coins](coins.md). 

The Banker module can be cast into 4 subtypes of bankers that expose different functionalities and safety features within your packages and realms.

[//]: # (The banker module is injected into the GnoVM runtime at execution. )

### Banker Types

1. `BankerTypeReadOnly` - read-only access to coin balances
2. `BankerTypeOrigSend` - full access to coins sent with the transaction that calls the banker
3. `BankerTypeRealmSend` - full access to coins that the realm itself owns, including the ones sent with the transaciton
4. `BankerTypeRealmIssue` - able to issue new coins
 
You can access the Banker from within the `std` namespace by calling `std.GetBanker(<BankerType>)` in your package/realm.

The Banker API can be found in [Banker Reference].





