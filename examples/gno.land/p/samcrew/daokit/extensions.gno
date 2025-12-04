package daokit

// Extensions system enables DAOs to expose additional functionality that can be
// accessed by other packages or realms. It provides a secure way to make
// specific DAO capabilities available without exposing internal implementation details.

import "gno.land/p/nt/avl"

// Interface must be implemented by any object that wants to be
// registered as a DAO extension. Extensions expose functionality through
// well-defined interfaces while maintaining encapsulation.
type Extension interface {
	// Returns metadata about this extension including its path, version,
	// query path for external access, and privacy settings.
	Info() ExtensionInfo
}

type ExtensionInfo struct {
	Path      string // Unique extension identifier (e.g., "gno.land/p/demo/basedao.MembersView")
	Version   string // Extension version (e.g., "1", "2.0", etc.)
	QueryPath string // Path for external queries to access this extension's data
	Private   bool   // If true, extension is only accessible from the same realm that registered it
}

type ExtensionsList interface {
	// Returns the total number of registered extensions.
	Len() int
	// Returns extension info at the specified index, or nil if index is out of bounds.
	Get(index int) *ExtensionInfo
	// Returns a slice of ExtensionInfo from startIndex to endIndex.
	Slice(startIndex, endIndex int) []ExtensionInfo
	// Iterates over all extensions, calling fn for each one.
	// Iteration stops if fn returns true.
	ForEach(fn func(index int, value ExtensionInfo) bool)
}

type ExtensionsStore struct {
	Tree avl.Tree
}

// Registers a new extension using its path as the key.
// Returns true if the extension was successfully registered.
// If an extension with the same path already exists, it will be replaced.
func (es *ExtensionsStore) Set(ext Extension) bool {
	info := ext.Info()
	return es.Tree.Set(info.Path, ext)
}

func (es *ExtensionsStore) Get(path string) (Extension, bool) {
	iface, ok := es.Tree.Get(path)
	if !ok {
		return nil, false
	}
	return iface.(Extension), true
}

// Remove unregisters an extension by its path.
// Returns the removed extension and true if found, nil and false otherwise.
func (es *ExtensionsStore) Remove(path string) (Extension, bool) {
	iface, ok := es.Tree.Remove(path)
	if !ok {
		return nil, false
	}
	return iface.(Extension), true
}

// Returns an ExtensionsList interface for iterating over all registered extensions.
// This provides read-only access to the extensions registry.
func (es *ExtensionsStore) List() ExtensionsList {
	return &extensionsList{es: es}
}

type extensionsList struct {
	es *ExtensionsStore
}

// Iterates over all extensions, calling fn for each extension with its index and info.
// Iteration stops early if fn returns true.
func (e *extensionsList) ForEach(fn func(index int, value ExtensionInfo) bool) {
	if e.es.Tree.Size() == 0 {
		return
	}

	index := 0
	e.es.Tree.IterateByOffset(0, e.es.Tree.Size(), func(_ string, value any) bool {
		result := fn(index, value.(Extension).Info())
		index++
		return result
	})
}

// Returns the ExtensionInfo at the specified index.
func (e *extensionsList) Get(index int) *ExtensionInfo {
	if index < 0 || index >= e.es.Tree.Size() {
		return nil
	}

	_, value := e.es.Tree.GetByIndex(index)
	info := value.(Extension).Info()
	return &info
}

// Returns the total number of registered extensions.
func (e *extensionsList) Len() int {
	return e.es.Tree.Size()
}

// Returns a slice of ExtensionInfo from startIndex (inclusive) to endIndex (exclusive).
// Bounds are automatically normalized: negative startIndex becomes 0, endIndex beyond size becomes size.
func (e *extensionsList) Slice(startIndex int, endIndex int) []ExtensionInfo {
	size := e.es.Tree.Size()

	// Normalize bounds
	if startIndex < 0 {
		startIndex = 0
	}

	if endIndex > size {
		endIndex = size
	}

	if startIndex >= endIndex {
		return nil
	}

	count := endIndex - startIndex
	result := make([]ExtensionInfo, count)

	i := 0
	e.es.Tree.IterateByOffset(startIndex, count, func(_ string, value any) bool {
		result[i] = value.(Extension).Info()
		i++
		return false
	})

	return result
}
