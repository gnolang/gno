package test

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
)

// storeAdapter wraps a gnolang.Store to implement profiler.Store
// This adapter is necessary because the profiler package cannot directly
// depend on gnolang package to avoid circular dependencies
type storeAdapter struct {
	store gnolang.Store
}

// NewStoreAdapter creates a new adapter that implements profiler.Store
// It allows the profiler to access source files stored in memory during tests
func NewStoreAdapter(store gnolang.Store) profiler.Store {
	return &storeAdapter{store: store}
}

// GetMemFile implements profiler.Store
func (sa *storeAdapter) GetMemFile(pkgPath, name string) *profiler.MemFile {
	memFile := sa.store.GetMemFile(pkgPath, name)
	if memFile == nil {
		return nil
	}
	// Convert std.MemFile to profiler.MemFile
	return &profiler.MemFile{
		Name: memFile.Name,
		Body: memFile.Body,
	}
}
