// This test file serves as a reference for the testdata directory tree.

package packages

// The structure of the testdata directory is as follows:
//
// testdata
// ├── abc.xy
// ├── nested
// │   ├── aa
// │   │   └── gno.mod
// │   └── nested
// │       ├── bb
// │       │   └── gno.mod
// │       └── cc
// │           └── gno.mod
// └── pkg
//     ├── aa
//     │   ├── file1.gno
//     │   └── gno.mod
//     ├── bb // depends on aa
//     │   ├── file1.gno
//     │   └── gno.mod
//     └── cc // depends on bb
//         ├── file1.gno
//         └── gno.mod

const (
	TestdataPkgA = "abc.xy/pkg/aa"
	TestdataPkgB = "abc.xy/pkg/bb"
	TestdataPkgC = "abc.xy/pkg/cc"
)

// List of testdata package paths
var testdataPkgs = []string{TestdataPkgA, TestdataPkgB, TestdataPkgC}

const (
	TestdataNestedA = "abc.xy/nested/aa"        // Path to nested package A
	TestdataNestedB = "abc.xy/nested/nested/bb" // Path to nested package B
	TestdataNestedC = "abc.xy/nested/nested/cc" // Path to nested package C
)

// List of nested package paths
var testdataNested = []string{TestdataNestedA, TestdataNestedB, TestdataNestedC}
