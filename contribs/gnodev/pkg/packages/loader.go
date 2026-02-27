package packages

type Loader interface {
	// Load resolves package paths and all their dependencies in the correct order.
	Load(paths ...string) ([]*Package, error)

	// Resolve processes a single package path and returns the corresponding Package.
	Resolve(path string) (*Package, error)

	// Name of the loader
	Name() string
}
