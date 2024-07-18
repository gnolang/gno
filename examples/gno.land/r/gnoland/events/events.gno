package events

import (
	"bytes"
	"sort"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/ownable/exts/authorizable"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

type Event struct {
	id          seqid.ID
	name        string    // name of event
	description string    // short description of event
	link        string    // link to a corresponding web2 page, ie eventbrite/luma or conference page
	location    string    // location of the event
	startTime   time.Time // start time of the event
	endTime     time.Time // end time of the event
}

type eventsSlice []*Event

var (
	a         *authorizable.Authorizable
	events    eventsSlice // sorted oldest > newest
	idCounter seqid.ID
)

const (
	maxDescLength = 100
	maxWidgetSize = 5
)

func init() {
	su := std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn
	a = authorizable.NewAuthorizableWithAddress(su)

	events = make(eventsSlice, 0)
}

func AddEvent(name, description, link, location string, startTime, endTime int64) {
	a.AssertOnAuthList()

	if name == "" {
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

	e := &Event{
		id:          idCounter.Next(),
		name:        name,
		description: description,
		link:        link,
		location:    location,
		startTime:   time.Unix(startTime, 0), // given in unix seconds
		endTime:     time.Unix(endTime, 0),   // given in unix seconds
	}

	events = append(events, e)
	sort.Sort(events)
}

func DelEvent(id seqid.ID) {
	a.AssertOnAuthList()

	for i, event := range events {
		if event.id == id {
			events = append(events[:i], events[i+1:]...)
			return
		}
	}
}

func EditEvent(id seqid.ID, name, description, link, location string, startTime, endTime int64) {
	a.AssertOnAuthList()

	var prevStartTime time.Time
	newStartTime := time.Unix(startTime, 0)
	newEndTime := time.Unix(endTime, 0)

	for i, event := range events {
		if event.id == id {
			prevStartTime = event.startTime
			events[i] = &Event{
				id:          id,
				name:        name,
				description: description,
				link:        link,
				location:    location,
				startTime:   newStartTime,
				endTime:     newEndTime,
			}

			return
		}
	}

	// If sort order was disrupted, sort again
	if newStartTime != prevStartTime {
		sort.Sort(events)
	}
}

// Rendering

// RenderEventWidget shows up to amt of the latest events to a caller
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

	// todo check for ordering
	for _, event := range events[eventNum-amt:] {
		output += ufmt.Sprintf("- [%s](%s)\n", event.name, event.link)
	}

	return output
}

func renderHome(admin bool) string {
	output := "# gno.land events\n\n"

	if len(events) == 0 {
		output += "No upcoming or past events."
		return output
	}

	output += "Below is a list of all upcoming and past gno.land events.\n\n"

	output += "---\n\n"

	// Find position of latest past event
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

	return renderHome(false)
}

func GetOwner() std.Address {
	return a.Owner()
}

func AddModerator(mod std.Address) {
	a.AssertCallerIsOwner()

	if err := a.AddToAuthList(mod); err != nil {
		panic(err)
	}
}

// Len returns the length of the slice
func (m eventsSlice) Len() int {
	return len(m)
}

// Less compares the startTime fields of two elements
func (m eventsSlice) Less(i, j int) bool {
	return m[i].startTime.After(m[j].startTime)
}

// Swap swaps two elements in the slice
func (m eventsSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
