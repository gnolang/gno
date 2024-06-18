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

	if len(description) > 80 {
		panic(errDescTooLong)
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

func RenderHome() string {
	if events.Size() == 0 {
		return "No upcoming or past events."
	}

	output := ""

	output = "<div>\n\n"
	events.ReverseIterate("", "", func(key string, value interface{}) bool {
		e := value.(*Event)

		if e.startTime.After(time.Now()) {
			output += e.RenderEvent()
		}
		if e.startTime.Before(time.Now()) {
			output += e.RenderEvent()
		}

		return false
	})

	output = "</div>\n\n"

	return output
}

func (e Event) RenderEvent() string {
	var buf bytes.Buffer

	buf.WriteString("<div class='columns-3'>")

	buf.WriteString(ufmt.Sprintf("#### %s\n\n", e.name))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.description))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.location))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.startTime.Format("02 Jan 2006")))

	if e.startTime.Before(time.Now()) {
		buf.WriteString("Finished\n\n")
	}

	buf.WriteString("</div>")

	return buf.String()
}

func Render(path string) string {
	return RenderHome()
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
