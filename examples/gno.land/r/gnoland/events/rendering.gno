package events

import (
	"bytes"
	"time"

	"gno.land/p/demo/ufmt"
)

const (
	MaxWidgetSize = 5
)

// RenderEventWidget shows up to eventsToRender of the latest events to a caller
func RenderEventWidget(eventsToRender int) (string, error) {
	numOfEvents := len(events)
	if numOfEvents == 0 {
		return "No events.", nil
	}

	if eventsToRender > MaxWidgetSize {
		return "", ErrWidgetMaxAmt
	}

	if eventsToRender < 1 {
		return "", ErrWidgetMinAmt
	}

	if eventsToRender > numOfEvents {
		eventsToRender = numOfEvents
	}

	output := ""

	for _, event := range events[eventsToRender:] {
		output += ufmt.Sprintf("- [%s](%s)\n", event.name, event.link)
	}

	return output, nil
}

// renderHome renders the home page of the events realm
func renderHome(admin bool) string {
	output := "# gno.land events\n\n"

	if len(events) == 0 {
		output += "No upcoming or past events."
		return output
	}

	output += "Below is a list of all gno.land events, including in progress, upcoming, and past ones.\n\n"
	output += "---\n\n"

	var (
		inProgress = ""
		upcoming   = ""
		past       = ""
		now        = time.Now()
	)

	for _, e := range events {
		if now.Before(e.startTime) {
			upcoming += e.Render(admin)
		} else if now.After(e.endTime) {
			past += e.Render(admin)
		} else {
			inProgress += e.Render(admin)
		}
	}

	if upcoming != "" {
		// Add upcoming events
		output += "## Upcoming events\n\n"
		output += "<div class='columns-3'>"

		output += upcoming

		output += "</div>\n\n"
		output += "---\n\n"
	}

	if inProgress != "" {
		output += "## Currently in progress\n\n"
		output += "<div class='columns-3'>"

		output += inProgress

		output += "</div>\n\n"
		output += "---\n\n"
	}

	if past != "" {
		// Add past events
		output += "## Past events\n\n"
		output += "<div class='columns-3'>"

		output += past

		output += "</div>\n\n"
	}

	return output
}

// Render returns the markdown representation of a single event instance
func (e Event) Render(admin bool) string {
	var buf bytes.Buffer

	buf.WriteString("<div>\n\n")
	buf.WriteString(ufmt.Sprintf("### %s\n\n", e.name))
	buf.WriteString(ufmt.Sprintf("%s\n\n", e.description))
	buf.WriteString(ufmt.Sprintf("**Location:** %s\n\n", e.location))

	_, offset := e.startTime.Zone() // offset is in seconds
	hoursOffset := offset / (60 * 60)
	sign := ""
	if offset >= 0 {
		sign = "+"
	}

	buf.WriteString(ufmt.Sprintf("**Start time:** %s UTC%s%d\n\n", e.startTime.Format("02 Jan 2006, 03:04 PM"), sign, hoursOffset))
	buf.WriteString(ufmt.Sprintf("**End time:** %s UTC%s%d\n\n", e.endTime.Format("02 Jan 2006, 03:04 PM"), sign, hoursOffset))

	if admin {
		buf.WriteString(ufmt.Sprintf("[EDIT](/r/gnoland/events?help&__func=EditEvent&id=%s)\n\n", e.id))
		buf.WriteString(ufmt.Sprintf("[DELETE](/r/gnoland/events?help&__func=DeleteEvent&id=%s)\n\n", e.id))
	}

	if e.link != "" {
		buf.WriteString(ufmt.Sprintf("[Website](%s)\n\n", e.link))
	}

	buf.WriteString("</div>")

	return buf.String()
}

// Render is the main rendering entry point
func Render(path string) string {
	if path == "admin" {
		return renderHome(true)
	}

	return renderHome(false)
}
