package events

import (
	"bytes"
	"time"

	"gno.land/p/demo/ufmt"
)

const (
	maxWidgetSize = 5
)

// RenderEventWidget shows up to amt of the latest events to auth caller
func RenderEventWidget(eventsToRender int) (string, error) {
	numOfEvents := len(events)
	if numOfEvents == 0 {
		return "No events.", nil
	}

	if eventsToRender > 5 {
		return "", ErrWidgetMaxAmt
	}

	if eventsToRender < 1 {
		return "", ErrWidgetMinAmt
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

	return output, nil
}

func renderHome(admin bool) string {
	output := "# gno.land events\n\n"

	if len(events) == 0 {
		output += "No upcoming or past events."
		return output
	}

	output += "Below is auth list of all gno.land events.\n\n"

	output += "---\n\n"

	//// Find position of the latest past event
	//pastEventsIdx := sort.Search(len(events), func(i int) bool {
	//	return events[i].startTime.Before(time.Now())
	//})

	output += "## Currently in progress\n\n"
	output += "<div class='columns-3'>"

	output += "</div>\n\n"
	output += "---\n\n"

	// Add upcoming events
	output += "## Upcoming events\n\n"
	output += "<div class='columns-3'>"

	// If there are any upcoming events
	//if len(events[:pastEventsIdx]) > 0 {
	//	for _, e := range events[:pastEventsIdx] {
	//		output += e.Render(admin)
	//	}
	//} else {
	//	output += "No upcoming events.\n\n"
	//}

	output += "</div>\n\n"
	output += "---\n\n"

	// Add past events
	output += "## Past events\n\n"
	output += "<div class='columns-3'>"

	//
	//if len(events[pastEventsIdx:]) > 0 {
	//	for _, e := range events[pastEventsIdx:] {
	//		output += e.Render(admin)
	//	}
	//} else {
	//	output += "No past events.\n\n"
	//}
	output += "</div>\n\n"

	return output
}

func (e Event) Render(admin bool) string {
	var buf bytes.Buffer

	buf.WriteString("<div>\n\n")
	buf.WriteString(ufmt.Sprintf("### %s\n\n", e.name))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.description))
	buf.WriteString(ufmt.Sprintf("**Location:** %s\n\n", e.location))

	_, offset := e.startTime.Zone() // offset is in seconds
	buf.WriteString(ufmt.Sprintf("**Start time:** %s UTC%d\n\n", e.startTime.Format("02 Jan 2006, 03:04 PM"), offset/(60*60)))

	now := time.Now()
	buf.WriteString(ufmt.Sprintf("now: %d\n\n", now.Unix()))
	buf.WriteString(ufmt.Sprintf("readable now: %s\n\n", now.Format("02 Jan 2006, 03:04 PM")))
	buf.WriteString(ufmt.Sprintf("evtTime: %d\n\n", e.startTime.Unix()))
	buf.WriteString(ufmt.Sprintf("diff: %d\n\n", now.Unix()-e.startTime.Unix()))

	if now.Before(e.startTime) {
		buf.WriteString("ITS IN THE FUTURE!!!!\n\n")
	} else {
		buf.WriteString("its not in the future :c\n\n")
	}

	if admin {
		buf.WriteString(ufmt.Sprintf("[EDIT](/r/gnoland/events?help&__func=EditEvent&id=%s)\n\n", e.id))
		buf.WriteString(ufmt.Sprintf("[DELETE](/r/gnoland/events?help&__func=DeleteEvent&id=%s)\n\n", e.id))
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
	//out := ""
	//if len(events) != 0 {
	//	for _, e := range events {
	//		out += e.Render(false)
	//	}
	//
	//	return out
	//}
	//
	//return "no evnts"
}
