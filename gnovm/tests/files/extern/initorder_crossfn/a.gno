package initorder_crossfn

// A depends on GetB which is defined in b.gno.
// findDependentNames must recurse into GetB's body (via GetExternNames) to
// discover that A transitively depends on B, even across file boundaries.

var A = GetB()
