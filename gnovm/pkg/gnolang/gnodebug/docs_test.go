package gnodebug

import (
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestDocFlags(t *testing.T) {
	packages.Load(&packages.Config{}, "github.com/gnolang/gno/gnovm/pkg/gnolang")
}
