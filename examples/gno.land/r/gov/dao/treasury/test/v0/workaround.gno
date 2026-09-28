package test

// This package exists solely to circumvent a limitation associated with the
// suffixed test package (a test package sharing the same folder as the main
// package to be tested but having the suffix _test in its name).
// Currently, the GnoVM no longer differentiates between the dependencies of a
// package and its test package, which causes circular dependencies issues.
