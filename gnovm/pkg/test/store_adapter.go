package test

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
)

// storeAdapter wraps a gnolang.Store to implement profiler.Store
type storeAdapter struct {
	store gnolang.Store
}

// NewStoreAdapter creates a new adapter that implements profiler.Store
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
