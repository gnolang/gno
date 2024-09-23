
id: realm-abi
---

# Gno Realm ABI specification

The Gno Realm Application Binary Interface (ABI) specifies the interface between realms (Gno smart contracts) and clients interacting with the realms.

## Overview and concepts

Realms implement live programs on the blockchain, also called smart contracts.

Each realm exists as a package, identified by its package path, and containing the package source files. A realm has the following properties:
- Once deployed, the realm is autonomous and permanent: its code can not be modified, and it can not be stopped, but it may internally self-terminate (to be clarified).
- The realm state (i.e. the set of its global variables, public and private) is persistent (i.e. stored on the blockchain).
- The realm must execute in a deterministic way, so its execution on mutiple nodes always result in the same state and a consensus can be achieved.
- The realm exported types, functions and methods are usable by external clients in transactions.

Realms are expressed in the Gno language, similar to Go, and represent the following kind of objects:

- _Types_ which determine a set of values together with operations and methods specific to those values.
- _Values_ which are instances of types, used as variables, constants, function parameters or returned result.
- _Functions_ which implement logic to process parameters and return values of certain types. A function defined for a certain type is a _method_.

To interact with a realm, the client connects to a node server (to be clarified), following the [ABCI] protocol. All the interactions between the realm and clients take place in ABCI transactions, called `Tx`.

A client interaction with the realm may consist to:
- call a realm function with some parameters and getting results in response. The realm function may or not alter the realm state. The related ABCI methods are: [CheckTx] and [DeliverTx].
- query the state of a realm, i.e. get the content of its exported objects by their name. Queries do not alter the realm state. The related ABCI method is [Query].

The ABCI protocol uses a [protobuf] encoding, with the `Tx` field containing the transaction content set as `bytes`. This document constitutes the extension to ABCI, describing `Tx`.

## Transaction generic data structures

### Coin

| Name   | Type   | Description | Field Number |
|--------|--------|-------------|--------------|
| Denom  | string | Coin name   | 1            |
| Amount | int64  | amount      | 2            |
 
### Fee

| Name      | Type           | Description                       | Field Number |
|-----------|----------------|-----------------------------------|--------------|
| GasWanted | int64          | gas requested for the transaction | 1            |
| GasFee    | [Coin](#coin)  | gas payment fee                   | 2            |

### Signature

| Name      | Type  | Description       | Field Number |
|-----------|-------|-------------------|--------------|
| PubKey    | bytes | signer public key | 1            |
| Signature | bytes | signature         | 2            |
  
### Tx

| Name       | Type                             | Description                         | Field Number |
|------------|----------------------------------|-------------------------------------|--------------|
| Msgs       | repeated string                  | [Message](#tx-messages) requests    | 1            |
| Fee        | [Fee](#fee)                      | fee                                 | 2            |
| Signatures | repeated [Signature](#signature) | signatures                          | 3            |
| Memo       | string                           | description                         | 4            |

Usually, 2 fees are specified in `Tx`:
- the gas wanted: the gas requested for `Tx`,
- the gas fee: the gas payment fee.

## Tx messages

`Tx` `Msgs` being protobuf strings, are encoded using a JSON representation. We describe here
only messages related to realms.

### MsgAddPkg

Create and initialize a new package.

In JSON, contains a first field `"@type": "/vm.m_addpkg"`.

| Name    | Type                    | Description                    | Field Number |
|---------|-------------------------|--------------------------------|--------------|
| Creator | string                  | creator address name or Bech32 | 1            |
| Package | [Package](#package)     | package definition to load     | 2            |
| Deposit | repeated [Coin](#coin)  | deposit                        | 3            |

### MsgCall

Execute a Gno statement.

In JSON, contains a first field `"@type": "/vm.m_call"`.

| Name    | Type                   | Description                          | Field Number |
|---------|------------------------|--------------------------------------|--------------|
| Caller  | string                 | caller address name or Bech32        | 1            |
| Send    | repeated [Coin](#coin) | amount to pay                        | 2            |
| PkgPath | string                 | package path of the function to call | 3            |
| Func    | string                 | function name                        | 4            |
| Args    | repeated string        | function arguments                   | 5            |

### MsgRun

Load and execute a Gno program. The code resides in a `main` package with no exports (can not be called externally), it is only executed once.

In JSON, contains a first field `"@type": "/vm.m_run"`.

| Name    | Type                   | Description                                             | Field Number |
|---------|------------------------|---------------------------------------------------------|--------------|
| Caller  | string                 | caller address name or Bech32                           | 1            |
| Send    | repeated [Coin](#coin) | amount to pay                                           | 2            |
| Package | [Package](#package)    | package definition to load. package name must be `main` | 3            |

### Package

Definition of package content, as used in [MsgAddPkg](#msgaddpkg) and [MsgRun](#msgrun)

| Name | Type                         | Descritption                       | Field Number |
|------|------------------------------|------------------------------------|--------------|
| Name | string                       | package name declared by `package` | 1            |
| Path | string                       | import path                        | 2            |
| Files| repeated [MemFile](#memfile) | gno source files                   | 3            |

### MemFile

Definition of a gno package file, used in [Package](#package)

| Name | Type   | Descritption                     | Field Number |
|------|--------|----------------------------------|--------------|
| Name | string | base file name, ending in `.gno` | 1            |
| Body | string | file content (gno source)        | 2            |

## Realm Types and Values Encoding

Code reference: `gnovm/pkg/gnolang/gnolang.proto`

### TypedValue

In JSON, contains a first field `"@type": "/gno.TypedValue"`.

| Name | Type  | Description                       | Field Number |
|------|-------|-----------------------------------|--------------|
| T    | Any   | type                              | 1            |
| V    | Any   | value                             | 2            |
| N    | Bytes | numeric bytes (for integers only) | 3            |

### PointerValue

In JSON, contains a first field `"@type": "/gno.PointerValue"`.

| Name  | Type                      | Description                    | Field Number |
|-------|---------------------------|--------------------------------|--------------|
| TV    | [TypedValue](#typedvalue) | typed value                    | 1            |
| Base  | Any                       | Array or Struct base value     | 2            |
| Index | int64                     | List or fields or values index | 3            |
| Key   | [TypedValue](#typedvalue) | for maps (optional in JSON)    | 4            |

### ArrayValue

In JSON, contains a first field `"@type": "/gno.ArrayValue"`.

| Name       | Type                               | Description       | Field Number |
|------------|------------------------------------|-------------------|--------------|
| ObjectInfo | [ObjectInfo](#objectinfo)          | object info       | 1            |
| List       | repeated [TypedValue](#typedvalue) | elements of array | 2            |
| Data       | bytes                              | internal data     | 3            |

### SliceValue

In JSON, contains a first field `"@type": "/gno.SliceValue"`.

| Name   | Type  | Description                | Field Number |
|--------|-------|----------------------------|--------------|
| Base   | Any   | Array or Struct base value | 1            |
| Offset | int64 | offset                     | 2            |
| Length | int64 | length                     | 3            |
| MaxCap | int64 | maximum capacity           | 4            |

### StructValue

In JSON, contains a first field `"@type": "/gno.StructValue"`.

| Name       | Type                               | Description   | Field Number |
|------------|------------------------------------|---------------|--------------|
| ObjectInfo | [ObjectInfo](#objectinfo)          | object info   | 1            |
| Fields     | repeated [TypedValue](#typedvalue) | struct fields | 2            |

### FuncValue

In JSON, contains a first field `"@type": "/gno.StructValue"`.

| Name       | Type   | Description                         | Field Number |
|------------|--------|-------------------------------------|--------------|
| Type       | Any    | normally a [FuncType](#functype)    | 1            |
| IsMethod   | bool   | true if function is a method        | 2            |
| Source     | Any    | normally a [BlockNode](#blocknode)  | 3            |
| Name       | string | function name                       | 4            |
| Closure    | Any    | normally a [RefValue](#refvalue)    | 5            |
| FileName   | string | file name where function is defined | 6            |
| PkgPath    | string | package path                        | 7            |
| NativePkg  | string | native package path                 | 8            |
| NativeName | string | native file name                    | 9            |

### MapValue

### MapList

### MapListItem

### TypeValue

### TypeType

### RefType

### StructType

### DeclaredType

### InterfaceType

[ABCI]: https://github.com/tendermint/tendermint/tree/master/spec/abci
[CheckTx]: https://github.com/tendermint/spec/blob/master/spec/abci/abci.md#checktx-1
[DeliverTx]: https://github.com/tendermint/spec/blob/master/spec/abci/abci.md#delivertx-1
[Query]: https://github.com/tendermint/spec/blob/master/spec/abci/abci.md#query-1
[protobuf]: https://protobuf.dev
