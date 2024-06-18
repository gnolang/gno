package events

import (
	"bytes"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable/exts/authorizable"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

type Event struct {
	name        string    // name of event
	description string    // short description of event
	link        string    // link to a corresponding web2 sign up page, ie eventbrite/luma
	location    string    // location of the event
	startTime   time.Time // start time of the event
	// FIXME: add duration/endtime?
}

var (
	a         *authorizable.Authorizable
	events    *avl.Tree
	idCounter seqid.ID
)

const maxDescLength = 100

func init() {
	su := std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn
	a = authorizable.NewAuthorizableWithAddress(su)

	events = avl.NewTree()
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
		name:        name,
		description: description,
		link:        link,
		location:    location,
		startTime:   time.Unix(startTime, 0), // given in unix seconds
	}

	_ = events.Set(genID(e.startTime), e)
}

// RenderEventWidget shows up to amt of the latest events to a caller
func RenderEventWidget(amt uint) string {
	if events.Size() == 0 {
		return "No events."
	}

	if amt > 5 {
		panic("maximum number of events in widget is 5")
	}

	var (
		i      = 0
		output = ""
	)

	events.ReverseIterate("", "", func(key string, value interface{}) bool {
		e := value.(*Event)
		if e.startTime.After(time.Now()) {
			output += ufmt.Sprintf("[%s](%s)\n", e.name, e.link)
			i++
		} // only return upcoming events

		return i >= int(amt)
	})

	return output
}

func renderHome() string {
	output := "# Gno.land Events\n\n"

	if events.Size() == 0 {
		output += "No upcoming or past events."
		return output
	}

	output += "<div class='columns-3'>"

	events.ReverseIterate("", "", func(key string, value interface{}) bool {
		e := value.(*Event)

		if e.startTime.After(time.Now()) {
			output += e.RenderListItem()
		}
		if e.startTime.Before(time.Now()) {
			output += e.RenderListItem()
		}

		output += ufmt.Sprintf("<!-- EventID: %s -->", key)

		return false
	})

	output += "</div>"

	return output
}

func (e Event) RenderListItem() string {
	var buf bytes.Buffer

	buf.WriteString("<div>\n\n")
	buf.WriteString(ufmt.Sprintf("### %s\n\n", e.name))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.description))
	buf.WriteString(ufmt.Sprintf("**Location:** %s\n\n", e.location))
	buf.WriteString(ufmt.Sprintf("**Time & Date:** %s\n\n", e.startTime.Format("02 Jan 2006")))
	buf.WriteString(ufmt.Sprintf("[See more](%s)\n\n", e.link))

	if e.startTime.Before(time.Now()) {
		buf.WriteString("*Event is over\n\n")
	}

	buf.WriteString("\n</div>")

	return buf.String()
}

func Render(path string) string {
	return renderHome()
}

// genID generates a unique id for the event
// By utilizing the AVL tree property which automatically sorts in lex order,
// we can automatically have events sorted by event start time
func genID(t time.Time) string {
	return t.Format(time.RFC3339) + "-" + idCounter.Next().String()
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
