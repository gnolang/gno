// PKGPATH: gno.land/r/definition/test
package test

import "gno.land/p/nt/commondao/v0/exts/definition"

func executor(string) error {
	return nil
}

func main(cur realm) {
	// Arrange
	def := definition.MustNew("Title", "Body", definition.WithExecutor(executor))
	fn := def.Executor()

	// Act
	err := fn(cur.PkgPath())

	// Assert
	println(err == nil)
}

// Output:
// true
