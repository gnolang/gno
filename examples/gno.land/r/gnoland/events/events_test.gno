package events

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/uassert"
)

var (
	suRealm = std.NewUserRealm(su)
	now     = "2009-02-13T23:31:30Z" // time.Now() is hardcoded to this value in the gno test machine currently
)

func TestAddEvent(t *testing.T) {
	std.TestSetOrigCaller(su)
	std.TestSetRealm(suRealm)

	parsedTimeNow, err := time.Parse(time.RFC3339, now)
	if err != nil {
		panic(err)
	}

	e1Start := parsedTimeNow.Add(time.Hour * 24 * 5)
	e1End := e1Start.Add(time.Hour * 4)

	AddEvent("Event 1", "this event is upcoming", "gno.land", "gnome land", e1Start.Format(time.RFC3339), e1End.Format(time.RFC3339))

	got := renderHome(false)

	if !strings.Contains(got, "Event 1") {
		t.Fatalf("Expected to find Event 1 in render")
	}

	e2Start := parsedTimeNow.Add(-time.Hour * 24 * 5)
	e2End := e2Start.Add(time.Hour * 4)

	AddEvent("Event 2", "this event is in the past", "gno.land", "gnome land", e2Start.Format(time.RFC3339), e2End.Format(time.RFC3339))

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
		t.Fatalf("Expected to find Event 2 on auth different pos")
	}

	// larger index => smaller startTime (future => past)
	if events[0].startTime.Unix() < events[1].startTime.Unix() {
		t.Fatalf("expected ordering to be different")
	}
}

func TestAddEventErrors(t *testing.T) {
	std.TestSetOrigCaller(su)
	std.TestSetRealm(suRealm)

	_, err := AddEvent("", "sample desc", "gno.land", "gnome land", "2009-02-13T23:31:31Z", "2009-02-13T23:33:31Z")
	uassert.ErrorIs(t, err, ErrEmptyName)

	_, err = AddEvent("sample name", "sample desc", "gno.land", "gnome land", "", "2009-02-13T23:33:31Z")
	uassert.ErrorContains(t, err, ErrInvalidStartTime.Error())

	_, err = AddEvent("sample name", "sample desc", "gno.land", "gnome land", "2009-02-13T23:31:31Z", "")
	uassert.ErrorContains(t, err, ErrInvalidEndTime.Error())

	_, err = AddEvent("sample name", "sample desc", "gno.land", "gnome land", "2009-02-13T23:31:31Z", "2009-02-13T23:30:31Z")
	uassert.ErrorIs(t, err, ErrEndBeforeStart)

	_, err = AddEvent("sample name", "sample desc", "gno.land", "gnome land", "2009-02-13T23:31:31+06:00", "2009-02-13T23:33:31+02:00")
	uassert.ErrorIs(t, err, ErrStartEndTimezonemMismatch)

	tooLongDesc := `Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean ma`
	_, err = AddEvent("sample name", tooLongDesc, "gno.land", "gnome land", "2009-02-13T23:31:31Z", "2009-02-13T23:33:31Z")
	uassert.ErrorContains(t, err, ErrDescriptionTooLong.Error())
}

func TestDeleteEvent(t *testing.T) {
	events = nil // remove elements from previous tests - see issue #1982

	parsedTimeNow, err := time.Parse(time.RFC3339, now)
	if err != nil {
		panic(err)
	}

	e1Start := parsedTimeNow.Add(time.Hour * 24 * 5)
	e1End := e1Start.Add(time.Hour * 4)

	id, _ := AddEvent("ToDelete", "description", "gno.land", "gnome land", e1Start.Format(time.RFC3339), e1End.Format(time.RFC3339))

	got := renderHome(false)

	if !strings.Contains(got, "ToDelete") {
		t.Fatalf("Expected to find ToDelete event in render")
	}

	DeleteEvent(seqid.ID(id))
	got = renderHome(false)

	if strings.Contains(got, "ToDelete") {
		t.Fatalf("Did not expect to find ToDelete event in render")
	}
}

func TestEditEvent(t *testing.T) {
	events = nil // remove elements from previous tests - see issue #1982

	parsedTimeNow, err := time.Parse(time.RFC3339, now)
	if err != nil {
		panic(err)
	}

	e1Start := parsedTimeNow.Add(time.Hour * 24 * 5)
	e1End := e1Start.Add(time.Hour * 4)

	id, _ := AddEvent("ToDelete", "description", "gno.land", "gnome land", e1Start.Format(time.RFC3339), e1End.Format(time.RFC3339))

	newName := "New Name"
	newDesc := "Normal description"
	newST := int64(1234)
	newET := int64(4567)

	EditEvent(id, newName, newDesc, "", "", newST, newET)
	edited, _, _ := GetEventByID(id)

	if edited.name != newName {
		t.Fatalf("Expected %s, got %s", newName, edited.name)
	}

	if edited.description != newDesc {
		t.Fatalf("Expected %s, got %s", newDesc, edited.description)
	}

	if edited.startTime != time.Unix(newST, 0) {
		t.Fatalf("Expected %d, got %d", newST, edited.startTime.Unix())
	}

	if edited.endTime != time.Unix(newET, 0) {
		t.Fatalf("Expected %d, got %d", newET, edited.endTime.Unix())
	}
}

//func TestInvalidEdit(t *testing.T) {
//	events = nil // remove elements from previous tests - see issue #1982
//	id := AddEvent("ToEdit", "Desc to edit", "gno.land", "gnome land", 123, 456)
//
//	uassert.PanicsWithMessage(t, ErrInvalidEndTime.Error(), func() {
//		EditEvent(id, "", "", "", "", 0, -5)
//	})
//}
