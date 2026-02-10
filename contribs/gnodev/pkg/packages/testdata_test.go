// This test file serves as a reference for the testdata directory tree.

package packages

// The structure of the testdata directory is as follows:
//
// testdata
// └── abc.xy
//     └── t
//         ├── aa
//         │   ├── file.gno
//         │   └── gnomod.toml
//         ├── bb
//         │   ├── file.gno
//         │   └── gnomod.toml
//         ├── cc
//         │   ├── file.gno
//         │   └── gnomod.toml
//         └── nested
//             ├── aa
//             │   ├── file.gno
//             │   └── gnomod.toml
//             └── nested
//                 ├── bb
//                 │   ├── file.gno
//                 │   └── gnomod.toml
//                 └── cc
//                     ├── file.gno
//                     └── gnomod.toml

const (
	TestdataPkgA = "abc.xy/t/aa"
	TestdataPkgB = "abc.xy/t/bb"
	TestdataPkgC = "abc.xy/t/cc"
)

// List of testdata package paths
var testdataPkgs = []string{TestdataPkgA, TestdataPkgB, TestdataPkgC}

const (
	TestdataNestedA = "abc.xy/t/nested/aa"        // Path to nested package A
	TestdataNestedB = "abc.xy/t/nested/nested/bb" // Path to nested package B
	TestdataNestedC = "abc.xy/t/nested/nested/cc" // Path to nested package C
)

// List of nested package paths
var testdataNested = []string{TestdataNestedA, TestdataNestedB, TestdataNestedC}
