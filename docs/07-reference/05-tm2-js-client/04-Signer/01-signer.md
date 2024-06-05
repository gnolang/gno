---
id: tm2-js-signer
---

# Overview

A `Signer` is an interface that abstracts the interaction with a single Secp256k1 key pair. It exposes methods for
signing data, verifying signatures, and getting metadata associated with the key pair, such as the address.

Currently, the `tm2-js-client` package provides support for two `Signer` implementations:

- [Key](02-key.md): a signer that is based on a raw Secp256k1 key pair.
- [Ledger](03-ledger.md): a signer that is based on a Ledger device, with all interaction flowing through the user's
  device.

## API

### getAddress

Returns the address associated with the signer's public key

Returns **Promise<string\>**

#### Usage

```ts
await signer.getAddress();
// "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"
```

### getPublicKey

Returns the signer's Secp256k1-compressed public key

Returns **Promise<Uint8Array\>**

#### Usage

```ts
await signer.getPublicKey();
// <Uint8Array>
```

### getPrivateKey

Returns the signer's actual raw private key

Returns **Promise<Uint8Array\>**

#### Usage

```ts
await signer.getPrivateKey();
// <Uint8Array>
```

### signData

Generates a data signature for arbitrary input

#### Parameters

* `data` **Uint8Array** the data to be signed

Returns **Promise<Uint8Array\>**

#### Usage

```ts
const dataToSign: Uint8Array = // ...

    await signer.signData(dataToSign);
// <Uint8Array>
```

### verifySignature

Verifies if the signature matches the provided raw data

#### Parameters

* `data` **Uint8Array** the raw data (not-hashed)
* `signature` **Uint8Array** the hashed-data signature

Returns **Promise<boolean\>**

#### Usage

```ts
const signedData: Uint8Array = // ...
const rawData: Uint8Array = // ...

    await signer.verifySignature(rawData, signedData);
// <Uint8Array>
```

