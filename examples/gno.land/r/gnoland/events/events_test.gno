package events

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
)

var (
	su      = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn
	suRealm = std.NewUserRealm(su)
)

func TestAddEvent(t *testing.T) {
	std.TestSetOrigCaller(su)
	std.TestSetRealm(suRealm)

	now := time.Now()

	e1Start := now.Add(time.Hour * 24 * 5)
	e1End := e1Start.Add(time.Hour * 4)
	AddEvent("Event 1", "this event is upcoming", "gno.land", "gnome land", e1Start.Unix(), e1End.Unix())

	got := renderHome(false)

	if !strings.Contains(got, "Event 1") {
		t.Fatalf("Expected to find Event 1 in render")
	}

	e2Start := now.Add(-time.Hour * 24 * 5)
	e2End := e2Start.Add(time.Hour * 4)

	AddEvent("Event 2", "this event is in the past", "gno.land", "gnome land", e2Start.Unix(), e2End.Unix())

	got = renderHome(false)

	upcomingPos := strings.Index(got, "## Upcoming events")
	pastPos := strings.Index(got, "## Past events")

	e1Pos := strings.Index(got, "Event 1")
	e2Pos := strings.Index(got, "Event 2")

	// expected index ordering: upcoming < e1 < past < e2
	if e1Pos < upcomingPos || e1Pos > pastPos {
		t.Fatalf("Expected to find Event 1 in Upcoming events")
	}

	if e2Pos < upcomingPos || e2Pos < pastPos || e2Pos < e1Pos {
		t.Fatalf("Expected to find Event 2 on a different pos")
	}

	// larger index => smaller startTime (future => past)
	if events[0].startTime.Unix() < events[1].startTime.Unix() {
		t.Fatalf("expected ordering to be different")
	}
}

func TestAddEventErrors(t *testing.T) {
	std.TestSetOrigCaller(su)
	std.TestSetRealm(suRealm)

	uassert.PanicsWithMessage(t, ErrEmptyName.Error(), func() {
		AddEvent("", "sample desc", "gno.land", "gnome land", 123, 456)
	})

	uassert.PanicsWithMessage(t, ErrInvalidStartTime.Error(), func() {
		AddEvent("sample name", "sample desc", "gno.land", "gnome land", 0, 456)
	})

	uassert.PanicsWithMessage(t, ErrInvalidEndTime.Error(), func() {
		AddEvent("sample name", "sample desc", "gno.land", "gnome land", 123, 0)
	})

	uassert.PanicsWithMessage(t, ErrInvalidEndTime.Error(), func() {
		AddEvent("sample name", "sample desc", "gno.land", "gnome land", 123, 1)
	})

	tooLongDesc := `Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean ma`
	// figure out a way to check panics with dynamic params in a better way
	uassert.PanicsWithMessage(t, "event description is 101 chars, while it cannot be more than 100 chars", func() {
		AddEvent("sample name", tooLongDesc, "gno.land", "gnome land", 123, 456)
	})
}

func TestDelEvent(t *testing.T) {
	events = nil // remove elements from previous tests - see issue #1982

	id := AddEvent("ToDelete", "sample desc", "gno.land", "gnome land", 123, 456)
	got := renderHome(false)

	if !strings.Contains(got, "ToDelete") {
		t.Fatalf("Expected to find ToDelete event in render")
	}

	DelEvent(id)
	got = renderHome(false)

	if strings.Contains(got, "ToDelete") {
		t.Fatalf("Did not expect to find ToDelete event in render")
	}
}

func TestEditEvent(t *testing.T) {
	events = nil // remove elements from previous tests - see issue #1982
	id := AddEvent("ToDelete", "sample desc", "gno.land", "gnome land", 123, 456)

}
