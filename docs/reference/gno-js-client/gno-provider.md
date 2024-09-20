---
id: gno-js-provider
---

# Gno Provider

The `Gno Provider` is an extension on the `tm2-js-client` `Provider`,
outlined [here](../tm2-js-client/Provider/provider.md). Both JSON-RPC and WS providers are included with the package.

## Realm Methods

### getRenderOutput

Executes the Render(path) method in read-only mode

#### Parameters

* `packagePath` **string** the gno package path
* `path` **string** the render path
* `height` **number** the height for querying.
  If omitted, the latest height is used (optional, default `0`)

Returns **Promise<string\>**

#### Usage

```ts
await provider.getRenderOutput('gno.land/r/demo/demo_realm', '');
// ## Hello World!
```

### getFunctionSignatures

Fetches public facing function signatures

#### Parameters

* `packagePath` **string** the gno package path
* `height` **number** the height for querying.
  If omitted, the latest height is used (optional, default `0`)

Returns **Promise<FunctionSignature[]\>**

#### Usage

```ts
await provider.getFunctionSignatures('gno.land/r/demo/foo20');
/*
[
  { FuncName: 'TotalSupply', Params: null, Results: [ [Object] ] },
  {
    FuncName: 'BalanceOf',
    Params: [ [Object] ],
    Results: [ [Object] ]
  },
  {
    FuncName: 'Allowance',
    Params: [ [Object], [Object] ],
    Results: [ [Object] ]
  },
  {
    FuncName: 'Transfer',
    Params: [ [Object], [Object] ],
    Results: null
  },
  {
    FuncName: 'Approve',
    Params: [ [Object], [Object] ],
    Results: null
  },
  {
    FuncName: 'TransferFrom',
    Params: [ [Object], [Object], [Object] ],
    Results: null
  },
  { FuncName: 'Faucet', Params: null, Results: null },
  { FuncName: 'Mint', Params: [ [Object], [Object] ], Results: null },
  { FuncName: 'Burn', Params: [ [Object], [Object] ], Results: null },
  { FuncName: 'Render', Params: [ [Object] ], Results: [ [Object] ] }
]
 */
```

### evaluateExpression

Evaluates any expression in readonly mode and returns the results

#### Parameters

* `packagePath` **string** the gno package path
* `expression` **string** the expression to be evaluated
* `height` **number** the height for querying.
  If omitted, the latest height is used (optional, default `0`)

Returns **Promise<string\>**

#### Usage

```ts
await provider.evaluateExpression('gno.land/r/demo/foo20', 'TotalSupply()')
// (10100000000 uint64)
```

### getFileContent

Fetches the file content, or the list of files if the path is a directory

#### Parameters

* `packagePath` **string** the gno package path
* `height` **number** the height for querying.
  If omitted, the latest height is used (optional, default `0`)

Returns **Promise<string\>**

#### Usage

```ts
await provider.getFileContent('gno.land/r/demo/foo20')
/*
foo20.gno
foo20_test.gno
 */
```
