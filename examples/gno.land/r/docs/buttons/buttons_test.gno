package buttons

import (
	"strings"
	"testing"
)

func TestRenderHomeListsBothPrimitives(t *testing.T) {
	res := Render("")
	if !strings.Contains(res, "<gno-form>") {
		t.Errorf("home page should mention the gno-form primitive:\n%s", res)
	}
	if !strings.Contains(res, "txlink.Call") {
		t.Errorf("home page should mention the txlink primitive:\n%s", res)
	}
}

func TestRenderMotdHasForm(t *testing.T) {
	res := Render("motd")
	const wantForm = `<gno-form exec="UpdateMOTD">`
	if !strings.Contains(res, wantForm) {
		t.Fatalf("%s\ndoes not contain the gno-form submission block: %s", res, wantForm)
	}
}

func TestRenderMotdHasHelpLink(t *testing.T) {
	res := Render("motd")
	const wantLink = "/r/docs/buttons$help&func=UpdateMOTD&newmotd=Message%21"
	if !strings.Contains(res, wantLink) {
		t.Fatalf("%s\ndoes not contain correct help page link: %s", res, wantLink)
	}
}

func TestRenderMotdZeroStateThenUpdate(cur realm, t *testing.T) {
	if got := Render("motd"); !strings.Contains(got, "Not updated yet") {
		t.Fatalf("before any update, render should say it is unset:\n%s", got)
	}

	UpdateMOTD(cross(cur), "Hello chain")

	got := Render("motd")
	if !strings.Contains(got, "Hello chain") {
		t.Fatalf("render should echo the new message:\n%s", got)
	}
	if !strings.Contains(got, "Last updated by") {
		t.Fatalf("render should attribute the last caller after an update:\n%s", got)
	}
}

func TestRenderMotdSanitizesMessage(cur realm, t *testing.T) {
	UpdateMOTD(cross(cur), "danger!")

	got := Render("motd")
	if strings.Contains(got, "danger!") {
		t.Fatalf("active markdown punctuation must be escaped in the rendered motd:\n%s", got)
	}
	if !strings.Contains(got, `danger\!`) {
		t.Fatalf("expected the escaped form of the message:\n%s", got)
	}
}
