package subscriptions

import (
	"chain/runtime"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
	"gno.land/p/sunspirit/table"
)

func Render(path string) string {
	router := mux.NewRouter()

	router.HandleFunc("", homeHandler)
	router.HandleFunc("u/{addr}", userHandler)
	router.HandleFunc("svc/{name}", serviceHandler)

	return router.Render(path)
}

func homeHandler(res *mux.ResponseWriter, _ *mux.Request) {
	out := md.H1("Subscriptions Service Module")
	out += md.Paragraph("This module provides a subscriptions service where users can subscribe to services, manage their subscriptions, and handle payments using native coins or GRC20 tokens.")
	out += md.H2("Services Table")
	t, _ := table.New(
		[]string{"Service Name", "Description", "Renewal Period", "Price", "Subscribers"},
		[][]string{},
	)
	r := runtime.CurrentRealm()
	linkpath := getLinkPath(r.PkgPath())
	services.Iterate("", "", func(key string, value interface{}) bool {
		svc := value.(*service)
		subscriberCount := ufmt.Sprintf("%d", svc.subscriberCount())
		t.AddRow([]string{
			md.Link(svc.displayName, linkpath+":svc/"+svc.displayName),
			svc.description,
			humanDuration(svc.renewalPeriod),
			svc.price.String(),
			string(subscriberCount),
		})
		return false
	})
	out += t.String()
	res.Write(out)
}

func userHandler(res *mux.ResponseWriter, req *mux.Request) {
	userAddress := req.GetVar("addr")
	out := md.H1("User Subscription Details")
	out += md.Paragraph("Details for user: " + userAddress)
	out += md.HorizontalRule()
	t, _ := table.New(
		[]string{"Service Name", "Status", "Balance", "Started At"},
		[][]string{},
	)
	r := runtime.CurrentRealm()
	linkpath := getLinkPath(r.PkgPath())
	users.Iterate("", "", func(key string, value interface{}) bool {
		if key != userAddress {
			return false
		}
		userSubs := value.(*avl.Tree)
		userSubs.Iterate("", "", func(svcName string, _ interface{}) bool {
			svc := mustGetServiceByName(svcName)
			subRaw, _ := svc.subscribers.Get(userAddress)
			sub := subRaw.(*subscription)
			status := "Inactive (insufficient funds)"
			if sub.isActive(svc.renewalPeriod, svc.price) {
				status = "Active"
			}
			t.AddRow([]string{
				md.Link(svcName, linkpath+":svc/"+svcName),
				status,
				sub.balance(svc.renewalPeriod, svc.price).String(),
				sub.startedAt.Format("2006-01-02"),
			})
			return false
		})
		return false
	})
	out += t.String()
	res.Write(out)
}

func serviceHandler(res *mux.ResponseWriter, req *mux.Request) {
	serviceName := req.GetVar("name")
	svc := mustGetServiceByName(serviceName)

	out := md.H1("Service Details: " + serviceName)
	out += md.Paragraph(svc.description)
	out += md.Paragraph("Renewal Period: " + humanDuration(svc.renewalPeriod))
	out += md.Paragraph("Price: " + svc.price.String())
	out += md.Paragraph("Total Balance: " + svc.balance().String())
	out += md.Paragraph("Total Subscribers: " + ufmt.Sprintf("%d", svc.subscriberCount()))
	out += md.HorizontalRule()

	t, _ := table.New(
		[]string{"Subscriber Address", "Status", "Balance", "Started At"},
		[][]string{},
	)
	r := runtime.CurrentRealm()
	linkpath := getLinkPath(r.PkgPath())
	svc.subscribers.Iterate("", "", func(addr string, value interface{}) bool {
		sub := value.(*subscription)
		status := "Inactive (insufficient funds)"
		if sub.isActive(svc.renewalPeriod, svc.price) {
			status = "Active"
		}
		t.AddRow([]string{
			md.Link(addr, linkpath+":u/"+addr),
			status,
			sub.balance(svc.renewalPeriod, svc.price).String(),
			sub.startedAt.Format("2006-01-02"),
		})
		return false
	})
	out += t.String()
	res.Write(out)
}

func getLinkPath(pkgPath string) string {
	slashIdx := strings.IndexRune(pkgPath, '/')
	if slashIdx != 1 {
		return pkgPath[slashIdx:]
	}
	return ""
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
