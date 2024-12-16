package buttons

import (
	"strings"
	"testing"
)

func TestRenderMotdLink(t *testing.T) {
	res := Render("motd")
	const wantLink = "/r/docs/buttons$help&func=UpdateMOTD&newmotd=Message!"
	if !strings.Contains(res, wantLink) {
		t.Fatalf("%s\ndoes not contain correct help page link: %s", res, wantLink)
	}
}
