package gnoland_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/jaekwon/testify/require"
)

func TestNewTestingApp(t *testing.T) {
	app := gnoland.NewTestingApp()
	require.NotNil(t, app)
	println(app)
}

func ExampleNewTestingApp() {
	app := gnoland.NewTestingApp()
	fmt.Println(app)
	// Output:
	// ...
}
