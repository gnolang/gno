---
id: tm2-js-ledger
---

# Ledger Signer

Ledger device-based signer instance

### new LedgerSigner

Creates a new instance of the Ledger device signer, using the provided `LedgerConnector`

#### Parameters

* `connector` **LedgerConnector** the Ledger connector
* `accountIndex` **number** the desired account index
* `addressPrefix` **string** the address prefix

#### Usage

```ts
const accountIndex: number = 10 // for ex. 10th account in the derivation
const connector: LedgerConnector = // ...

new LedgerSigner(connector, accountIndex);
// new Ledger device signer created
```
