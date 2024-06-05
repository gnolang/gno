---
id: tm2-js-wallet
---

# Wallet

A `Wallet` is a user-facing API that is used to interact with an account. A `Wallet` instance is tied to a single key
pair and essentially wraps the given `Provider` for that specific account.

A wallet can be generated from a randomly generated seed, a private key, or instantiated using a Ledger device.

Using the `Wallet`, users can easily interact with the Tendermint2 chain using their account without having to worry
about account management.

## Initialization

### createRandom

Generates a private key-based wallet, using a random seed

#### Parameters

* `options?` **AccountWalletOption** the account options

Returns **Promise<Wallet\>**

#### Usage

```ts
const wallet: Wallet = await Wallet.createRandom();
// random wallet created
```

### fromMnemonic

Generates a bip39 mnemonic-based wallet

#### Parameters

* `mnemonic` **string** the bip39 mnemonic
* `options?` **CreateWalletOptions** the wallet generation options

Returns **Promise<Wallet\>**

#### Usage

```ts
const mnemonic: string = // ...
const wallet: Wallet = await Wallet.fromMnemonic(mnemonic);
// wallet created from mnemonic
```

### fromPrivateKey

Generates a private key-based wallet

#### Parameters

* `privateKey` **string** the private key
* `options?` **AccountWalletOption** the wallet generation options

Returns **Promise<Wallet\>**

#### Usage 

```ts
// Generate the private key from somewhere
const {publicKey, privateKey} = await generateKeyPair(
    entropyToMnemonic(generateEntropy()),
    index ? index : 0
);

const wallet: Wallet = await Wallet.fromPrivateKey(privateKey);
// wallet created from private key
```

### fromLedger

Creates a Ledger-based wallet

#### Parameters

* `connector` **LedgerConnector** the Ledger device connector
* `options?` **CreateWalletOptions** the wallet generation options

Returns **Wallet**

#### Usage

```ts
const connector: LedgerConnector = // ...

const wallet: Wallet = await Wallet.fromLedger(connector);
// wallet created from Ledger device connection
```

## Provider Methods

### connect

Connects the wallet to the specified Provider

#### Parameters

* `provider` **Provider** the active Provider, if any

#### Usage

```ts
const provider: Provider = // ...

    wallet.connect(provider);
// Provider connected to Wallet
```

### getProvider

Returns the connected provider, if any

Returns **Provider**

#### Usage

```ts
wallet.getProvider();
// connected provider, if any (undefined if not)
```

## Account Methods

### getAddress

Fetches the address associated with the wallet

Returns **Promise<string\>**

#### Usage

```ts
await wallet.getAddress();
// "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"
```

### getSequence

Fetches the account sequence for the wallet

#### Parameters

* `height` **number** the block height (optional, default `latest`)

#### Usage

```ts
await wallet.getSequence();
// 42
```

Returns **Promise<number\>**

### getAccountNumber

Fetches the account number for the wallet. Errors out if the
account is not initialized

#### Parameters

* `height` **number** the block height (optional, default `latest`)

Returns **Promise<number\>**

#### Usage

```ts
await wallet.getAccountNumber();
// 10
```

### getBalance

Fetches the account balance for the specific denomination

#### Parameters

* `denomination` **string** the fund denomination (optional, default `ugnot`)

Returns **Promise<number\>**

#### Usage

```ts
await wallet.getBalance('ugnot');
// 5000
```

### getGasPrice

Fetches the current (recommended) average gas price

Returns **Promise<number\>**

#### Usage

```ts
await wallet.getGasPrice();
// 63000
```

### estimateGas

Estimates the gas limit for the transaction

#### Parameters

* `tx` **Tx** the transaction that needs estimating

Returns **Promise<number\>**

#### Usage

```ts
const tx: Tx = // ...

    await wallet.estimateGas(tx);
// 120000
```

### signTransaction

Generates a transaction signature, and appends it to the transaction

#### Parameters

* `tx` **Tx** the transaction to be signed

Returns **Promise<Tx\>**

#### Usage

```ts
const tx: Tx = // ...

    await wallet.signTransaction(tx);
// transaction with appended signature
```

### sendTransaction

Signs and sends the transaction. Returns the transaction hash (base-64)

#### Parameters

* `tx` **Tx** the unsigned transaction

Returns **Promise<string\>**

#### Usage

```ts
await wallet.sendTransaction(tx);
// returns the transaction hash
```

### getSigner

Returns the associated signer

Returns **Signer**

#### Usage

```ts
wallet.getSigner(tx);
// Signer instance
```
