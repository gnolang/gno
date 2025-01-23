// This test file serves as a reference for the testdata directory tree.

package packages

// The structure of the testdata directory is as follows:
//
// testdata
// ├── abc.xy
// ├── nested
// │   ├── a
// │   │   └── gno.mod
// │   └── nested
// │       ├── b
// │       │   └── gno.mod
// │       └── c
// │           └── gno.mod
// └── pkg
//     ├── a
//     │   ├── file1.gno
//     │   └── gno.mod
//     ├── b // depends on a
//     │   ├── file1.gno
//     │   └── gno.mod
//     └── c // depends on b
//         ├── file1.gno
//         └── gno.mod

const (
	TestdataPkgA = "abc.xy/pkg/a"
	TestdataPkgB = "abc.xy/pkg/b"
	TestdataPkgC = "abc.xy/pkg/c"
)

// List of testdata package paths
var testdataPkgs = []string{TestdataPkgA, TestdataPkgB, TestdataPkgC}

const (
	TestdataNestedA = "abc.xy/nested/a"        // Path to nested package A
	TestdataNestedB = "abc.xy/nested/nested/b" // Path to nested package B
	TestdataNestedC = "abc.xy/nested/nested/c" // Path to nested package C
)

// List of nested package paths
var testdataNested = []string{TestdataNestedA, TestdataNestedB, TestdataNestedC}
