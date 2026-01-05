package subscriptions

import (
	"chain"
	"chain/runtime"
	"net/url"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
	"gno.land/p/sunspirit/table"
)

const (
	// Pagination constants
	servicesPerPage      = 10
	subscriptionsPerPage = 10
	subscribersPerPage   = 20
)

func Render(path string) string {
	router := mux.NewRouter()

	router.HandleFunc("", homeHandler)
	router.HandleFunc("u", userHandler)
	router.HandleFunc("svc/{name}", serviceHandler)

	return router.Render(path)
}

func homeHandler(res *mux.ResponseWriter, req *mux.Request) {
	out := md.H1("💳 Subscription Services")
	out += md.Paragraph("Manage recurring subscription services with automated payments using native coins and GRC20 tokens.")

	// Action buttons
	r := runtime.CurrentRealm()
	linkpath := getLinkPath(r.PkgPath())
	out += md.Paragraph(md.Link("➕ Create New Service", linkpath+"$help&func=NewServiceSimple") + " • " + md.Link("👤 My Subscriptions", linkpath+":u"))
	out += md.HorizontalRule()

	// Count services
	serviceCount := services.Size()

	if serviceCount == 0 {
		out += md.Blockquote("ℹ️ No services available yet. Create one to get started!")
		res.Write(out)
		return
	}

	// Stats summary
	out += md.H3("📊 Platform Statistics")
	out += md.BulletList([]string{
		md.Bold("Total Services:") + " " + ufmt.Sprintf("%d", serviceCount),
	})
	out += md.HorizontalRule()

	// Setup pager
	pg := pager.NewPager(services, servicesPerPage, true) // reversed order
	page := pg.MustGetPageByPath(req.RawPath)

	out += md.H3("Available Services")
	t, _ := table.New(
		[]string{"Service", "Owner", "Period", "Price", "👥 Subs"},
		[][]string{},
	)

	// Iterate over page items
	for _, item := range page.Items {
		svc := item.Value.(*service)
		svcKey := item.Key
		subscriberCount := svc.subscriberCount()
		t.AddRow([]string{
			md.Bold(md.Link(svc.displayName, linkpath+":svc/"+url.PathEscape(svcKey))) + " - " + md.Italic(svc.description),
			svc.owner.Owner().String(),
			humanDuration(svc.renewalPeriod),
			md.Bold(formatCoinsWithLinks(svc.price)),
			ufmt.Sprintf("%d", subscriberCount),
		})
	}
	out += t.String()

	// Add pager navigation
	picker := page.Picker(req.RawPath)
	if picker != "" {
		out += "\n\n" + picker
	}

	res.Write(out)
}

func userHandler(res *mux.ResponseWriter, req *mux.Request) {
	userAddress := req.Query.Get("addr")
	r := runtime.CurrentRealm()
	linkpath := getLinkPath(r.PkgPath())

	out := md.H1("User Subscriptions")

	// Navigation
	out += md.Paragraph(md.Link("← Back to Services", linkpath))

	// Check if address is provided
	if userAddress == "" {
		// Add form to input address
		out += md.H3("Enter User Address")
		out += md.Paragraph("Enter a Gno address to view their subscriptions:")
		out += "\n<gno-form path=\"u\">\n"
		out += "  <gno-input name=\"addr\" type=\"text\" placeholder=\"g1...\" description=\"Gno Address\" />\n"
		out += "</gno-form>\n"

		res.Write(out)
		return
	}

	out += md.Paragraph("Address: `" + userAddress + "`")
	out += md.HorizontalRule()

	// Get user's subscriptions tree
	userSubsRaw, exists := users.Get(userAddress)
	if !exists {
		out += md.Blockquote("This user has no active subscriptions.")
		res.Write(out)
		return
	}

	userSubs := userSubsRaw.(*avl.Tree)

	// Setup pager for user's subscriptions first
	pg := pager.NewPager(userSubs, subscriptionsPerPage, true) // reversed order
	page := pg.MustGetPageByPath(req.RawPath)

	// Count active subscriptions (only iterate over current page items)
	activeCount := 0
	inactiveCount := 0
	for _, item := range page.Items {
		svcKey := item.Key
		svc := mustGetService(svcKey)
		subRaw, _ := svc.subscribers.Get(userAddress)
		sub := subRaw.(*subscription)
		if sub.isActive(svc.renewalPeriod, svc.price) {
			activeCount++
		} else {
			inactiveCount++
		}
	}

	out += md.H3("Summary")
	out += md.BulletList([]string{
		md.Bold("Active:") + " " + ufmt.Sprintf("%d", activeCount) + " (on this page)",
		md.Bold("Inactive:") + " " + ufmt.Sprintf("%d", inactiveCount) + " (on this page)",
		md.Bold("Total:") + " " + ufmt.Sprintf("%d", userSubs.Size()),
	})
	out += md.HorizontalRule()

	out += md.H3("Subscriptions")
	t, _ := table.New(
		[]string{"Service", "Owner", "Status", "Balance", "Started", "Actions"},
		[][]string{},
	)

	// Iterate over page items
	for _, item := range page.Items {
		svcKey := item.Key
		svc := mustGetService(svcKey)
		subRaw, _ := svc.subscribers.Get(userAddress)
		sub := subRaw.(*subscription)
		status := "Inactive"
		if sub.isActive(svc.renewalPeriod, svc.price) {
			status = "Active"
		}

		// Action links for this subscription
		actions := md.Link("Top Up", linkpath+"$help&func=Topup&svcKey="+url.QueryEscape(svcKey)) + " • " +
			md.Link("Withdraw", linkpath+"$help&func=Withdraw&svcKey="+url.QueryEscape(svcKey)) + " • " +
			md.Link("Unsubscribe", linkpath+"$help&func=Unsubscribe&svcKey="+url.QueryEscape(svcKey))

		t.AddRow([]string{
			md.Bold(md.Link(svc.displayName, linkpath+":svc/"+url.PathEscape(svcKey))),
			svc.owner.Owner().String(),
			status,
			md.Bold(formatCoinsWithLinks(sub.balance(svc.renewalPeriod, svc.price))),
			sub.startedAt.Format("Jan 02, 2006"),
			actions,
		})
	}
	out += t.String()

	// Add pager navigation
	picker := page.Picker(req.RawPath)
	if picker != "" {
		out += "\n\n" + picker
	}
	res.Write(out)
}

func serviceHandler(res *mux.ResponseWriter, req *mux.Request) {
	svcKey := req.GetVar("name")
	svc := mustGetService(svcKey)
	r := runtime.CurrentRealm()
	linkpath := getLinkPath(r.PkgPath())

	out := md.H1(svc.displayName)
	out += md.Paragraph(svc.description)

	// Navigation
	out += md.Paragraph(md.Link("← Back to Services", linkpath))
	out += md.HorizontalRule()

	// Action buttons
	out += md.H3("Quick Actions")
	actionLinks := []string{
		md.Link("Subscribe", linkpath+"$help&func=Subscribe&svcKey="+url.QueryEscape(svcKey)),
		md.Link("Top Up", linkpath+"$help&func=Topup&svcKey="+url.QueryEscape(svcKey)),
		md.Link("Unsubscribe", linkpath+"$help&func=Unsubscribe&svcKey="+url.QueryEscape(svcKey)),
	}
	// Add owner-only action
	caller := runtime.PreviousRealm().Address()
	if caller == svc.owner.Owner() {
		actionLinks = append(actionLinks, md.Link("💸 Claim Vault (Owner)", linkpath+"$help&func=ServiceClaimVault&svcKey="+url.QueryEscape(svcKey)))
	}
	out += md.Paragraph(strings.Join(actionLinks, " • "))
	out += md.HorizontalRule()

	// Service details in a nice format
	out += md.H3("Service Information")
	subscriberCount := svc.subscriberCount()

	out += md.BulletList([]string{
		md.Bold("Owner:") + " " + svc.owner.Owner().String(),
		md.Bold("Billing Period:") + " " + humanDuration(svc.renewalPeriod),
		md.Bold("Price:") + " " + formatCoinsWithLinks(svc.price),
		md.Bold("Total Subscribers:") + " " + ufmt.Sprintf("%d", subscriberCount),
	})

	// Add GRC20 payment instructions if service uses GRC20 tokens
	hasGRC20 := false
	var grc20Paths []string
	for _, coin := range svc.price {
		if strings.HasPrefix(coin.Denom, "/grc20/") {
			hasGRC20 = true
			// Extract the realm path from /grc20//r/path/to/token
			path := strings.TrimPrefix(coin.Denom, "/grc20/")
			grc20Paths = append(grc20Paths, path)
		}
	}

	if hasGRC20 {
		out += "\n\n"
		out += md.HorizontalRule()
		out += md.H4("💡 How to Pay with GRC20 Tokens")
		out += md.Blockquote("ℹ️ **Important:** To subscribe or top up with GRC20 tokens, you must first approve this realm to spend your tokens.")
		out += md.Bold("Tip:") + "You can find a complete example workflow in the " + md.Link("`internal/`", "/r/samcrew/subscriptions/internal/grc20_test_script/test_grc20.gno") + " folder." + "\n\n"
		out += md.Bold("Steps:") + "\n"
		for i, tokenPath := range grc20Paths {
			out += md.OrderedList([]string{
				"Visit the GRC20 token: " + md.Link(tokenPath, tokenPath),
				"Call the " + md.InlineCode("Approve") + " function with:\n   - " + md.Bold("spender:") + " " + md.InlineCode(runtime.CurrentRealm().Address().String()) + "\n   - " + md.Bold("amount:") + " the amount you want to deposit (or a large value for multiple top-ups)",
				"Return here and use " + md.Link("Subscribe", linkpath+"$help&func=Subscribe&svcKey="+url.QueryEscape(svcKey)) + " or " + md.Link("Top Up", linkpath+"$help&func=Topup&svcKey="+url.QueryEscape(svcKey)) + " with the GRC20 token details",
			})
			if i < len(grc20Paths)-1 {
				out += md.HorizontalRule()
			}
		}
	}

	out += md.HorizontalRule()

	// Subscribers table
	if subscriberCount == 0 {
		out += md.Blockquote("ℹ️ No active subscribers yet.")
	} else {
		// Setup pager for subscribers
		subPg := pager.NewPager(svc.subscribers, subscribersPerPage, true) // reversed order
		subPage := subPg.MustGetPageByPath(req.RawPath)

		startIdx := (subPage.PageNumber-1)*subPage.PageSize + 1
		endIdx := startIdx + len(subPage.Items) - 1

		out += md.H3("Subscribers")
		out += md.Paragraph(md.Italic(ufmt.Sprintf("Showing %d-%d of %d total subscribers",
			startIdx, endIdx, subPage.TotalItems)))

		t, _ := table.New(
			[]string{"Address", "Status", "💰 Balance", "📅 Joined"},
			[][]string{},
		)

		// Count active vs inactive (only on current page)
		activeCount := 0
		inactiveCount := 0

		for _, item := range subPage.Items {
			addr := item.Key
			sub := item.Value.(*subscription)

			status := "🔴 Inactive"
			if sub.isActive(svc.renewalPeriod, svc.price) {
				status = "🟢 Active"
				activeCount++
			} else {
				inactiveCount++
			}

			t.AddRow([]string{
				md.Link(addr, linkpath+":u?addr="+addr),
				status,
				formatCoinsWithLinks(sub.balance(svc.renewalPeriod, svc.price)),
				sub.startedAt.Format("Jan 02, 2006"),
			})
		}

		// Show stats for current page
		out += ufmt.Sprintf("On this page: 🟢 %d active • 🔴 %d inactive",
			activeCount, inactiveCount)

		out += t.String()

		// Add pager navigation
		picker := subPage.Picker(req.RawPath)
		if picker != "" {
			out += "\n\n" + picker
		}
	}
	res.Write(out)
}

func getLinkPath(pkgPath string) string {
	slashIdx := strings.IndexRune(pkgPath, '/')
	if slashIdx != 1 {
		return pkgPath[slashIdx:]
	}
	return ""
}

// formatCoinsWithLinks formats coins and adds links to GRC20 tokens
func formatCoinsWithLinks(coins chain.Coins) string {
	if len(coins) == 0 {
		return "0"
	}

	// Calculate total for each coin type
	totals := make(map[string]int64)
	for _, coin := range coins {
		totals[coin.Denom] += coin.Amount
	}

	result := ""
	first := true
	for denom, amount := range totals {
		if !first {
			result += ", "
		}
		first = false

		// Check if it's a GRC20 token
		if strings.HasPrefix(denom, "/grc20/") {
			// Extract the path after /grc20/ (e.g., "/r/demo/defi/foo20")
			displayPath := strings.TrimPrefix(denom, "/grc20/")
			// Format: "5000000/grc20 /r/demo/defi/foo20"
			result += ufmt.Sprintf("%d/grc20 ", amount) + md.Link(displayPath, displayPath)
		} else if strings.HasPrefix(denom, "/native/") {
			// For native coins: "1000000ugnot"
			cleanDenom := strings.TrimPrefix(denom, "/native/")
			result += ufmt.Sprintf("%d%s", amount, cleanDenom)
		} else {
			// Fallback for unexpected format
			result += ufmt.Sprintf("%d%s", amount, denom)
		}
	}

	return result
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
