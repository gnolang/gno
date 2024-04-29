---
id: tm2-js-utility
---

# Utility Helpers

## Provider Helpers

### extractBalanceFromResponse

Extracts the specific balance denomination from the ABCI response

#### Parameters

* `abciData` **(string | null)** the base64-encoded ABCI data
* `denomination` **string** the required denomination

### extractSequenceFromResponse

Extracts the account sequence from the ABCI response

#### Parameters

* `abciData` **(string | null)** the base64-encoded ABCI data

Returns **number**

### extractAccountNumberFromResponse

Extracts the account number from the ABCI response

#### Parameters

* `abciData` **(string | null)** the base64-encoded ABCI data

Returns **number**

### waitForTransaction

Waits for the transaction to be committed to a block in the chain
of the specified provider. This helper does a search for incoming blocks
and checks if a transaction

#### Parameters

* `provider` **Provider** the provider instance
* `hash` **string** the base64-encoded hash of the transaction
* `fromHeight` **number** the starting height for the search. If omitted, it is the latest block in the chain (
  optional, default `latest`)
* `timeout` **number** the timeout in MS for the search (optional, default `15000`)

Returns **Promise<Tx\>**

## Request Helpers

### newRequest

Creates a new JSON-RPC 2.0 request

#### Parameters

* `method` **string** the requested method
* `params` **Array<string\>?** the requested params, if any

Returns **RPCRequest**

### newResponse

Creates a new JSON-RPC 2.0 response

#### Parameters

* `result` **Result** the response result, if any
* `error` **RPCError** the response error, if any

Returns **RPCResponse<Result\>**

### parseABCI

Parses the base64 encoded ABCI JSON into a concrete type

#### Parameters

* `data` **string** the base64-encoded JSON

Returns **Result**

### stringToBase64

Converts a string into base64 representation

#### Parameters

* `str` **string** the raw string

Returns **string**

### base64ToUint8Array

Converts a base64 string into a Uint8Array representation

#### Parameters

* `str` **string** the base64-encoded string

Returns **Uint8Array**

### uint8ArrayToBase64

Converts a Uint8Array into base64 representation

#### Parameters

* `data` **Uint8Array** the Uint8Array to be encoded

Returns **string**
