package multisig

import (
	"strconv"
	"strings"
	"testing"
)

// Runs last (filename order) so the code-realm switch at the end
// does not change the origin-send recipient seen by earlier tests.
func TestRenderAndWrappers(cur realm, t *testing.T) {
	at(t0, alice)
	for SafeCount() < 21 {
		Create(cross(cur), "extra", alice.String(), 1, 0)
	}
	home := Render("")
	if !strings.Contains(home, "Showing the latest 20") {
		t.Error("home should truncate")
	}

	id := Create(cross(cur), "pages", alice.String(), 1, 0)
	sid := strconv.FormatInt(id, 10)
	if !strings.Contains(Render(sid), "None yet") {
		t.Error("empty proposal list")
	}
	if !strings.Contains(Render(sid+"/99"), "Unknown proposal") {
		t.Error("unknown proposal page")
	}

	coin := ProposeCoin(cross(cur), id, dest, "ugnot", 1, t0+10_000, "")
	page := Render(sid + "/" + strconv.FormatInt(coin, 10))
	if !strings.Contains(page, "None.") {
		t.Error("proposal with no approvals")
	}
	Approve(cross(cur), id, coin)
	Revoke(cross(cur), id, coin)
	Cancel(cross(cur), id, coin)

	WatchDenom(cross(cur), id, "ugnot")
	WatchToken(cross(cur), id, "gno.land/r/demo/missing.NO")
	watched := Render(sid)
	if !strings.Contains(watched, "ugnot") || !strings.Contains(watched, "not in the registry") {
		t.Error("watched rows")
	}
	UnwatchDenom(cross(cur), id, "ugnot")
	UnwatchToken(cross(cur), id, "gno.land/r/demo/missing.NO")

	pause := ProposePause(cross(cur), id, t0+10_000, "")
	Approve(cross(cur), id, pause)
	Execute(cross(cur), id, pause)
	if !strings.Contains(Render(sid), "paused") {
		t.Error("paused page")
	}
	unpause := ProposeUnpause(cross(cur), id, t0+10_000, "")
	Approve(cross(cur), id, unpause)
	Execute(cross(cur), id, unpause)

	busy := Create(cross(cur), "busy", alice.String(), 1, 0)
	for i := 0; i < 21; i++ {
		ProposeCoin(cross(cur), busy, dest, "ugnot", 1, t0+10_000, "")
	}
	if !strings.Contains(Render(strconv.FormatInt(busy, 10)), "Showing the latest") {
		t.Error("proposal list should truncate")
	}

	call := ProposeCall(cross(cur), id, "gno.land/r/demo/worker", "secret", t0+10_000, "memo")
	Approve(cross(cur), id, call)
	Execute(cross(cur), id, call)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/worker"))
	if Consume(cross(cur), id, call) != "secret" {
		t.Error("consume")
	}
	if !strings.Contains(Render(sid+"/"+strconv.FormatInt(call, 10)), "already consumed") {
		t.Error("consumed page")
	}

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/evil"))
	caught := revive(func() { Deposit(cross(cur), id, "ugnot", 1) })
	msg := ""
	switch v := caught.(type) {
	case string:
		msg = v
	case error:
		msg = v.Error()
	}
	if !strings.Contains(msg, "direct user call") {
		t.Errorf("deposit caller: %v", caught)
	}
}
