package grants

import (
	"chain"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gno.land/p/moul/txlink/v0"
	gp "gno.land/p/nt/grants/v0"
	"gno.land/p/nt/markdown/sanitize/v0"
)

const (
	pageSize    = 25
	lastReports = 3
)

// Render serves the desk at "", active grants at "active/<n>", requests
// waiting for GovDAO at "waiting/<n>", every grant at "all/<n>", one grant at "grant/<id>" (its reports at
// "grant/<id>/reports"), and every grant under a namespace at "ns/<name>".
//
// Pages use gnoweb's extensions (cards, alerts, collapsible sections, icons
// and call forms). Everything a grantee wrote is escaped line by line and
// quoted, so none of it can take the shape of the desk's own markup.
func Render(p string) string {
	parts := strings.Split(p, "/")
	switch {
	case p == "":
		return home()
	case len(parts) == 2 && parts[0] == "active":
		if n := pageNum(parts[1]); n > 0 {
			return listPage("Active grants", "active", n, desk.ActiveCount(), desk.Active((n-1)*pageSize, pageSize))
		}
	case len(parts) == 2 && parts[0] == "waiting":
		if n := pageNum(parts[1]); n > 0 {
			return listPage("Waiting for GovDAO, oldest first", "waiting", n, desk.WaitingCount(), desk.Waiting((n-1)*pageSize, pageSize))
		}
	case len(parts) == 2 && parts[0] == "all":
		if n := pageNum(parts[1]); n > 0 {
			return listPage("Every grant, newest first", "all", n, desk.Size(), desk.Page((n-1)*pageSize, pageSize))
		}
	case parts[0] == "grant" && (len(parts) == 2 || len(parts) == 3 && parts[2] == "reports"):
		if id, err := strconv.Atoi(parts[1]); err == nil {
			if g, ok := desk.Get(id); ok {
				if len(parts) == 3 {
					return reportsPage(g)
				}
				return grantPage(g)
			}
		}
	case len(parts) == 2 && parts[0] == "ns" && isName(parts[1]):
		return nsPage(parts[1])
	}
	return notFound
}

const notFound = "# Not found\n\nNo such page.\n\n<gno-button href=\"/r/gov/grants/v0\" label=\"Back to the desk\" />\n"

// ----- desk -----

func home() string {
	r := desk.Rules()
	active := desk.Active(0, pageSize)
	waiting := desk.Waiting(0, pageSize)
	var b strings.Builder
	b.WriteString("# <gno-icon name=\"coins\" /> GovDAO grants for realms\n\n")
	b.WriteString("Projects, teams and DAOs ask GovDAO for funds **as code, not as people**: from an account in " +
		"[the grant wallet](" + realmLink(walletPath) + "), or from a realm of their own. One GovDAO vote approves a grant; " +
		"each payment then waits for a public report on what it pays for and " + days(r.Window) +
		" in which GovDAO can stop it, and only ever goes to the grantee's own address.\n\n")
	if desk.Closed() {
		b.WriteString("> [!CAUTION] Closed\n> GovDAO closed this desk: it takes no new request or approval. " +
			"Grants already running are paid as approved unless GovDAO cancels them.\n\n")
	}

	b.WriteString(cards(
		card("wallet", "Pool", gnot(Balance())),
		card("lock", "Committed", gnot(Committed())),
		card("coins", "Available", gnot(Available())),
		card("play", "Active grants", itoa(desk.ActiveCount())),
	))
	b.WriteString("Pool address: " + sanitize.InlineCode(Address().String()) + "\n\n")

	// Every active grant: only GovDAO makes one, so this stays bounded.
	attention(&b, append(desk.Active(0, desk.ActiveCount()), waiting...))

	b.WriteString("## <gno-icon name=\"play\" /> Active grants\n\n")
	table(&b, active, true)
	if desk.ActiveCount() > pageSize {
		b.WriteString("<gno-button href=\"" + link("active/1") + "\" label=\"All " + itoa(desk.ActiveCount()) + " active grants\" />\n\n")
	}
	b.WriteString("## <gno-icon name=\"clock\" /> Waiting for GovDAO\n\n")
	table(&b, waiting, true)
	if desk.WaitingCount() > pageSize {
		b.WriteString("<gno-button href=\"" + link("waiting/1") + "\" label=\"All " + itoa(desk.WaitingCount()) + " requests\" />\n\n")
	}
	b.WriteString("## <gno-icon name=\"files-history\" /> Latest grants\n\n")
	table(&b, desk.Page(0, 5), false)
	b.WriteString("<gno-button href=\"" + link("all/1") + "\" label=\"Every grant\" variant=\"outline\" />\n\n")

	if !desk.Closed() {
		howToApply(&b)
	}

	b.WriteString("<gno-details summary=\"How GovDAO decides\">\n\n")
	b.WriteString("Members sponsor, cancel or freeze with the MsgRun printed on each grant's page; a sponsor states what they " +
		"checked, and that review stays on the grant. MsgRun needs an allowlisted " +
		"account: if you have none, share the page with a member who does. Once a vote passes, anyone executes it with " +
		"[ExecuteProposal](" + txlink.Realm("gno.land/r/gov/dao").Call("ExecuteProposal") + ").\n\n")
	b.WriteString("The pool is funded by sending GNOT to its address: by whoever holds the Ecosystem Treasury account, " +
		"or, when the GovDAO treasury holds funds, with a treasury payment proposal:\n\n")
	b.WriteString(script("\"chain\"\n\n\t\"gno.land/p/nt/treasury/v0\"\n\t\"gno.land/r/gov/dao\"\n\t\"gno.land/r/gov/dao/impl/v0\"",
		"pay := treasury.NewCoinsPayment(chain.NewCoins(chain.NewCoin(\"ugnot\", UGNOT_AMOUNT)), address(\""+Address().String()+"\"))\n"+
			"\tdao.MustCreateProposal(cross(cur), impl.NewTreasuryPaymentRequest(cross(cur), pay, \"Fund the grants pool\"))"))
	b.WriteString("\nThe same MsgRun shape files the desk's other proposals: " +
		"`grants.NewReturnRequest(cross(cur), address(\"TO\"), UGNOT_AMOUNT, \"REASON\")` sends uncommitted funds back where they came from, " +
		"`grants.NewCloseRequest(cross(cur), address(\"TO\"), \"REASON\", false)` closes the desk (`true` also cancels the running grants), " +
		"and `grants.NewFreezeRequest(cross(cur), \"GRANTEE\", false, \"REASON\")` unfreezes a grantee.\n")
	b.WriteString("\n</gno-details>\n\n")

	b.WriteString("<gno-details summary=\"Rules\">\n\n")
	b.WriteString("- A realm has one pending request at a time; a new one supersedes it. A request nobody approves within " + days(r.Expiry) + " expires.\n")
	b.WriteString("- A grant is approved only if the pool covers all of it on top of what is already committed.\n")
	b.WriteString("- Every payment needs a report on what it pays for, then " + days(r.Window) + " in which GovDAO can object, and at least " +
		"the grant's period (" + days(r.MinPeriod) + " or more) after the previous payment. The periods add up to a year at most; " +
		"late reports push payments later.\n")
	b.WriteString("- The one exception is an advance: a first payment of at most " + itoa(r.Advance) + "% of the grant is due " +
		days(r.Window) + " after approval, without a report.\n")
	b.WriteString("- GovDAO can cancel at any time; paid payments stay paid. The member who files a cancellation can hold the next payment " +
		days(r.HoldFor) + " more, once per payment. GovDAO can also freeze a grantee: no payment reaches it and a wallet account cannot spend.\n")
	b.WriteString("- The grantee can withdraw a request or a grant; what it was not paid goes back to the pool.\n")
	b.WriteString("- A payment left unreported for " + days(r.Expiry) + " after its period lets anyone expire the grant and free its reservation.\n")
	b.WriteString("- GovDAO can close the desk in one vote: no new request or approval, running grants go on (or are cancelled, if the vote says so), " +
		"and the uncommitted pool goes back to the treasury.\n\n")
	b.WriteString("</gno-details>\n")
	return b.String()
}

func howToApply(b *strings.Builder) {
	b.WriteString("<gno-details summary=\"How to apply\">\n\n")
	b.WriteString("**The simple way: [open an account](" + realmLink(walletPath) + ") in the grant wallet** named after a " +
		"namespace you can deploy under: your registered name, or your g1 address (`gnogolf`, or `gnogolf/cup` for a second " +
		"project, once that name is yours). Then apply, report and spend from its page. Every account runs the same code, " +
		"so voters only need to know who owns yours.\n\n")
	b.WriteString("**Or apply from a realm of your own**, for instance `r/gnogolf/grant`, when the grant should follow your own " +
		"rules. GovDAO members must then read its code before voting:\n\n")
	b.WriteString("```go\nid := grants.Apply(cross(cur), \"Gnogolf v3\", \"Why this grows gno.land\", []grants.Tranche{\n" +
		"\t{Deliverable: \"v3 design published\", Amount: 1_000_000_000}, // ugnot; at most " + itoa(desk.Rules().Advance) +
		"% of the total: an advance, paid without a report\n" +
		"\t{Deliverable: \"v3 live, 100 weekly players\", Amount: 4_000_000_000},\n" +
		"}, 30*24*time.Hour)\n\n" +
		"// Later, in its own transaction, once the grant is approved and the advance claimed:\n" +
		"grants.Report(cross(cur), id, \"Done: numbers, what is next\", \"/r/gnogolf/game:v3\")\n```\n\n")
	b.WriteString("Payments are not pushed: once one is due, anyone (the grantee, a bot) releases it with the Claim form on the grant's page.\n\n")
	b.WriteString("</gno-details>\n\n")
}

// attention lists what someone can act on right now: a payment to release,
// a grant to expire, a report GovDAO can still object to.
func attention(b *strings.Builder, open []gp.Grant) {
	var items []string
	for _, g := range open {
		label := "[#" + itoa(g.ID) + " " + sanitize.InlineText(g.Title) + "](" + link("grant/"+itoa(g.ID)) + ")"
		switch {
		case due(g):
			items = append(items, label+": payment "+itoa(len(g.PaidAt)+1)+" can be released, [Claim]("+tx("Claim", g.ID)+")")
		case expirable(g) && g.Status == gp.Active:
			items = append(items, label+": left unattended, [Expire]("+tx("Expire", g.ID)+") frees its reservation")
		case expirable(g):
			items = append(items, label+": nobody approved it in time, [Expire]("+tx("Expire", g.ID)+") closes it")
		case g.Status == gp.Active && g.ReportedAt != 0 && !Frozen(g.Grantee):
			if at, err := desk.PayableAt(g.ID); err == nil {
				items = append(items, label+": report on payment "+itoa(len(g.PaidAt)+1)+" to read, pays "+date(at.Unix())+
					" unless GovDAO objects")
			}
		}
	}
	if len(items) == 0 {
		return
	}
	b.WriteString("> [!WARNING] Needs attention\n")
	for _, it := range items {
		b.WriteString("> - " + it + "\n")
	}
	b.WriteString("\n")
}

func listPage(title, route string, n, total int, list []gp.Grant) string {
	if len(list) == 0 && n > 1 {
		return notFound
	}
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	table(&b, list, route != "all")
	if n > 1 {
		b.WriteString("<gno-button href=\"" + link(route+"/"+itoa(n-1)) + "\" label=\"Previous\" variant=\"outline\" /> ")
	}
	if n*pageSize < total {
		b.WriteString("<gno-button href=\"" + link(route+"/"+itoa(n+1)) + "\" label=\"Next\" variant=\"outline\" />")
	}
	b.WriteString("\n\n<gno-button href=\"" + link("") + "\" label=\"Back to the desk\" />\n")
	return b.String()
}

func table(b *strings.Builder, list []gp.Grant, next bool) {
	if len(list) == 0 {
		b.WriteString("None.\n\n")
		return
	}
	last := "Status"
	if next {
		last = "Next step"
	}
	b.WriteString("| # | Grant | Grantee | Paid | " + last + " |\n| --- | --- | --- | --- | --- |\n")
	for _, g := range list {
		state := badge(g)
		if next {
			state = nextStep(g)
		}
		b.WriteString("| " + itoa(g.ID) +
			" | [" + sanitize.TableCell(g.Title) + "](" + link("grant/"+itoa(g.ID)) + ")" +
			" | " + sanitize.InlineCode(short(g.Grantee)) +
			" | " + progress(g) + " " + gnot(g.Paid()) + " of " + gnot(g.Total()) +
			" | " + state + " |\n")
	}
	b.WriteString("\n")
}

// ----- one grant -----

func grantPage(g gp.Grant) string {
	r := desk.Rules()
	var b strings.Builder
	b.WriteString("# <gno-icon name=\"" + icon(g) + "\" /> Grant #" + itoa(g.ID) + ": " + sanitize.InlineText(g.Title) + "\n\n")
	b.WriteString(badge(g) + " · " + realmLinks(g.Grantee) +
		" · namespace [" + namespace(g.Grantee) + "](" + link("ns/"+namespace(g.Grantee)) + ")\n\n")

	if Frozen(g.Grantee) {
		why, _ := frozen.Get(g.Grantee).(string)
		b.WriteString("> [!CAUTION] Grantee frozen by GovDAO\n> No payment reaches it until GovDAO unfreezes it. Reason: " + escape(why) + "\n\n")
	}
	status(&b, g)

	next := "—"
	if at, err := desk.PayableAt(g.ID); err == nil {
		next = when(at)
	} else if g.Status == gp.Active {
		next = "after a report"
	}
	b.WriteString(cards(
		card("coins", "Total", gnot(g.Total())),
		card("check-circle", "Paid", gnot(g.Paid())+"  \n"+progress(g)),
		card("calendar", "Next payment", next),
		card("wallet", "Paid to", sanitize.InlineCode(short(g.Grantee))),
	))
	b.WriteString("Payouts go to " + sanitize.InlineCode(chain.PackageAddress(g.Grantee).String()) +
		", the grantee's own address. Where they go next is on [its page](" + granteePage(g.Grantee) + ").\n\n")

	b.WriteString("## <gno-icon name=\"list\" /> Payments\n\n| # | Deliverable, in the grantee's words | Amount | Due | State |\n| --- | --- | --- | --- | --- |\n")
	for i, t := range g.Tranches {
		due := "report + " + days(r.Window)
		if i == 0 && g.Advance {
			due = "advance, " + days(r.Window) + " after approval"
		} else if i > 0 {
			due += ", ≥ " + days(g.Period) + " after payment " + itoa(i)
		}
		state := "<gno-icon name=\"circle\" /> to come"
		if i < len(g.PaidAt) {
			state = "<gno-icon name=\"check-circle\" /> paid " + date(g.PaidAt[i])
		} else if g.Status == gp.Active && i == len(g.PaidAt) {
			state = "<gno-icon name=\"arrow-right-circle\" /> next"
		} else if g.Status != gp.Pending && g.Status != gp.Active {
			state = "<gno-icon name=\"close-circle\" /> not paid"
		}
		b.WriteString("| " + itoa(i+1) + " | " + sanitize.TableCell(t.Deliverable) + " | " + gnot(t.Amount) +
			" | " + due + " | " + state + " |\n")
	}

	b.WriteString("\n## <gno-icon name=\"timeline\" /> Timeline\n\n")
	timeline(&b, g)

	b.WriteString("## <gno-icon name=\"document\" /> Pitch\n\n_Written by the grantee:_\n\n" + quote(g.Pitch))
	if g.ApprovedAt != 0 && g.Review != "" {
		b.WriteString("## <gno-icon name=\"shield-check\" /> Sponsor's review\n\n_Written by " + sanitize.InlineCode(g.Sponsor) +
			", who built the approval request a GovDAO member filed:_\n\n" + quote(g.Review))
	}

	b.WriteString("## <gno-icon name=\"chat\" /> Reports\n\n")
	reports(&b, g, lastReports)
	if len(g.Reports) > lastReports {
		b.WriteString("<gno-button href=\"" + link("grant/"+itoa(g.ID)+"/reports") + "\" label=\"All " +
			itoa(len(g.Reports)) + " reports\" variant=\"outline\" />\n\n")
	}

	members(&b, g)
	return b.String()
}

// status opens a grant page with what it is waiting for, and the form for
// the call anyone can make when one is due.
func status(b *strings.Builder, g gp.Grant) {
	n := itoa(len(g.PaidAt) + 1)
	switch g.Status {
	case gp.Pending:
		if expirable(g) {
			b.WriteString("> [!CAUTION] Expired unapproved\n> Nobody approved it within " + days(desk.Rules().Expiry) + ". Anyone can close it:\n\n")
			b.WriteString(form("Expire", g.ID))
			return
		}
		if desk.Closed() {
			b.WriteString("> [!CAUTION] Desk closed\n> GovDAO closed the desk: this request can no longer be approved. It expires on " +
				when(desk.ExpiresAt(g)) + ".\n\n")
			return
		}
		if hostFrozen(g.Grantee) {
			b.WriteString("> [!CAUTION] Grantee frozen\n> GovDAO froze the grantee or its host: this request cannot be approved " +
				"until it is unfrozen. It expires on " + when(desk.ExpiresAt(g)) + ".\n\n")
			return
		}
		b.WriteString("> [!INFO] Waiting for GovDAO\n> A GovDAO member sponsors it and GovDAO approves it before " +
			when(desk.ExpiresAt(g)) + ", or it expires. Once sponsored, its vote is in [GovDAO's proposals](/r/gov/dao), titled " +
			"_" + escape(approveTitle(g)) + "_; check that its executor was created in `" + path + "`.\n\n")
	case gp.Active:
		at, err := desk.PayableAt(g.ID)
		switch {
		case err == gp.ErrNotReported && expirable(g):
			b.WriteString("> [!CAUTION] Payment " + n + " never reported\n> The grantee did not report on it in time. " +
				"Anyone can expire the grant, which returns what it was not paid to the pool:\n\n")
			b.WriteString(form("Expire", g.ID))
		case err == gp.ErrNotReported:
			// As payableAt: the report plus the window, no sooner than the
			// period after the last payment, and a hold adds HoldFor after
			// whichever is later, that or when it was placed.
			wait := desk.Rules().Window
			var earliest time.Time
			if len(g.PaidAt) > 0 {
				earliest = time.Unix(g.PaidAt[len(g.PaidAt)-1], 0).Add(g.Period)
			}
			if g.Held() {
				wait += desk.Rules().HoldFor
				if h := time.Unix(g.Holds[len(g.Holds)-1].At, 0); h.After(earliest) {
					earliest = h
				}
				earliest = earliest.Add(desk.Rules().HoldFor)
			}
			b.WriteString("> [!NOTE] Waiting for the grantee's report on payment " + n + "\n> The report is due before " +
				when(desk.ExpiresAt(g)) + ". Payment follows " + days(wait) + " after the report")
			if !earliest.IsZero() {
				b.WriteString(", and no sooner than " + when(earliest))
			}
			b.WriteString(", unless GovDAO objects.\n\n")
		case !time.Now().Before(at) && Frozen(g.Grantee):
			b.WriteString("> [!NOTE] Payment " + n + " is due and waits\n> GovDAO froze the grantee; the payment can be released once it is unfrozen.\n\n")
		case !time.Now().Before(at):
			b.WriteString("> [!SUCCESS] Payment " + n + " can be released\n> " + gnot(g.Tranches[len(g.PaidAt)].Amount) +
				" to the grantee realm. Anyone can do it; the coins can only go there.\n\n")
			b.WriteString(form("Claim", g.ID))
		default:
			b.WriteString("> [!NOTE] Payment " + n + " due " + when(at) + "\n> " + gnot(g.Tranches[len(g.PaidAt)].Amount) +
				" to the grantee realm, unless GovDAO cancels before.\n\n")
		}
		if g.Held() && !due(g) {
			ref := i64(g.Holds[len(g.Holds)-1].Ref)
			b.WriteString("> [!WARNING] Held\n> A GovDAO member filed [proposal #" + ref + "](/r/gov/dao:" + ref +
				") to cancel this grant and held payment " + n + " " + days(desk.Rules().HoldFor) + " more while GovDAO votes.\n\n")
		}
	case gp.Completed:
		b.WriteString("> [!SUCCESS] Paid in full\n> Every payment was made, each after the report it asked for.\n\n")
	default:
		why := escape(g.Reason)
		if g.Status == gp.Expired {
			why = "Left unattended past its deadline; what it was not paid went back to the pool."
			if g.ApprovedAt == 0 {
				why = "GovDAO did not approve it in time; nothing was reserved for it."
			}
		}
		b.WriteString("> [!CAUTION] " + title(Label(g)) + "\n> " + why + "\n\n")
	}
}

// timeline lists everything that happened to a grant, oldest first.
func timeline(b *strings.Builder, g gp.Grant) {
	type event struct {
		at   int64
		what string
	}
	evs := []event{{g.AppliedAt, "<gno-icon name=\"create\" /> Requested by the grantee"}}
	if g.ApprovedAt != 0 {
		by := ""
		if g.Sponsor != "" {
			by = ", sponsored by " + sanitize.InlineCode(g.Sponsor)
		}
		evs = append(evs, event{g.ApprovedAt, "<gno-icon name=\"check\" /> Approved by GovDAO" + by + ", " + gnot(g.Total()) + " reserved"})
	}
	for i, t := range g.PaidAt {
		evs = append(evs, event{t, "<gno-icon name=\"coins\" /> Payment " + itoa(i+1) + " released: " + gnot(g.Tranches[i].Amount)})
	}
	for _, rp := range g.Reports {
		evs = append(evs, event{rp.At, "<gno-icon name=\"chat\" /> Report on payment " + itoa(rp.Tranche+1)})
	}
	for _, h := range g.Holds {
		evs = append(evs, event{h.At, "<gno-icon name=\"lock\" /> Payment " + itoa(h.Tranche+1) +
			" held by [proposal #" + i64(h.Ref) + "](/r/gov/dao:" + i64(h.Ref) + ")"})
	}
	if g.ClosedAt != 0 {
		evs = append(evs, event{g.ClosedAt, "<gno-icon name=\"" + icon(g) + "\" /> " + title(Label(g))})
	}
	// Insertion sort: a few dozen events at most, mostly in order already.
	for i := 1; i < len(evs); i++ {
		for j := i; j > 0 && evs[j].at < evs[j-1].at; j-- {
			evs[j], evs[j-1] = evs[j-1], evs[j]
		}
	}
	b.WriteString("| When | What |\n| --- | --- |\n")
	for _, e := range evs {
		b.WriteString("| " + whenU(e.at) + " | " + e.what + " |\n")
	}
	b.WriteString("\n")
}

func members(b *strings.Builder, g gp.Grant) {
	r := desk.Rules()
	open := ""
	switch g.Status {
	case gp.Pending:
		if expirable(g) || desk.Closed() {
			return
		}
		open = " open"
	case gp.Active:
	default:
		return
	}
	b.WriteString("## <gno-icon name=\"users\" /> For GovDAO members\n\n<gno-details summary=\"Scripts and review checklist\"" + open + ">\n\n")
	if g.Status == gp.Pending && hostFrozen(g.Grantee) {
		b.WriteString("GovDAO froze this grantee or its host: it cannot be approved until GovDAO unfreezes it.\n\n")
	} else if g.Status == gp.Pending {
		b.WriteString("Not sponsored yet? Read the checklist below, then file it with this MsgRun; your review is shown to voters " +
			"and stays on the grant:\n\n")
		b.WriteString(script(daoImports, "dao.MustCreateProposal(cross(cur), grants.NewApproveRequest(cross(cur), "+itoa(g.ID)+
			", \"REVIEW: who controls the grantee and how you know them; what you checked; any conflict of interest\"))"))
	} else {
		body := "dao.MustCreateProposal(cross(cur), grants.NewCancelRequest(cross(cur), " + itoa(g.ID) + ", \"REASON\"))"
		if g.Held() {
			b.WriteString("Cancel it with this MsgRun (payment " + itoa(len(g.PaidAt)+1) + " is already held):\n\n")
		} else {
			b.WriteString("Cancel it with this MsgRun, which also holds the next payment " + days(r.HoldFor) + " more while GovDAO votes:\n\n")
			body = "pid := " + body + "\n\tgrants.Hold(cross(cur), " + itoa(g.ID) + ", int64(pid))"
		}
		b.WriteString(script(daoImports, body))
	}
	if Frozen(g.Grantee) {
		b.WriteString("\nUnfreeze the grantee with:\n\n")
	} else {
		b.WriteString("\nFreeze the grantee (no payment reaches it, it cannot apply, a wallet account cannot spend) with:\n\n")
	}
	b.WriteString(script(daoImports, "dao.MustCreateProposal(cross(cur), grants.NewFreezeRequest(cross(cur), "+
		strconv.Quote(g.Grantee)+", "+strconv.FormatBool(!Frozen(g.Grantee))+", \"REASON\"))"))
	b.WriteString(checklist(g))
	b.WriteString("\n</gno-details>\n")
}

func reportsPage(g gp.Grant) string {
	var b strings.Builder
	b.WriteString("# Reports on grant #" + itoa(g.ID) + ": " + sanitize.InlineText(g.Title) + "\n\n")
	reports(&b, g, len(g.Reports))
	b.WriteString("<gno-button href=\"" + link("grant/"+itoa(g.ID)) + "\" label=\"Back to the grant\" />\n")
	return b.String()
}

// reports writes the latest n reports, newest first and open, each in its
// own collapsible section: they are the grantee's words, quoted.
func reports(b *strings.Builder, g gp.Grant, n int) {
	if len(g.Reports) == 0 {
		b.WriteString("None yet.\n\n")
		return
	}
	for i := len(g.Reports) - 1; i >= 0 && i >= len(g.Reports)-n; i-- {
		rp := g.Reports[i]
		open := ""
		if i == len(g.Reports)-1 {
			open = " open"
		}
		b.WriteString("<gno-details summary=\"On payment " + itoa(rp.Tranche+1) + ", " + whenU(rp.At) + "\"" + open + ">\n\n" +
			"_Written by the grantee:_\n\n" + quote(rp.Body))
		if u := sanitize.URL(rp.Link); u != "" {
			b.WriteString("<gno-icon name=\"link\" /> What it shows: [" + escape(rp.Link) + "](" + u + ")\n\n")
		}
		b.WriteString("</gno-details>\n\n")
	}
}

// nsPage lists a namespace's grants: those of its wallet accounts, then those
// of its own realms.
func nsPage(ns string) string {
	list, total := nsGrants(ns, 100)
	var open, paid int64
	for _, g := range list {
		if g.Status == gp.Active {
			open += g.Total() - g.Paid()
		}
		paid += g.Paid()
	}
	var b strings.Builder
	b.WriteString("# <gno-icon name=\"users\" /> Grants to the " + sanitize.InlineCode(ns) + " namespace\n\n")
	b.WriteString("Its [wallet accounts](" + realmLink(walletPath) + ") and its realms under " + sanitize.InlineCode("gno.land/r/"+ns) + ".\n\n")
	b.WriteString(cards(
		card("list", "Grants filed", count(total)),
		card("coins", "Still to pay, approved grants", gnot(open)),
		card("check-circle", "Paid", gnot(paid)),
	))
	if total > len(list) {
		b.WriteString("The first " + itoa(len(list)) + " by path, totals over those.\n\n")
	}
	table(&b, list, false)
	return b.String()
}

// nsGrants returns up to max grants of a namespace, its wallet accounts then
// its own realms, and how many there are, counting up to gp.MaxListed. The
// wallet lives under r/gov: its accounts never count as the gov namespace's
// own.
func nsGrants(ns string, max int) ([]gp.Grant, int) {
	list, w := desk.ByGrantee(walletPath+"#"+ns, 0, max)
	own, n := desk.ByGranteeExcept("gno.land/r/"+ns, walletPath, 0, max-len(list))
	total := w + n
	if total > gp.MaxListed { // each count stops at the cap, so a sum at the cap is "or more"
		total = gp.MaxListed
	}
	return append(list, own...), total
}

// ----- proposal texts -----

// approveText is what GovDAO members read before voting on an approval.
// Facts the desk computes come first; the sponsor's and the grantee's words
// come last or in their own column, escaped and clamped, so they cannot pose
// as terms.
func approveText(g gp.Grant, sponsor, review string) string {
	r := desk.Rules()
	var b strings.Builder
	b.WriteString("If this passes, " + sanitize.InlineCode(g.Grantee) + " can receive up to **" + gnot(g.Total()) +
		"** in " + payments(len(g.Tranches)) + ": ")
	if g.Advance {
		b.WriteString("the first, an advance, " + days(r.Window) + " after this executes")
		if len(g.Tranches) > 1 {
			b.WriteString("; each later one only after a public report, " + days(r.Window) + " to object, and at least " +
				days(g.Period) + " after the previous one")
		}
	} else {
		b.WriteString("each only after a public report on it and " + days(r.Window) + " to object")
		if len(g.Tranches) > 1 {
			b.WriteString(", at least " + days(g.Period) + " apart")
		}
	}
	b.WriteString(". No further vote is needed unless GovDAO cancels or freezes it.\n\n")
	code := "a realm of its own, " + realmLinks(g.Grantee) + ": **read it before voting**"
	if hosted(g.Grantee) {
		code = "an account in the [grant wallet](" + granteePage(g.Grantee) + "), whose code every account shares ([source](" +
			realmLink(walletPath) + "$source)); check its owner there (an owner can hand the account over; every change is in its journal)"
	}
	b.WriteString("- Grantee: " + code + ". Payouts go to " + sanitize.InlineCode(chain.PackageAddress(g.Grantee).String()) + ".\n")
	b.WriteString("- Sponsor: " + sanitize.InlineCode(sponsor) + ", whose review is below.\n")
	_, n := nsGrants(namespace(g.Grantee), 0)
	b.WriteString("- Grants filed by this namespace, this one included: " + count(n) +
		" ([list](" + link("ns/"+namespace(g.Grantee)) + ")).\n")
	b.WriteString("- Pool when proposed: " + gnot(Available()) + " available of " + gnot(Balance()) + ".")
	if Available() < g.Total() {
		b.WriteString(" **Short by " + gnot(g.Total()-Available()) + ": execution fails until the pool is funded.**")
	}
	b.WriteString("\n- Must execute before " + when(desk.ExpiresAt(g)) + ", or the request expires.\n")
	b.WriteString("- Reports, reviewer checklist: [grant #" + itoa(g.ID) + "](" + link("grant/"+itoa(g.ID)) + ").\n\n")
	b.WriteString("| # | Amount | Due | Deliverable, in the grantee's words |\n| --- | --- | --- | --- |\n")
	for i, t := range g.Tranches {
		due := "after a report"
		if i == 0 && g.Advance {
			due = "advance, " + days(r.Window) + " after approval"
		}
		b.WriteString("| " + itoa(i+1) + " | " + gnot(t.Amount) + " | " + due + " | " +
			sanitize.TableCell(t.Deliverable) + " |\n")
	}
	b.WriteString("\nPitch, in the grantee's words (the full text is on the grant page): " + sanitize.InlineText(clamp(g.Pitch, 400)) + "\n")
	b.WriteString("\nSponsor's review, in full:\n\n" + quote(review))
	return b.String()
}

func cancelText(g gp.Grant, reason string) string {
	effect := "the request is refused; nothing was reserved for it. If GovDAO approves the request first, this fails."
	if g.Status == gp.Active {
		effect = "what was paid stays paid (" + gnot(g.Paid()) + " when proposed), and the rest (" + gnot(g.Total()-g.Paid()) +
			" when proposed) returns to the pool's free balance."
	}
	return "If this passes, grant #" + itoa(g.ID) + " to " + sanitize.InlineCode(g.Grantee) + " stops: " + effect +
		"\n\n[Grant #" + itoa(g.ID) + "](" + link("grant/"+itoa(g.ID)) + ")\n\nReason: " + escape(reason) + "\n"
}

// count prints a number ByGrantee counted, which stops at its cap.
func count(n int) string {
	if n >= gp.MaxListed {
		return itoa(n) + " or more"
	}
	return itoa(n)
}

// nextStep is a one-line status for tables, with the call anyone can make
// when one is due.
func nextStep(g gp.Grant) string {
	n := itoa(len(g.PaidAt) + 1)
	switch {
	case g.Status == gp.Pending && expirable(g):
		return "<gno-icon name=\"alarm\" /> expired unapproved · [Expire](" + tx("Expire", g.ID) + ")"
	case g.Status == gp.Pending && desk.Closed():
		return "<gno-icon name=\"lock\" /> desk closed · expires " + date(desk.ExpiresAt(g).Unix())
	case g.Status == gp.Pending && hostFrozen(g.Grantee):
		return "<gno-icon name=\"lock\" /> grantee frozen by GovDAO · expires " + date(desk.ExpiresAt(g).Unix())
	case g.Status == gp.Pending:
		return "<gno-icon name=\"clock\" /> needs a GovDAO vote before " + date(desk.ExpiresAt(g).Unix())
	case g.Status != gp.Active:
		return badge(g)
	case expirable(g):
		return "<gno-icon name=\"alarm\" /> payment " + n + " never reported · [Expire](" + tx("Expire", g.ID) + ")"
	case Frozen(g.Grantee):
		return "<gno-icon name=\"lock\" /> grantee frozen by GovDAO"
	case due(g):
		return "<gno-icon name=\"coins\" /> payment " + n + " ready · [Claim](" + tx("Claim", g.ID) + ")"
	}
	held := ""
	if g.Held() {
		held = " <gno-icon name=\"lock\" label=\"held\" />"
	}
	if at, err := desk.PayableAt(g.ID); err == nil {
		return "<gno-icon name=\"calendar\" /> payment " + n + " due " + date(at.Unix()) + held
	}
	return "<gno-icon name=\"chat\" /> waiting for a report on payment " + n + held
}

// checklist is what a member checks before sponsoring or voting. A wallet
// account runs code shared by every account, so what is left is who
// controls it; a realm of its own needs its code read.
func checklist(g gp.Grant) string {
	people := "- Who controls it, and is that person or team known to GovDAO? Name them in your review.\n" +
		"- Do you, or anyone voting, have an interest in it? Say so in your review, and abstain.\n" +
		"- Do past reports and spends match what was promised? See [its other grants](" + link("ns/"+namespace(g.Grantee)) + ").\n"
	if hosted(g.Grantee) {
		return "\n**This grantee is an account in [the grant wallet](" + granteePage(g.Grantee) + "),** whose [code](" +
			realmLink(walletPath) + "$source) every account shares: only its owner can spend, every spend is journaled there, " +
			"and GovDAO can freeze it.\n\n" + people +
			"- Is the owner a realm's address? Then that realm's code decides who spends: read it, and do not fund a private one.\n"
	}
	src := realmLink(gp.Host(g.Grantee)) + "$source"
	return "\n**This grantee is a realm of its own: read [its code](" + src + ") before you vote.**\n\n" + people +
		"- Is `private = true` in its `gnomod.toml`? A private realm can be redeployed with new code at the same address after the vote: do not fund one.\n" +
		"- Who can make it call `Report`? That report is the gate before each payment.\n" +
		"- Every `SendCoins` and `NewBanker`: to whom, behind which check?\n" +
		"- Can an admin change who that is: an ownership transfer, a stored address, func or interface?\n" +
		"- Does it pass `cur` to code it does not own, or import realms with admins of their own?\n" +
		"- Can the coins get stuck, with no way to spend them?\n\n" +
		"Automated help: [the audit pattern harness](https://github.com/gnolang/gno/tree/master/misc/audit-pattern-harness) " +
		"and [the contract review guide](https://github.com/gnolang/gno/blob/master/docs/resources/gno-ai-contract-review.md). " +
		"If it only holds and spends grant money, suggest the wallet instead.\n"
}

// ----- helpers -----

// due reports whether anyone can release the next payment right now.
func due(g gp.Grant) bool {
	at, err := desk.PayableAt(g.ID)
	return err == nil && !time.Now().Before(at) && !Frozen(g.Grantee)
}

func expirable(g gp.Grant) bool {
	at := desk.ExpiresAt(g)
	return !at.IsZero() && !time.Now().Before(at)
}

func icon(g gp.Grant) string {
	switch g.Status {
	case gp.Pending:
		if expirable(g) {
			return "alarm"
		}
		return "clock"
	case gp.Active:
		if g.Held() {
			return "lock"
		}
		return "play"
	case gp.Completed:
		return "check-circle"
	case gp.Cancelled:
		return "ban"
	case gp.Superseded:
		return "arrow-right-circle"
	}
	return "alarm"
}

func title(s string) string { return strings.ToUpper(s[:1]) + s[1:] }

// Label is a grant's status as the desk's pages name it, for a grantee's
// own pages to use too.
func Label(g gp.Grant) string {
	switch {
	case g.Status == gp.Pending && expirable(g):
		return "expired unapproved"
	case g.Status == gp.Active && expirable(g):
		return "report overdue"
	case g.Status == gp.Completed:
		return "paid in full"
	case g.Status == gp.Cancelled && g.Reason == gp.WithdrawnReason:
		return "withdrawn"
	}
	return g.Status.String()
}

func badge(g gp.Grant) string {
	// As Claim and approval check: a request waits on its host too, an
	// approved grant only on itself.
	frozen := ""
	if g.Status == gp.Active && Frozen(g.Grantee) || g.Status == gp.Pending && hostFrozen(g.Grantee) {
		frozen = " · <gno-icon name=\"lock\" /> frozen"
	}
	return "<gno-icon name=\"" + icon(g) + "\" /> **" + Label(g) + "**" + frozen
}

// progress draws one block per payment, filled once paid.
func progress(g gp.Grant) string {
	n := len(g.PaidAt)
	return strings.Repeat("▰", n) + strings.Repeat("▱", len(g.Tranches)-n)
}

func card(icon, label, value string) string {
	return "<gno-icon name=\"" + icon + "\" /> " + label + "\n\n### " + value
}

// cards lays values out as a row of cards.
func cards(items ...string) string {
	var b strings.Builder
	b.WriteString("<gno-columns>\n")
	for i, it := range items {
		if i > 0 {
			b.WriteString("<gno-columns-sep />\n")
		}
		b.WriteString("<gno-frame>\n" + it + "\n</gno-frame>\n")
	}
	b.WriteString("</gno-columns>\n\n")
	return b.String()
}

// form is a one-click call on this realm for grant id.
func form(fn string, id int) string {
	return "<gno-form exec=\"" + fn + "\">\n<gno-input name=\"id\" type=\"number\" value=\"" + itoa(id) +
		"\" readonly=\"true\" />\n</gno-form>\n\n"
}

const daoImports = "\"gno.land/r/gov/dao\"\n\tgrants \"" + path + "\""

func script(imports, body string) string {
	return "```go\npackage main\n\nimport (\n\t" + imports + "\n)\n\nfunc main(cur realm) {\n\t" + body + "\n}\n```\n"
}

func tx(fn string, id int) string { return txlink.Realm(path).Call(fn, "id", itoa(id)) }

func payments(n int) string {
	if n == 1 {
		return "one payment"
	}
	return itoa(n) + " payments"
}

// realmLinks names a grantee and links its page and code. A sub-realm
// identity ("host#sub") runs its host's code, so that is what the source link
// opens; a wallet account's page is its own.
func realmLinks(grantee string) string {
	return "[" + sanitize.InlineCode(short(grantee)) + "](" + granteePage(grantee) + ") ([source](" + realmLink(gp.Host(grantee)) + "$source))"
}

// hosted reports whether grantee is an account in the grant wallet.
func hosted(grantee string) bool { return strings.HasPrefix(grantee, walletPath+"#") }

// granteePage is where a grantee shows what it does with its grants.
func granteePage(grantee string) string {
	if hosted(grantee) {
		return realmLink(walletPath) + ":" + grantee[len(walletPath)+1:]
	}
	return realmLink(gp.Host(grantee))
}

func realmLink(p string) string { return "/" + short(p) }

func short(p string) string { return strings.TrimPrefix(p, "gno.land/") }

// namespace is the first segment after r/ ("gno.land/r/team/x" -> "team"),
// or a wallet account's first segment.
func namespace(p string) string {
	if hosted(p) {
		ns, _, _ := strings.Cut(p[len(walletPath)+1:], "/")
		return ns
	}
	parts := strings.SplitN(gp.Host(p), "/", 4)
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}

func isName(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

func pageNum(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 1_000_000 {
		return 0
	}
	return n
}

func link(sub string) string {
	l := realmLink(path)
	if sub == "" {
		return l
	}
	return l + ":" + sub
}

func escape(s string) string { return sanitize.InlineText(s) }

// quote renders the grantee's own prose as a blockquote, one escaped line
// per line. Escaping each line inline, rather than as a block, leaves the
// grantee no heading, list, link or code fence: nothing in it can pass for
// the desk's own chrome, such as the MsgRun scripts members copy.
func quote(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		line = sanitize.InlineText(line)
		// "=" is left alone inline, but a line of them under a line of
		// text would turn it into a heading.
		if t := strings.TrimLeft(line, " \t"); strings.HasPrefix(t, "=") {
			line = "\\" + t
		}
		b.WriteString("> " + line + "\n")
	}
	return b.String() + "\n"
}

func when(t time.Time) string { return t.UTC().Format("2006-01-02 15:04 UTC") }

func whenU(u int64) string { return when(time.Unix(u, 0)) }

func date(u int64) string { return time.Unix(u, 0).UTC().Format("2006-01-02") }

func days(d time.Duration) string { return itoa(int(d/(24*time.Hour))) + " days" }

// GNOT formats an ugnot amount as GNOT, with thousands separators, as the
// desk's pages do.
func GNOT(ugnot int64) string { return gnot(ugnot) }

func gnot(n int64) string {
	sign, u := "", uint64(n)
	if n < 0 {
		sign, u = "-", -u
	}
	whole := strconv.FormatUint(u/1_000_000, 10)
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	if frac := u % 1_000_000; frac != 0 {
		whole += "." + strings.TrimRight(strconv.FormatUint(frac+1_000_000, 10)[1:], "0")
	}
	return sign + whole + " GNOT"
}

// clamp cuts s to at most n bytes on a rune boundary.
func clamp(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n] + "…"
}
