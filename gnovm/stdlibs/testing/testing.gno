// Shim for Go's "testing" package to support minimal testing types.
package testing

import (
	"testing/base"
)

// NOTE almost everything actually declared in testing/base

// NOTE These native functions are duplicated in testing and testing/base,
// and also all overridden in test/stdlibs/testing and test/stdlibs/testing/base.
func matchString(pat, str string) (bool, string)
func recoverWithStacktrace() (interface{}, string)
func unixNano() int64

// ----------------------------------------
// Top level functions

// Recover functions like recover(), but it ensures that the recovered error is
// not an internal error of the testing package.
//
// Due to a lack of goroutines and thus runtime.Goexit, gno's testing system resorts
// to panics to abort testing with FailNow (and Fatal* functions) or SkipNow
// (and Skip* functions).
//
// NOTE: Recover() is likely to be removed.
var Recover = base.Recover

type Setter = base.Setter

var Short = base.Short

var Verbose = base.Verbose

// Like AllocsPerRun() but returns an integer.
// TODO: actually compute allocations; for now return 0.
var AllocsPerRun2 = base.AllocsPerRun2

// ----------------------------------------
// T

type T = base.T

var NewT = base.NewT

type Report = base.Report

// ----------------------------------------
// B
// TODO: actually implement

type B = base.B

// ----------------------------------------
// PB
// TODO: actually implement

type PB = base.PB

type InternalTest = base.InternalTest

var RunTest = base.RunTest
