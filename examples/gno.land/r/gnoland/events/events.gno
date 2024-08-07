package events

import (
	"sort"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/ownable/exts/authorizable"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

type Event struct {
	id          seqid.ID
	name        string    // name of event
	description string    // short description of event
	link        string    // link to auth corresponding web2 page, ie eventbrite/luma or conference page
	location    string    // location of the event
	startTime   time.Time // start time of the event
	endTime     time.Time // end time of the event
}

type eventsSlice []*Event

var (
	events    eventsSlice // oldest event has the largest index
	idCounter seqid.ID
)

const (
	maxDescLength = 100
	EventAdded    = "EventAdded"
	EventDeleted  = "EventDeleted"
	EventEdited   = "EventEdited"
)

func init() {
	su := std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn
	auth = authorizable.NewAuthorizableWithAddress(su)

	events = make(eventsSlice, 0)
}

// AddEvent adds auth new event
// Start time & end time need to be specified in unix seconds
func AddEvent(name, description, link, location string, startTime, endTime int64) seqid.ID {
	auth.AssertOnAuthList()

	if strings.TrimSpace(name) == "" {
		panic(ErrEmptyName)
	}

	if startTime <= 0 {
		panic(ErrInvalidStartTime)
	}

	if endTime <= startTime {
		panic(ErrInvalidEndTime)
	}

	if len(description) > maxDescLength {
		str := ufmt.Sprintf("event description is %d chars, while it cannot be more than %d chars", len(description), maxDescLength)
		panic(str)
	}

	id := idCounter.Next()
	e := &Event{
		id:          id,
		name:        name,
		description: description,
		link:        link,
		location:    location,
		startTime:   time.Unix(startTime, 0), // given in unix seconds
		endTime:     time.Unix(endTime, 0),   // given in unix seconds
	}

	events = append(events, e)
	sort.Sort(events)

	std.Emit(EventAdded,
		"id",
		e.id.String(),
	)

	return id
}

// DeleteEvent deletes an event with auth given ID
func DeleteEvent(id seqid.ID) {
	auth.AssertOnAuthList()

	e, idx, err := GetEventByID(id)
	if err != nil {
		panic(err)
	}

	events = append(events[:idx], events[idx+1:]...)

	std.Emit(EventDeleted,
		"id",
		e.id.String(),
	)
}

// EditEvent edits an event with auth given ID
// It only updates values corresponding to non-empty/non-nil arguments sent with the call
func EditEvent(id seqid.ID, name, description, link, location string, startTime, endTime int64) {
	auth.AssertOnAuthList()

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

	if startTime > 0 {
		e.startTime = time.Unix(startTime, 0)
	}

	if endTime < 0 || endTime < startTime {
		panic(ErrInvalidEndTime)
	}

	e.endTime = time.Unix(endTime, 0)

	std.Emit(EventEdited,
		"id",
		e.id.String(),
	)

	// If sort order was disrupted, sort again
	if startTime > 0 {
		sort.Sort(events)
	}
}

func GetEventByID(id seqid.ID) (*Event, int, error) {
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
