package events

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"time"

	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"gno.land/p/demo/ownable/exts/authorizable"
)

type Event struct {
	name        string    // name of event
	description string    // short description of event
	link        string    // link to a corresponding web2 sign up page, ie eventbrite/luma
	location    string    // location of the event
	startTime   time.Time // start time of the event
	// add duration/endtime?
}

var (
	a         *authorizable.Authorizable
	events    *avl.Tree
	idCounter seqid.ID
)

func init() {
	su := std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq") // @moul
	a = authorizable.NewWithAddress(su)

	a.AddToAuthList(std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")) // @leohhhn

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
func RenderEventWidget(amt int) string {
	if events.Size() == 0 {
		return "No events."
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

		return i >= amt
	})

	return output
}

func RenderHome() string {
	if events.Size() == 0 {
		return "No upcoming or past events."
	}

	output := ""

	events.ReverseIterate("", "", func(key string, value interface{}) bool {
		e := value.(*Event)
		if e.startTime.After(time.Now()) {

			i++
		} // only return upcoming events

		return i >= amt
	})

	return output
}

// genID generates a unique id for the event
// By utilizing the AVL tree property which automatically sorts in lex order,
// we can automatically have events sorted by event start time
func genID(t time.Time) string {
	return t.Format(time.RFC3339) + "-" + idCounter.Next().String()
}
