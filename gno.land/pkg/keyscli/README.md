## keycli

`keycli` is an extension of `tm2/keys/client`, enhancing its functionality. It provides the following features:

- **addpkg**: Allows you to upload a new package to the blockchain.
- **run**: Execute Gno code by invoking the main() function from the target package.
- **call**: Executes a single function call within a Realm.
- **maketx**: Compose a transaction (tx) document to sign (and possibly broadcast).

--- 

Most of these features have been extracted from `tm2/keys/client` to ensure that `tm2` remains completely independent of `gnovm` and `gno.land`. For more detailed information regarding this change, please refer to [PR#1483](https://github.com/gnolang/gno/pull/1483)
