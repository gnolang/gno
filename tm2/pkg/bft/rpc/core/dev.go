package core

import (
	"os"
	"runtime/pprof"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

// UnsafeFlushMempool removes all transactions from the mempool.
func (env *Environment) UnsafeFlushMempool(ctx *rpctypes.Context) (*ctypes.ResultUnsafeFlushMempool, error) {
	env.Mempool.Flush()
	return &ctypes.ResultUnsafeFlushMempool{}, nil
}

// profFile is process-global because CPU profiling itself is process-wide;
// only one profiler can run at a time regardless of how many Environments
// exist.
var profFile *os.File

// UnsafeStartCPUProfiler starts a pprof profiler using the given filename.
func (env *Environment) UnsafeStartCPUProfiler(ctx *rpctypes.Context, filename string) (*ctypes.ResultUnsafeProfile, error) {
	var err error
	profFile, err = os.Create(filename)
	if err != nil {
		return nil, err
	}
	err = pprof.StartCPUProfile(profFile)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultUnsafeProfile{}, nil
}

// UnsafeStopCPUProfiler stops the running pprof profiler.
func (env *Environment) UnsafeStopCPUProfiler(ctx *rpctypes.Context) (*ctypes.ResultUnsafeProfile, error) {
	pprof.StopCPUProfile()
	if err := profFile.Close(); err != nil {
		return nil, err
	}
	return &ctypes.ResultUnsafeProfile{}, nil
}

// UnsafeWriteHeapProfile dumps a heap profile to the given filename.
func (env *Environment) UnsafeWriteHeapProfile(ctx *rpctypes.Context, filename string) (*ctypes.ResultUnsafeProfile, error) {
	memProfFile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	if err := pprof.WriteHeapProfile(memProfFile); err != nil {
		return nil, err
	}
	if err := memProfFile.Close(); err != nil {
		return nil, err
	}

	return &ctypes.ResultUnsafeProfile{}, nil
}
