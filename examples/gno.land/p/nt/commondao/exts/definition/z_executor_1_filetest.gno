// PKGPATH: gno.land/r/definition/test
package test

import (
	"errors"

	"gno.land/p/nt/commondao/exts/definition"
)

var wantErr = errors.New("test error")

func executor(realm) error {
	return wantErr
}

func main() {
	// Arrange
	def := definition.MustNew("Title", "Body", definition.WithExecutor(executor))

	// Act
	err := def.Execute()

	// Assert
	println(err == wantErr)
}

// Output:
// true
