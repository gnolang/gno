package security_patterns

import (
	"strings"
	"testing"

	"gno.land/p/nt/markdown/sanitize/v0"
)

func TestSetMessageRequiresLiveAdmin(cur realm, t *testing.T) {
	// cross(cur) at the call site performs the realm crossing: PreviousRealm
	// inside SetMessage becomes the UserRealm set just above, so assertAdmin
	// can verify the caller's address.
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

func TestRenderEscapesMessage(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(admin))
	SetMessage(cross(cur), "[x](javascript:alert(1)) > quote")

	out := Render("")
	if strings.Contains(out, "[x](javascript") {
		t.Fatalf("message rendered as a live markdown link: %s", out)
	}
	if want := sanitize.InlineText("[x](javascript:alert(1)) > quote"); !strings.Contains(out, want) {
		t.Fatalf("message not escaped via sanitize.InlineText: %q not in %s", want, out)
	}
}

func TestRenderPathUsesSafeCodeSpan(t *testing.T) {
	// A backtick in path must not close the code span early; InlineCode picks
	// a wider backtick fence, whereas a naive "`" + path + "`" would let the
	// user's backtick break out.
	path := "a`b`c"
	out := Render(path)
	if want := sanitize.InlineCode(path); !strings.Contains(out, want) {
		t.Fatalf("path not rendered via a safe code span: %q not in %s", want, out)
	}
}
