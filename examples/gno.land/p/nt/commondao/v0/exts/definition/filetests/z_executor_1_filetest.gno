// PKGPATH: gno.land/r/definition/test
package test

import (
	"errors"

	"gno.land/p/nt/commondao/v0/exts/definition"
)

var wantErr = errors.New("test error")

func executor(realm) error {
	return wantErr
}

func main() {
	// Arrange
	def := definition.MustNew("Title", "Body", definition.WithExecutor(executor))
	fn := def.Executor()

	// Act
	err := fn(cross)

	// Assert
	println(err == wantErr)
}

// Output:
// true
