package events

import (
	"bytes"
	"sort"
	"strconv"
	"time"

	"gno.land/p/demo/ufmt"
)

const (
	maxWidgetSize = 5
)

// RenderEventWidget shows up to amt of the latest events to auth caller
func RenderEventWidget(eventsToRender int) string {
	numOfEvents := len(events)
	if numOfEvents == 0 {
		return "No events."
	}

	if eventsToRender > 5 {
		panic("maximum number of events in widget is " + strconv.Itoa(maxWidgetSize))
	}

	if eventsToRender < 1 {
		panic("you need to request at least 1 event to render")
	}

	if eventsToRender > numOfEvents {
		eventsToRender = numOfEvents
	}

	output := ""

	for _, event := range events[numOfEvents-eventsToRender:] {
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

	// Find position of the latest past event
	pastEventsIdx := sort.Search(len(events), func(i int) bool {
		return events[i].startTime.Before(time.Now())
	})

	// Add upcoming events
	output += "## Upcoming events\n\n"
	output += "<div class='columns-3'>"

	// If there are any upcoming events
	if len(events[:pastEventsIdx]) > 0 {
		for _, e := range events[:pastEventsIdx] {
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

	if len(events[pastEventsIdx:]) > 0 {
		for _, e := range events[pastEventsIdx:] {
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
