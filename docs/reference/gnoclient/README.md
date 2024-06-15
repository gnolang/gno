---
id: gnoclient
---

# Gnoclient

[Gnoclient](https://github.com/gnolang/gno/tree/master/gno.land/pkg/gnoclient) 
allows you to easily access Gno blockchains from your Go code, through exposing 
APIs for common functionality.

## Key Features
                
- Connect to a Gno chain via RPC
- Use local keystore to sign & broadcast transactions containing any type of 
Gno message
- Sign & broadcast transactions with batch messages
- Use [ABCI queries](../../gno-tooling/cli/gnokey.md#make-an-abci-query) in
your Go code

## Installation

To add Gnoclient to your Go project, run the following command:
```bash
go get github.com/gnolang/gno/gno.land/pkg/gnoclient
```