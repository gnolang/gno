package profile

import (
	"fmt"
	"testing"

	"gno.land/r/profile"
)

func TestRender(t *testing.T) {
	got := profile.Render("")
	fmt.Println(got)
	//_ = profile.Render
	// various data types
	// avatar
	// ip address
}
