package main

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadNativePackage(t *testing.T) {
	dir := "../../pkg/benchops/gno/native"
	diskStore := benchmarkDiskStore()
	gstore := diskStore.gnoStore
	t.Cleanup(func() { diskStore.Delete() })
	pv := addPackage(gstore, dir, nativePkgPath)
	pb := pv.GetBlock(gstore)

	assert := assert.New(t)
	require := require.New(t)

	// These are the functions used to benchmark the OpCode in the benchmarking contract.
	// We call each to benchmark a group of OpCodes.
	funcValues := []string{
		"Print_1",
		"Print_1000",
		"Print_10000",
	}

	for i := 0; i < len(funcValues); i++ {
		tv := pb.Values[i]
		fv, ok := tv.V.(*gno.FuncValue)
		require.True(ok, "it should be a FuncValue")
		fn := funcValues[i]
		assert.Equal(fn, string(fv.Name), "the declared type name should be "+fn)
	}
}
