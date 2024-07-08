This package is invalid because 'a' is defined twice.
NOTE: the Go parser itself returns an error for redefinitions in the same file,
but testing for redeclarations across files requires our own custom logic.
(arguably we should check ourself either way).
