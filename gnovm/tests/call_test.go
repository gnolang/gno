package	tests

import (
	"testing"
	"bytes"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestCreatePackage(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdout := os.Stdout
	stderr := new(bytes.Buffer)
	store := TestStore("..", "", stdin)
}