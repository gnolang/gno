package subscriptions

import (
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"
	"gno.land/p/sunspirit/table"
)

func Render(path string) string {
	res := ""
	res += md.H1("Subscriptions Service Module")
	res += md.Paragraph("This module provides a subscriptions service where users can subscribe to services, manage their subscriptions, and handle payments using native coins or GRC20 tokens.")
	res += md.H2("Services Table")
	t, _ := table.New(
		[]string{"Service Name", "Description", "Renewal Period", "Price", "Subscribers"},
		[][]string{},
	)
	services.Iterate("", "", func(key string, value interface{}) bool {
		svc := value.(*service)
		subscriberCount := ufmt.Sprintf("%d", svc.subscriberCount())
		t.AddRow([]string{
			md.Link(svc.displayName, path+"subscriptions:svc/"+svc.displayName),
			svc.description,
			humanDuration(svc.renewalPeriod),
			svc.price.String(),
			string(subscriberCount),
		})
		return false
	})
	res += t.String()

	return res
}

func humanDuration(d time.Duration) string {
	seconds := int64(d.Seconds())
	if seconds < 0 {
		seconds = -seconds
	}

	const (
		secPerMinute = 60
		secPerHour   = 60 * secPerMinute
		secPerDay    = 24 * secPerHour
	)

	days := seconds / secPerDay
	seconds %= secPerDay

	hours := seconds / secPerHour
	seconds %= secPerHour

	minutes := seconds / secPerMinute
	seconds %= secPerMinute

	result := ""

	add := func(v int64, unit string) {
		if v > 0 {
			if result != "" {
				result += " "
			}
			result += ufmt.Sprintf("%d%s", v, unit)
		}
	}

	add(days, "d")
	add(hours, "h")
	add(minutes, "m")
	add(seconds, "s")

	if result == "" {
		return "0s"
	}

	return result
}
