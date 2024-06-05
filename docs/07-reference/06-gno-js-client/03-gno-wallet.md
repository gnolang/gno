---
id: gno-js-wallet
---

# Gno Wallet

The `Gno Wallet` is an extension on the `tm2-js-client` `Wallet`, outlined [here](../05-tm2-js-client/02-wallet.md).

## Account Methods

### transferFunds

Initiates a native currency transfer transaction between accounts

#### Parameters

* `to` **string** the bech32 address of the receiver
* `funds` **Map<string, number\>** the denomination -> value map for funds
* `fee` **TxFee** the custom transaction fee, if any

Returns **Promise<string\>**

#### Usage

```ts
let fundsMap = new Map<string, number>([
    ["ugnot", 10],
]);

await wallet.transferFunds('g1flk9z2qmkgqeyrl654r3639rzgz7xczdfwwqw7', fundsMap);
// returns the transaction hash
```

### callMethod

Invokes the specified method on a GNO contract

#### Parameters

* `path` **string** the gno package / realm path
* `method` **string** the method name
* `args` **string[]** the method arguments, if any
* `funds` **Map<string, number\>** the denomination -> value map for funds
* `fee` **TxFee** the custom transaction fee, if any

Returns **Promise<string\>**

#### Usage

```ts
let fundsMap = new Map<string, number>([
    ["ugnot", 10],
]);

await wallet.callMethod('gno.land/r/demo/foo20', 'TotalBalance', []);
// returns the transaction hash
```

### deployPackage

Deploys the specified package / realm

#### Parameters

* `gnoPackage` **MemPackage** the package / realm to be deployed
* `funds` **Map<string, number>** the denomination -> value map for funds
* `fee` **TxFee** the custom transaction fee, if any

Returns **Promise<string\>**

#### Usage

```ts
const memPackage: MemPackage = // ...

    await wallet.deployPackage(memPackage);
// returns the transaction hash
```
