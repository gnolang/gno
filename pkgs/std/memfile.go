package std

type MemFile struct {
	Name string
	Body string
}

// NOTE: in the future, a MemPackage may represent
// updates/additional-files for an existing package.
type MemPackage struct {
	Name  string
	Path  string
	Files []MemFile
}
