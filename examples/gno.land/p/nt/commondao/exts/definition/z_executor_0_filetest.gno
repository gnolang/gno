// PKGPATH: gno.land/r/definition/test
package test

import "gno.land/p/nt/commondao/exts/definition"

func executor(realm) error {
	return nil
}

func main() {
	// Arrange
	def := definition.MustNew("Title", "Body", definition.WithExecutor(executor))

	// Act
	err := def.Execute()

	// Assert
	println(err == nil)
}

// Output:
// true
