# gno

`gno` (formerly `gnodev`) is a tool for managing Gno source code.

## Usage

`gno <command> [arguments]`

## Usage

[embedmd]:#(../../.tmp/gno-help.txt)
```txt
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:345 : || +o OpSliceType
DEBUG:   op_eval.go:348 : ||| +o OpEval
DEBUG:   op_eval.go:350 : ||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:363 : ||| +o OpEval
DEBUG:   op_eval.go:368 : |||| +o OpEval
DEBUG:   op_eval.go:368 : ||||| +o OpEval
DEBUG:   op_eval.go:325 : ||||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||||| +o OpEval
DEBUG:   op_eval.go:350 : |||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : |||| +o OpFieldType
DEBUG:   op_eval.go:328 : ||||| +o OpEval
DEBUG:   op_eval.go:345 : ||||| +o OpSliceType
DEBUG:   op_eval.go:348 : |||||| +o OpEval
DEBUG:   op_eval.go:350 : |||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   op_eval.go:350 : |||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:813 :  +o OpHalt
DEBUG:   machine.go:814 : | +o OpPopBlock
DEBUG:   machine.go:816 : || +o OpStaticTypeOf
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:363 : ||| +o OpEval
DEBUG:   op_eval.go:368 : |||| +o OpEval
DEBUG:   op_eval.go:325 : |||| +o OpFieldType
DEBUG:   op_eval.go:328 : ||||| +o OpEval
DEBUG:   op_eval.go:350 : ||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:813 :  +o OpHalt
DEBUG:   machine.go:814 : | +o OpPopBlock
DEBUG:   machine.go:816 : || +o OpStaticTypeOf
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:363 : ||| +o OpEval
DEBUG:   op_eval.go:368 : |||| +o OpEval
DEBUG:   op_eval.go:368 : ||||| +o OpEval
DEBUG:   op_eval.go:325 : ||||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||||| +o OpEval
DEBUG:   op_eval.go:350 : |||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : |||| +o OpFieldType
DEBUG:   op_eval.go:328 : ||||| +o OpEval
DEBUG:   op_eval.go:350 : ||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:372 : || +o OpMapType
DEBUG:   op_eval.go:375 : ||| +o OpEval
DEBUG:   op_eval.go:378 : |||| +o OpEval
DEBUG:   op_eval.go:350 : |||| +o OpInterfaceType
DEBUG:   op_eval.go:350 : ||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:368 : ||| +o OpEval
DEBUG:   op_eval.go:368 : |||| +o OpEval
DEBUG:   op_eval.go:325 : |||| +o OpFieldType
DEBUG:   op_eval.go:328 : ||||| +o OpEval
DEBUG:   op_eval.go:372 : ||||| +o OpMapType
DEBUG:   op_eval.go:375 : |||||| +o OpEval
DEBUG:   op_eval.go:378 : ||||||| +o OpEval
DEBUG:   op_eval.go:350 : ||||||| +o OpInterfaceType
DEBUG:   op_eval.go:350 : |||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   op_eval.go:350 : |||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:813 :  +o OpHalt
DEBUG:   machine.go:814 : | +o OpPopBlock
DEBUG:   machine.go:816 : || +o OpStaticTypeOf
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:363 : ||| +o OpEval
DEBUG:   op_eval.go:368 : |||| +o OpEval
DEBUG:   op_eval.go:325 : |||| +o OpFieldType
DEBUG:   op_eval.go:328 : ||||| +o OpEval
DEBUG:   op_eval.go:350 : ||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:345 : || +o OpSliceType
DEBUG:   op_eval.go:348 : ||| +o OpEval
DEBUG:   op_eval.go:350 : ||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:363 : ||| +o OpEval
DEBUG:   op_eval.go:368 : |||| +o OpEval
DEBUG:   op_eval.go:368 : ||||| +o OpEval
DEBUG:   op_eval.go:325 : ||||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||||| +o OpEval
DEBUG:   op_eval.go:350 : |||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : |||| +o OpFieldType
DEBUG:   op_eval.go:328 : ||||| +o OpEval
DEBUG:   op_eval.go:345 : ||||| +o OpSliceType
DEBUG:   op_eval.go:348 : |||||| +o OpEval
DEBUG:   op_eval.go:350 : |||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   op_eval.go:350 : |||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:363 : ||| +o OpEval
DEBUG:   op_eval.go:368 : |||| +o OpEval
DEBUG:   op_eval.go:325 : |||| +o OpFieldType
DEBUG:   op_eval.go:328 : ||||| +o OpEval
DEBUG:   op_eval.go:350 : ||||| +o OpInterfaceType
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   op_eval.go:350 : |||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:345 : || +o OpSliceType
DEBUG:   op_eval.go:348 : ||| +o OpEval
DEBUG:   op_eval.go:350 : ||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:368 : ||| +o OpEval
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   op_eval.go:345 : |||| +o OpSliceType
DEBUG:   op_eval.go:348 : ||||| +o OpEval
DEBUG:   op_eval.go:350 : ||||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:345 : || +o OpSliceType
DEBUG:   op_eval.go:348 : ||| +o OpEval
DEBUG:   op_eval.go:350 : ||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:368 : ||| +o OpEval
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   op_eval.go:345 : |||| +o OpSliceType
DEBUG:   op_eval.go:348 : ||||| +o OpEval
DEBUG:   op_eval.go:350 : ||||| +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:350 : || +o OpInterfaceType
DEBUG:   machine.go:784 :  +o OpHalt
DEBUG:   machine.go:785 : | +o OpPopBlock
DEBUG:   machine.go:787 : || +o OpEval
DEBUG:   op_eval.go:359 : || +o OpFuncType
DEBUG:   op_eval.go:363 : ||| +o OpEval
DEBUG:   op_eval.go:325 : ||| +o OpFieldType
DEBUG:   op_eval.go:328 : |||| +o OpEval
DEBUG:   op_eval.go:350 : |||| +o OpInterfaceType
USAGE
  gno <command> [arguments]

SUBCOMMANDS
  bug      start a bug report
  clean    remove generated and cached data
  doc      show documentation for package or symbol
  env      print gno environment information
  fmt      gnofmt (reformat) package sources
  mod      module maintenance
  run      run gno packages
  test     test packages
  tool     run specified gno tool
  version  display installed gno version

```

## Install

    go install github.com/gnolang/gno/gnovm/cmd/gno

Or

    > git clone git@github.com:gnolang/gno.git
    > cd ./gno
    > make install_gno

## Getting started

TODO
