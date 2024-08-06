package events

import (
	"bytes"
	"sort"
	"std"
	"strconv"
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
	auth      *authorizable.Authorizable
	events    eventsSlice // oldest event has the largest index
	idCounter seqid.ID
)

const (
	maxDescLength = 100
	maxWidgetSize = 5
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

// DelEvent deletes an event with auth given ID
func DelEvent(id seqid.ID) {
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

// Rendering

// RenderEventWidget shows up to amt of the latest events to auth caller
func RenderEventWidget(amt int) string {
	eventNum := len(events)
	if eventNum == 0 {
		return "No events."
	}

	if amt > 5 {
		panic("maximum number of events in widget is " + strconv.Itoa(maxWidgetSize))
	}

	if amt < 1 {
		panic("amt needs to be more than 0")
	}

	if amt > eventNum {
		amt = eventNum
	}

	output := ""

	for _, event := range events[eventNum-amt:] {
		if event.startTime.After(time.Now()) {
			output += ufmt.Sprintf("- [%s](%s)\n", event.name, event.link)
		}
	}

	return output
}

func renderHome(admin bool) string {
	output := "# gno.land events\n\n"

	if len(events) == 0 {
		output += "No upcoming or past events."
		return output
	}

	output += "Below is auth list of all upcoming and past gno.land events.\n\n"
	output += ufmt.Sprintf(time.Now().Format(time.RFC850))

	output += "---\n\n"

	// Find position of la1test past event
	pos := sort.Search(len(events), func(i int) bool {
		return events[i].startTime.Before(time.Now())
	})

	// Add upcoming events
	output += "## Upcoming events\n\n"
	output += "<div class='columns-3'>"

	// If there are any upcoming events
	if len(events[:pos]) > 0 {
		for _, e := range events[:pos] {
			output += e.Render(admin)
		}
	} else {
		output += "No upcoming events.\n\n"
	}

	output += "</div>\n\n"
	output += "---\n\n"

	// Add past events
	output += "## Past events\n\n"
	output += "<div class='columns-3'>"

	if len(events[pos:]) > 0 {
		for _, e := range events[pos:] {
			output += e.Render(admin)
		}
	} else {
		output += "No past events.\n\n"
	}

	output += "</div>\n\n"

	output += "---\n\n"
	output += "#### *All times are local."

	return output
}

func (e Event) Render(admin bool) string {
	var buf bytes.Buffer

	buf.WriteString("<div>\n\n")
	buf.WriteString(ufmt.Sprintf("### %s\n\n", e.name))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.description))
	buf.WriteString(ufmt.Sprintf("**Location:** %s\n\n", e.location))
	buf.WriteString(ufmt.Sprintf("**Start time:** %s\n\n", e.startTime.Format("02 Jan 2006, 03:04 PM")))
	buf.WriteString(ufmt.Sprintf("**End time:** %s\n\n", e.endTime.Format("02 Jan 2006, 03:04 PM")))

	if admin {
		buf.WriteString(ufmt.Sprintf("[EDIT](/r/gnoland/events?help&__func=EditEvent&id=%s)\n\n", e.id))
		buf.WriteString(ufmt.Sprintf("[DELETE](/r/gnoland/events?help&__func=DelEvent&id=%s)\n\n", e.id))
	}

	if e.link != "" {
		buf.WriteString(ufmt.Sprintf("[See more](%s)\n\n", e.link))
	}

	buf.WriteString("</div>")

	return buf.String()
}

func Render(path string) string {
	if path == "admin" {
		return renderHome(true)
	}
	//out := ""
	//
	//for _, e := range events {
	//	out += e.Render(false)
	//}
	//return out
	return renderHome(false)
}

func GetOwner() std.Address {
	return auth.Owner()
}

func AddModerator(mod std.Address) {
	auth.AssertCallerIsOwner()

	if err := auth.AddToAuthList(mod); err != nil {
		panic(err)
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
