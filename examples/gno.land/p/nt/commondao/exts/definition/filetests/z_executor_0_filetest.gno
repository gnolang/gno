// PKGPATH: gno.land/r/definition/test
package test

import "gno.land/p/nt/commondao/exts/definition"

func executor(realm) error {
	return nil
}

func main() {
	// Arrange
	def := definition.MustNew("Title", "Body", definition.WithExecutor(executor))
	fn := def.Executor()

	// Act
	err := fn(cross)

	// Assert
	println(err == nil)
}

// Output:
// true
