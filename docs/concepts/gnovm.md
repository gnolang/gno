---
id: gnovm
---

# GnoVM

GnoVM is a virtual machine that interprets Gno, a custom version of Go optimized for blockchains, featuring automatic state management, full determinism, and idiomatic Go.
It works with Tendermint2 and enables smarter, more modular, and transparent appchains with embedded smart-contracts.
It can be adapted for use in TendermintCore, forks, and non-Cosmos blockchains.

Read the ["Intro to Gnoland"](https://test3.gno.land/r/gnoland/blog:p/intro) blog post for more information.

This folder focuses on the VM, language, stdlibs, tests, and tools, independent of the blockchain.
This enables non-web3 developers to contribute without requiring an understanding of the broader context.

## Language Features

* Like interpreted Go, but more ambitious.
* Completely deterministic, for complete accountability.
* Transactional persistence across data realms.
* Designed for concurrent blockchain smart contracts systems.

## Getting started

Install [`gno`](../getting-started/local-setup/local-setup.md) and refer to the [`examples`](https://github.com/gnolang/gno/tree/master/examples) folder to start developing contracts.

## Enhance or modify gnoVM

To enhance or modify gnoVM, Gno, and stdlibs, familiarize yourself with the [Makefile](https://github.com/gnolang/gno/blob/master/gnovm/Makefile). 

The following are parts of the gnoVM Makefile: environment variables, dev tools, test suite, and code gen. See the sections below for more details on each part.

### Environment variables

Environment variables define the virtual machine settings. Refer to the [Makefile](https://github.com/gnolang/gno/blob/master/gnovm/Makefile) for reference. 

These variables can be overwritten in the command line by passing the variable name and value. For example, to update the test suite timeout from the default `30m` to `15m`, pass `GOTEST_FLAGS= -v -p 1 -timeout=15m`. 

Environment Variable | Definition | Default
-------------------- | ---------- | ---------
`CGO_ENABLED`        | Enable cgo to use any C code. Use may require additional dependencies, and is not strictly required by any tm2 code. See [Go's documentation on cgo with the go command](https://pkg.go.dev/cmd/cgo#hdr-Using_cgo_with_the_go_command). | cgo is disabled by default. `CGO_ENABLED` must be exported; use `export CGO_ENABLED`.
`GOFMT_FLAGS`       | Change formatting of Go output. `-w` flag writes the output to the destination flag. See [Go's documentation on gofmt flags](https://pkg.go.dev/cmd/gofmt). | Uses the standard Go formatting by default.
`GNOFMT_FLAGS`      | Change formatting of Gno output. `-w` flag writes the output to the destination flag. | Uses the standard Gno formatting by default.
`GOIMPORTS_FLAGS`   | Adds Go formatting flags for `make imports`. | Uses values defined in `GOFMT_FLAGS` by default.
`GOTEST_FLAGS`      | Modifies Go test suites through flags. | Prints verbose output (`-v`), defines the number of programs that can be run in parallel (`-p 1`), and times out after 30 minutes (`-timeout=30m`). 
`GNOROOT_DIR`       | Sets the default GNOROOT. | In local development, sets GNOROOT to the supposed path of the Gno repository clone.
`GOBUILD_FLAGS`     | Sets Go build flags in an argument list. See [Go's build flags documentation](https://pkg.go.dev/cmd/go). | Sets the arguments to pass on each Go tool link invocation to the Gno project.
`GOTEST_COVER_PROFILE` | Defines the file where to place the cover profile. | The output location is set to `cmd-profile.out` by default/

### Dev tools

The dev tools section defines aliases for Go commands. See the [Makefile](https://github.com/gnolang/gno/blob/master/gnovm/Makefile) for comprehensive Go commands corresponding to the aliases.

Alias command | Definition
------------- | ----------
`build`       | Builds Gno with `GNOBUILD_FLAGS`.
`install`     | Installs Gno with `GNOBUILD_FLAGS`.
`clean`       | Removes and refreshes build.
`lint`        | Lints the golang configuration.
`fmt`         | Formats using `GNOFMT_FLAGS`.
`imports`     | Imports using `GOIMPORT_FLAGS`.

### Test suite

The test suite section defines aliases for Go commands related to tests. See the [Makefile](https://github.com/gnolang/gno/blob/master/gnovm/Makefile) for Go commands corresponding to the aliases.

Alias command | Definition 
------------- | -----------
`test`        | Executes `_test.cmd`, `_test.pkg`, and `_test.gnolang`. 
`_test.cmd`   | Executes tests at file location `./cmd/...` with `GOTEST_FLAGS`.
`_test.pkg`   | Executes tests at file location `./pkg/...` with `GOTEST_FLAGS`.
`_test.gnolang` | Executes Gnolang tests, including `native`, `stdlibs`, and `realm`. See [Makefile](https://github.com/gnolang/gno/blob/master/gnovm/Makefile) for comprehensive list.
`test.cmd.coverage` | Runs tests on `./cmd/` and saves the output to `GOTEST_COVER_PROFILE`. 
`test.cmd.coverage_view` | Runs `test.cmd.coverage`, then shows the results rendered in the HTML browser, and deletes the original file.

### Code gen

Alias command | Definition
------------- | ----------
`generate`    | Executes `_dev.stringer` and `_dev_generate`.
`_dev.stringer` | Executes [golang stringer](golang.org/x/tools/cmd/stringer).
`_dev.generate` | Executes `go generate` with any added flags.
