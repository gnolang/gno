package main

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOpcodesPackage(t *testing.T) {
	dir := "../gno/opcodes"
	bstore := benchmarkDiskStore()
	pv := addPackage(bstore, dir, opcodesPkgPath)
	pb := pv.GetBlock(bstore)

	assert := assert.New(t)
	require := require.New(t)

	declTypes := []string{
		"foo",
		"dog",
		"foofighter",
	}
	for i := 0; i < len(declTypes); i++ {
		tv := pb.Values[i]
		v, ok := tv.V.(gno.TypeValue)
		require.True(ok, "it should be a TypeValue")
		dtv, ok2 := v.Type.(*gno.DeclaredType)
		tn := declTypes[i]

		require.True(ok2, "it should be a DeclaredType")
		assert.Equal(tn, string(dtv.Name), "the declared type name should be "+tn)
	}

	// These are the functions used to benchmark the OpCode in the benchmarking contract.
	// We call each to benchmark a group of OpCodes.
	funcValues := []string{
		"OpDecl",
		"OpEvalInt",
		"OpEvalFloat",
		"StmtOps",
		"ControlOps",
		"OpDefer",
		"OpUnary",
		"OpBinary",
		"ExprOps",
		"OpLor",
		"OpLand",
		"OpPanic",
		"OpTypeSwitch",
		"OpCallDeferNativeBody",
		"OpRange",
		"OpForLoop",
		"OpTypes",
		"OpOpValues",
	}

	for i := 3; i < 3+len(funcValues); i++ {
		j := i - 3
		tv := pb.Values[i]
		fv, ok := tv.V.(*gno.FuncValue)
		require.True(ok, "it should be a FuncValue")
		fn := funcValues[j]
		assert.Equal(fn, string(fv.Name), "the declared type name should be "+fn)
	}
}
