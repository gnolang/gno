package echo

import (
	"testing"
)

func Test(t *testing.T) {
	if Render("aa") != "aa" {
		t.Fail()
	}
	if Render("") != "" {
		t.Fail()
	}
}
