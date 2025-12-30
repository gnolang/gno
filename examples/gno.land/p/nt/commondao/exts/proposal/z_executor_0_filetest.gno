// PKGPATH: gno.land/r/proposal/test
package test

import "gno.land/p/nt/commondao/exts/proposal"

func executor(realm) error {
	return nil
}

func main() {
	// Arrange
	def := proposal.MustNewDefinition("Title", "Body", proposal.WithExecutor(executor))

	// Act
	err := def.Execute()

	// Assert
	println(err == nil)
}

// Output:
// true
