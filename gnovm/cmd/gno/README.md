# gno

`gno` (formerly `gnodev`) is a tool for managing Gno source code.

## Usage

`gno <command> [arguments]`

## Usage

[embedmd]:#(../../.tmp/gno-help.txt)
```txt
USAGE
  gno <command> [arguments]

SUBCOMMANDS
  bug      start a bug report
  clean    remove generated and cached data
  doc      show documentation for package or symbol
  env      print gno environment information
  fix      update and fix old gno source files
  fmt      gnofmt (reformat) package sources
  list     lists the named packages
  lint     runs the linter for the specified packages
  mod      module maintenance
  repl     starts a GnoVM REPL
  run      run gno packages
  test     test packages
  tool     run specified gno tool
  version  display installed gno version

FLAGS
  -C ...  change to directory before running command

```

## Install

    go install github.com/gnolang/gno/gnovm/cmd/gno

Or

    > git clone git@github.com:gnolang/gno.git
    > cd ./gno
    > make install.gno

## Getting started

TODO
