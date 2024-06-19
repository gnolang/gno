package events

import (
	"bytes"
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"sort"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/ownable/exts/authorizable"
	"gno.land/p/demo/ufmt"
)

type Event struct {
	id          seqid.ID
	name        string    // name of event
	description string    // short description of event
	link        string    // link to a corresponding web2 sign up page, ie eventbrite/luma
	location    string    // location of the event
	startTime   time.Time // start time of the event
	// FIXME: add duration/endtime?
}

type eventsSlice []*Event

var (
	a         *authorizable.Authorizable
	events    eventsSlice
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

func AddEvent(name, description, link, location string, startTime int64) {
	a.AssertOnAuthList()

	if name == "" {
		panic(errEmptyName)
	}

	if startTime <= 0 {
		panic(errInvalidStartTime)
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
	}

	events = append(events, e)
	sort.Sort(events)
}

func DelEvent(id seqid.ID) {
	a.AssertOnAuthList()

	sort.Find()

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
		amt = eventNum - 1
	}

	output := ""

	for _, _ = range events[eventNum-amt:] {
		//output += ufmt.Sprintf("- [%s](%s)\n", event.name, event.link)
		output += "1"
	}

	return output
}

func renderHome() string {
	output := "# Gno.land Events\n\n"

	if len(events) == 0 {
		output += "No upcoming or past events."
		return output
	}

	output += "Below is a list of all upcoming and past Gno.land events.\n\n"
	output += "---\n\n"

	// Find position of latest past event
	pos := sort.Search(len(events), func(i int) bool {
		return events[i].startTime.After(time.Now())
	})

	// Add events
	for i, e := range events {
		output += "<div class='columns-3'>"
		output += e.RenderListItem()
		output += "</div>\n\n"

		// Add separator if we are at the boundary between upcoming and past events
		if i == pos-1 {
			output += "---\n\n"
		}
	}

	output += "---\n\n"
	output += "##### *All times are local."

	return output
}

func (e Event) RenderListItem() string {
	var buf bytes.Buffer

	buf.WriteString("<div>\n\n")
	buf.WriteString(ufmt.Sprintf("### %s\n\n", e.name))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.description))
	buf.WriteString(ufmt.Sprintf("**Location:** %s\n\n", e.location))
	buf.WriteString(ufmt.Sprintf("**Time & Date:** %s\n\n", e.startTime.Format("02 Jan 2006, 15:04")))

	if e.startTime.Before(time.Now()) {
		buf.WriteString("*Event is over\n\n")
	}

	buf.WriteString(ufmt.Sprintf("[See more](%s)\n\n", e.link))

	buf.WriteString("\n</div>")

	return buf.String()
}

func Render(path string) string {
	return renderHome()
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
	return m[i].startTime.Before(m[j].startTime)
}

// Swap swaps two elements in the slice
func (m eventsSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
