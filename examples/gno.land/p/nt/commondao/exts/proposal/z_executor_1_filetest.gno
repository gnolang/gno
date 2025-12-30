// PKGPATH: gno.land/r/proposal/test
package test

import (
	"errors"

	"gno.land/p/nt/commondao/exts/proposal"
)

var wantErr = errors.New("test error")

func executor(realm) error {
	return wantErr
}

func main() {
	// Arrange
	def := proposal.MustNewDefinition("Title", "Body", proposal.WithExecutor(executor))

	// Act
	err := def.Execute()

	// Assert
	println(err == wantErr)
}

// Output:
// true
