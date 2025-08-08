package events

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

var (
	suRealm = std.NewUserRealm(su)

	now              = "2009-02-13T23:31:30Z" // time.Now() is hardcoded to this value in the gno test machine currently
	parsedTimeNow, _ = time.Parse(time.RFC3339, now)
)

func TestAddEvent(t *testing.T) {
	testing.SetRealm(suRealm)

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
	testing.SetRealm(suRealm)

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
	testing.SetRealm(suRealm)

	e1Start := parsedTimeNow.Add(time.Hour * 24 * 5)
	e1End := e1Start.Add(time.Hour * 4)

	id, _ := AddEvent("ToDelete", "description", "gno.land", "gnome land", e1Start.Format(time.RFC3339), e1End.Format(time.RFC3339))

	got := renderHome(false)

	if !strings.Contains(got, "ToDelete") {
		t.Fatalf("Expected to find ToDelete event in render")
	}

	DeleteEvent(id)
	got = renderHome(false)

	if strings.Contains(got, "ToDelete") {
		t.Fatalf("Did not expect to find ToDelete event in render")
	}
}

func TestEditEvent(t *testing.T) {
	testing.SetRealm(suRealm)

	e1Start := parsedTimeNow.Add(time.Hour * 24 * 5)
	e1End := e1Start.Add(time.Hour * 4)
	loc := "gnome land"

	id, _ := AddEvent("ToDelete", "description", "gno.land", loc, e1Start.Format(time.RFC3339), e1End.Format(time.RFC3339))

	newName := "New Name"
	newDesc := "Normal description"
	newLink := "new Link"
	newST := e1Start.Add(time.Hour)
	newET := newST.Add(time.Hour)

	EditEvent(id, newName, newDesc, newLink, "", newST.Format(time.RFC3339), newET.Format(time.RFC3339))
	edited, _, _ := GetEventByID(id)

	// Check updated values
	uassert.Equal(t, edited.name, newName)
	uassert.Equal(t, edited.description, newDesc)
	uassert.Equal(t, edited.link, newLink)
	uassert.True(t, edited.startTime.Equal(newST))
	uassert.True(t, edited.endTime.Equal(newET))

	// Check if the old values are the same
	uassert.Equal(t, edited.location, loc)
}

func TestInvalidEdit(t *testing.T) {
	testing.SetRealm(suRealm)

	uassert.PanicsWithMessage(t, ErrNoSuchID.Error(), func() {
		EditEvent("123123", "", "", "", "", "", "")
	})
}

func TestParseTimes(t *testing.T) {
	// times not provided
	// end time before start time
	// timezone Missmatch

	_, _, err := parseTimes("", "")
	uassert.ErrorContains(t, err, ErrInvalidStartTime.Error())

	_, _, err = parseTimes(now, "")
	uassert.ErrorContains(t, err, ErrInvalidEndTime.Error())

	_, _, err = parseTimes("2009-02-13T23:30:30Z", "2009-02-13T21:30:30Z")
	uassert.ErrorContains(t, err, ErrEndBeforeStart.Error())

	_, _, err = parseTimes("2009-02-10T23:30:30+02:00", "2009-02-13T21:30:33+05:00")
	uassert.ErrorContains(t, err, ErrStartEndTimezonemMismatch.Error())
}

func TestRenderEventWidget(t *testing.T) {
	testing.SetRealm(suRealm)

	// No events yet
	events = nil
	out, err := RenderEventWidget(1)
	uassert.NoError(t, err)
	uassert.Equal(t, out, "No events.")

	// Too many events
	out, err = RenderEventWidget(MaxWidgetSize + 1)
	uassert.ErrorIs(t, err, ErrMaxWidgetSize)

	// Too little events
	out, err = RenderEventWidget(0)
	uassert.ErrorIs(t, err, ErrMinWidgetSize)

	// Ordering & if requested amt is larger than the num of events that exist
	e1Start := parsedTimeNow.Add(time.Hour * 24 * 5)
	e1End := e1Start.Add(time.Hour * 4)

	e2Start := parsedTimeNow.Add(time.Hour * 24 * 10) // event 2 is after event 1
	e2End := e2Start.Add(time.Hour * 4)

	_, err = AddEvent("Event 1", "description", "gno.land", "loc", e1Start.Format(time.RFC3339), e1End.Format(time.RFC3339))
	urequire.NoError(t, err)

	_, err = AddEvent("Event 2", "description", "gno.land", "loc", e2Start.Format(time.RFC3339), e2End.Format(time.RFC3339))
	urequire.NoError(t, err)

	out, err = RenderEventWidget(MaxWidgetSize)
	urequire.NoError(t, err)

	uniqueSequence := "- [" // sequence that is displayed once per each event as per the RenderEventWidget function
	uassert.Equal(t, 2, strings.Count(out, uniqueSequence))

	uassert.True(t, strings.Index(out, "Event 1") > strings.Index(out, "Event 2"))
}
