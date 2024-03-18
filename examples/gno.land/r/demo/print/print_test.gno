package priint

import (
	"testing"
)

func Test(t *testing.T) {
	if Render("") == "" {
		t.Fail()
	}
}
