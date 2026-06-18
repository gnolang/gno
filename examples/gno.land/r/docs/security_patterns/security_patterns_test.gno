package securitypatterns

import (
	"strings"
	"testing"
)

func TestSetMessageRequiresLiveAdmin(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(admin))
	SetMessage(cross(cur), "hello")
	if Message() != "hello" {
		t.Fatalf("message was not updated")
	}

	testing.SetRealm(testing.NewUserRealm(address("g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh")))
	if rec := revive(func() { SetMessage(cross(cur), "attacker") }); rec == nil {
		t.Fatalf("expected non-admin caller to panic")
	}
	if Message() != "hello" {
		t.Fatalf("non-admin caller changed message")
	}
}

func TestRenderEscapesMarkdown(t *testing.T) {
	out := Render("[x](javascript:alert(1)) > quote")
	if strings.Contains(out, "[x](javascript") {
		t.Fatalf("render output contains raw markdown link: %s", out)
	}
	if !strings.Contains(out, "\\[x\\]\\(javascript:alert\\(1\\)\\)") {
		t.Fatalf("render output did not escape markdown punctuation: %s", out)
	}
	if !strings.Contains(out, "\\> quote") {
		t.Fatalf("render output did not escape markdown blockquote marker: %s", out)
	}
}
