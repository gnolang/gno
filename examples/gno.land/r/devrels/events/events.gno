// Package events allows you to upload data about specific IRL/online events
// It includes dynamic support for updating rendering events based on their
// status, ie if they are upcoming, in progress, or in the past.
package events

import (
	"chain"
	"sort"
	"strings"
	"time"

	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
)

type (
	Event struct {
		id          string
		name        string    // name of event
		description string    // short description of event
		link        string    // link to auth corresponding web2 page, ie eventbrite/luma or conference page
		location    string    // location of the event
		startTime   time.Time // given in RFC3339
		endTime     time.Time // end time of the event, given in RFC3339
	}

	eventsSlice []*Event
)

var (
	Ownable   = ownable.NewWithAddress(address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")) // @leohhhn
	events    = make(eventsSlice, 0)                                                        // sorted
	idCounter seqid.ID
)

const (
	maxDescLength = 100
	EventAdded    = "EventAdded"
	EventDeleted  = "EventDeleted"
	EventEdited   = "EventEdited"
)

// AddEvent adds auth new event
// Start time & end time need to be specified in RFC3339, ie 2024-08-08T12:00:00+02:00
func AddEvent(_ realm, name, description, link, location, startTime, endTime string) (string, error) {
	Ownable.AssertOwnedByPrevious()

	if strings.TrimSpace(name) == "" {
		return "", ErrEmptyName
	}

	if len(description) > maxDescLength {
		return "", ufmt.Errorf("%s: provided length is %d, maximum is %d", ErrDescriptionTooLong, len(description), maxDescLength)
	}

	// Parse times
	st, et, err := parseTimes(startTime, endTime)
	if err != nil {
		return "", err
	}

	id := idCounter.Next().String()
	e := &Event{
		id:          id,
		name:        name,
		description: description,
		link:        link,
		location:    location,
		startTime:   st,
		endTime:     et,
	}

	events = append(events, e)
	sort.Sort(events)

	chain.Emit(EventAdded,
		"id", e.id,
	)

	return id, nil
}

// DeleteEvent deletes an event with auth given ID
func DeleteEvent(_ realm, id string) {
	Ownable.AssertOwnedByPrevious()

	e, idx, err := GetEventByID(id)
	if err != nil {
		panic(err)
	}

	events = append(events[:idx], events[idx+1:]...)

	chain.Emit(EventDeleted,
		"id", e.id,
	)
}

// EditEvent edits an event with auth given ID
// It only updates values corresponding to non-empty arguments sent with the call
// Note: if you need to update the start time or end time, you need to provide both every time
func EditEvent(_ realm, id string, name, description, link, location, startTime, endTime string) {
	Ownable.AssertOwnedByPrevious()

	e, _, err := GetEventByID(id)
	if err != nil {
		panic(err)
	}

	// Set only valid values
	if strings.TrimSpace(name) != "" {
		e.name = name
	}

	if strings.TrimSpace(description) != "" {
		e.description = description
	}

	if strings.TrimSpace(link) != "" {
		e.link = link
	}

	if strings.TrimSpace(location) != "" {
		e.location = location
	}

	if strings.TrimSpace(startTime) != "" || strings.TrimSpace(endTime) != "" {
		st, et, err := parseTimes(startTime, endTime)
		if err != nil {
			panic(err) // need to also revert other state changes
		}

		oldStartTime := e.startTime
		e.startTime = st
		e.endTime = et

		// If sort order was disrupted, sort again
		if oldStartTime != e.startTime {
			sort.Sort(events)
		}
	}

	chain.Emit(EventEdited,
		"id", e.id,
	)
}

func GetEventByID(id string) (*Event, int, error) {
	for i, event := range events {
		if event.id == id {
			return event, i, nil
		}
	}

	return nil, -1, ErrNoSuchID
}

// Len returns the length of the slice
func (m eventsSlice) Len() int {
	return len(m)
}

// Less compares the startTime fields of two elements
// In this case, events will be sorted by largest startTime first (upcoming > past)
func (m eventsSlice) Less(i, j int) bool {
	return m[i].startTime.After(m[j].startTime)
}

// Swap swaps two elements in the slice
func (m eventsSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// parseTimes parses the start and end time for an event and checks for possible errors
func parseTimes(startTime, endTime string) (time.Time, time.Time, error) {
	st, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return time.Time{}, time.Time{}, ufmt.Errorf("%s: %s", ErrInvalidStartTime, err.Error())
	}

	et, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return time.Time{}, time.Time{}, ufmt.Errorf("%s: %s", ErrInvalidEndTime, err.Error())
	}

	if et.Before(st) {
		return time.Time{}, time.Time{}, ErrEndBeforeStart
	}

	_, stOffset := st.Zone()
	_, etOffset := et.Zone()
	if stOffset != etOffset {
		return time.Time{}, time.Time{}, ErrStartEndTimezonemMismatch
	}

	return st, et, nil
}
