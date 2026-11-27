// Package wallet holds grant accounts for teams and projects: the way to take
// GovDAO grants from gno.land/r/gov/grants/v0 without deploying code.
//
// An account is named after a namespace: "gnogolf", or "gnogolf/cup" for a
// second project. Whoever may deploy under that namespace (r/sys/names) opens
// it and owns it, and can hand it over, for instance to a multisig or a DAO
// realm, which takes it over by accepting it. Each account is a sub-realm identity of this realm,
// gno.land/r/gov/grants/wallet/v0#<name>, with an address of its own: the
// desk pays its grants there, and only its owner moves coins out, each spend
// with a memo in the account's public journal. There is no other way out: no
// admin, no stored callback, no upgrade. GovDAO can freeze an account through
// the desk: a frozen account cannot spend or change hands, and is paid
// nothing.
//
// Every account runs this same code, to review once; a grant to an account
// leaves voters one thing to check, who owns it.
package wallet

import (
	"chain"
	"chain/banker"
	"strconv"
	"strings"
	"time"

	"gno.land/p/nt/avl/v0"
	gp "gno.land/p/nt/grants/v0"
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/seqid/v0"

	grants "gno.land/r/gov/grants/v0"
	"gno.land/r/sys/names"
)

const (
	// MaxMemo bounds the note every spend carries.
	MaxMemo = 280

	path     = "gno.land/r/gov/grants/wallet/v0"
	pageSize = 50
)

type account struct {
	owner   address
	offered address   // who may accept the account, "" for nobody
	journal *avl.Tree // seqid key -> entry
	next    seqid.ID
	spent   int64
}

// entry is one line of an account's journal: a spend, the opening, or a
// change of owner.
type entry struct {
	At     int64
	To     address
	Amount int64 // ugnot; 0 for the opening or a change of owner
	Memo   string
}

var accounts = avl.NewTree() // name -> *account

// Open opens the account name for the caller, who must be allowed to deploy
// under its namespace, the part before any "/".
func Open(cur realm, name string) {
	caller := cur.Previous().Address()
	if !validName(name) {
		panic("an account name is a namespace, or namespace/project: lowercase letters and digits, with _ . - inside")
	}
	ns, _, _ := strings.Cut(name, "/")
	if !names.IsAuthorizedAddressForNamespace(caller, ns) {
		panic("only who may deploy under r/" + ns + " can open " + name)
	}
	if accounts.Has(name) {
		panic("account " + name + " is already open")
	}
	a := &account{owner: caller, journal: avl.NewTree()}
	accounts.Set(name, a)
	a.record(entry{At: time.Now().Unix(), To: caller, Memo: "opened"})
	chain.Emit("AccountOpened", "name", name, "owner", caller.String())
}

// Apply asks GovDAO for a grant to account name and returns its id. tranches
// reads "deliverable: GNOT; deliverable: GNOT" in whole GNOT, each paid once
// you report its deliverable done, at least periodDays apart. A first tranche
// of at most the desk's advance share is paid without a report. Owner only.
func Apply(cur realm, name, title, pitch, tranches string, periodDays int) int {
	mustOwn(cur, name)
	if periodDays < 0 || periodDays > 366 {
		panic("periodDays must be between 0 and 366")
	}
	ts, err := gp.ParseTranches(tranches)
	if err != nil {
		panic(err.Error())
	}
	for i := range ts {
		if ts[i].Amount <= 0 || ts[i].Amount > 1<<42 {
			panic("tranche amounts are whole GNOT, at least 1")
		}
		ts[i].Amount *= 1_000_000
	}
	sub := cur.Sub(name)
	return grants.Apply(cross(sub), title, pitch, ts, time.Duration(periodDays)*24*time.Hour)
}

// Report tells GovDAO what grant id delivered for its next payment, with
// link to what shows it (a gno.land path or an https URL), which may be
// empty. Owner only.
func Report(cur realm, name string, id int, body, link string) {
	mustOwn(cur, name)
	sub := cur.Sub(name)
	grants.Report(cross(sub), id, body, link)
}

// Withdraw gives grant id up; what it was not paid goes back to GovDAO's
// pool. Owner only.
func Withdraw(cur realm, name string, id int) {
	mustOwn(cur, name)
	sub := cur.Sub(name)
	grants.Withdraw(cross(sub), id)
}

// Spend pays amount GNOT ("12" or "0.5") out of account name. memo says what
// it pays for; it goes into the account's journal, which reviewers read next
// to its reports at the desk. Owner only, and never while GovDAO has the
// account frozen.
func Spend(cur realm, name string, to address, amount, memo string) {
	a := mustOwn(cur, name)
	mustNotFrozen(name)
	// Nothing can spend from either.
	if !to.IsValid() || to == Address(name) || to == chain.PackageAddress(path) {
		panic("invalid recipient: an address other than the account's or the wallet's own")
	}
	ugnot, ok := parseGNOT(amount)
	if !ok {
		panic("amount is in GNOT, more than 0, with at most 6 decimals: 12 or 0.5")
	}
	// Zero-width and bidi characters alone would leave an empty journal line.
	if strings.TrimSpace(sanitize.StripBidiAndZeroWidth(memo)) == "" || len(memo) > MaxMemo {
		panic("a memo of 1 to " + strconv.Itoa(MaxMemo) + " bytes is required")
	}
	a.record(entry{At: time.Now().Unix(), To: to, Amount: ugnot, Memo: memo})
	a.spent += ugnot
	sub := cur.Sub(name)
	banker.NewBanker(banker.BankerTypeRealmSend, sub).
		SendCoins(sub.Address(), to, chain.NewCoins(chain.NewCoin("ugnot", ugnot)))
	chain.Emit("Spend", "account", name, "to", to.String(), "amount", strconv.FormatInt(ugnot, 10))
}

// TransferOwnership offers account name to newOwner: a user's, a multisig's
// or a realm's address. It changes hands once newOwner calls AcceptOwnership,
// which proves it can act: an address that never can, such as another
// account's, cannot lock the account. Until then the owner keeps it and can
// offer it again. Owner only, and never while GovDAO has the account frozen;
// both steps are journaled.
func TransferOwnership(cur realm, name string, newOwner address) {
	a := mustOwn(cur, name)
	mustNotFrozen(name)
	if !newOwner.IsValid() {
		panic("invalid new owner: a user's, a multisig's or a realm's address")
	}
	a.offered = newOwner
	a.record(entry{At: time.Now().Unix(), To: newOwner, Memo: "handover offered"})
	chain.Emit("OwnershipOffered", "account", name, "to", newOwner.String())
}

// AcceptOwnership makes the caller the owner of account name, once its owner
// offered it to the caller with TransferOwnership. Never while GovDAO has
// the account frozen.
func AcceptOwnership(cur realm, name string) {
	a := get(name)
	caller := cur.Previous().Address()
	if a == nil || a.offered == "" || a.offered != caller {
		panic("account " + name + " was not offered to you")
	}
	mustNotFrozen(name)
	a.owner, a.offered = caller, ""
	a.record(entry{At: time.Now().Unix(), To: caller, Memo: "new owner"})
	chain.Emit("OwnerChanged", "account", name, "owner", caller.String())
}

// Owner is who acts for account name, and the zero address if it is not open.
func Owner(name string) address {
	if a := get(name); a != nil {
		return a.owner
	}
	return ""
}

// Address is where account name receives its grants. It panics on a name no
// account can have, so no one is handed an address nothing can spend from.
func Address(name string) address {
	if !validName(name) {
		panic("no account can be named " + strconv.Quote(name))
	}
	return chain.DerivePkgSubAddr(path, name)
}

func mustOwn(cur realm, name string) *account {
	a := get(name)
	if a == nil {
		panic("no account " + name + ": open it first")
	}
	if cur.Previous().Address() != a.owner {
		panic("only the owner of " + name + " can do this")
	}
	return a
}

func mustNotFrozen(name string) {
	if grants.Frozen(path + "#" + name) {
		panic("GovDAO froze account " + name)
	}
}

func get(name string) *account {
	v := accounts.Get(name)
	if v == nil {
		return nil
	}
	return v.(*account)
}

func (a *account) record(e entry) { a.journal.Set(a.next.Next().String(), e) }

// validName accepts one or two segments of lowercase letters and digits,
// with "_", "." or "-" inside a segment: what cur.Sub accepts, kept short.
func validName(s string) bool {
	if len(s) == 0 || len(s) > 64 || strings.Count(s, "/") > 1 {
		return false
	}
	for _, seg := range strings.Split(s, "/") {
		if seg == "" {
			return false
		}
		for i := 0; i < len(seg); i++ {
			c := seg[i]
			alnum := c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
			if !alnum && (i == 0 || i == len(seg)-1 || strings.IndexByte("_.-", c) < 0) {
				return false
			}
		}
	}
	return true
}

// parseGNOT reads a positive GNOT amount with at most 6 decimals as ugnot.
func parseGNOT(s string) (int64, bool) {
	whole, frac, _ := strings.Cut(strings.TrimSpace(s), ".")
	if whole == "" || len(whole) > 12 || len(frac) > 6 {
		return 0, false
	}
	frac += strings.Repeat("0", 6-len(frac))
	for _, c := range whole + frac {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(whole+frac, 10, 64)
	return n, err == nil && n > 0
}

// ----- pages -----

// Render serves the list of accounts at "" (more at "_<n>"), an account at
// "<name>", and older lines of its journal at "<name>/_<n>". Names never
// start a segment with "_".
func Render(p string) string {
	page := 1
	if i := strings.LastIndex(p, "_"); i >= 0 && (i == 0 || p[i-1] == '/') {
		n, err := strconv.Atoi(p[i+1:])
		if err != nil || n < 1 || n > 1_000_000 {
			return notFound
		}
		page, p = n, strings.TrimSuffix(p[:i], "/")
	}
	if p == "" {
		return home(page)
	}
	if a := get(p); a != nil && validName(p) {
		return accountPage(p, a, page)
	}
	return notFound
}

const notFound = "# Not found\n\nNo such account.\n\n<gno-button href=\"/r/gov/grants/wallet/v0\" label=\"All accounts\" />\n"

func home(page int) string {
	var b strings.Builder
	b.WriteString("# <gno-icon name=\"wallet\" /> Grant wallet\n\n")
	b.WriteString("Accounts that take GovDAO grants from [the grants desk](/r/gov/grants/v0) for a team or a project, " +
		"without deploying code. Every account runs [the same code](/r/gov/grants/wallet/v0$source), so voters only need to know " +
		"who owns one: only an account's owner can spend, every spend is in its public journal, and GovDAO can freeze it.\n\n")
	b.WriteString("| Account | Owner | Balance | Spent |\n| --- | --- | --- | --- |\n")
	accounts.IterateByOffset((page-1)*pageSize, pageSize, func(name string, v any) bool {
		a := v.(*account)
		b.WriteString("| [" + sanitize.InlineCode(name) + "](" + sanitize.URL("/r/gov/grants/wallet/v0:"+name) + ") | " +
			sanitize.InlineCode(a.owner.String()) + " | " + gnot(balance(name)) + " | " + gnot(a.spent) + " |\n")
		return false
	})
	b.WriteString("\n")
	if page*pageSize < accounts.Size() {
		b.WriteString("<gno-button href=\"/r/gov/grants/wallet/v0:_" + strconv.Itoa(page+1) + "\" label=\"More accounts\" variant=\"outline\" />\n\n")
	}
	b.WriteString("## <gno-icon name=\"plus\" /> Open an account\n\n" +
		"Name it after a namespace you may deploy under: your registered name, or your g1 address (`gnogolf`, or `gnogolf/cup` " +
		"for a second project, once that name is yours). You become its owner, and can hand the account to a multisig or a DAO " +
		"realm later; voters will then read that realm's code.\n\n" +
		"<gno-form exec=\"Open\">\n<gno-input name=\"name\" type=\"text\" placeholder=\"gnogolf\" />\n</gno-form>\n")
	return b.String()
}

func accountPage(name string, a *account, page int) string {
	self := path + "#" + name
	link := sanitize.URL("/r/gov/grants/wallet/v0:" + name)
	list := grants.GrantsOf(self)
	var received int64
	for _, g := range list {
		received += g.Paid()
	}
	from := "Received from grants"
	if len(list) == 20 { // GrantsOf lists the latest 20
		from = "Received, latest 20 grants"
	}

	var b strings.Builder
	b.WriteString("# <gno-icon name=\"briefcase\" /> " + sanitize.InlineText(name) + "\n\n")
	b.WriteString("A grant account in [the grant wallet](/r/gov/grants/wallet/v0). Its owner asks GovDAO for grants, reports " +
		"and spends; every spend and every change of owner is in the journal below. Grants arrive at " +
		sanitize.InlineCode(Address(name).String()) + ".\n\n")
	frozen := grants.Frozen(self)
	if frozen {
		b.WriteString("> [!CAUTION] Frozen by GovDAO\n> This account cannot spend or change hands, and the desk pays it nothing, " +
			"until GovDAO unfreezes it. Reason: " + sanitize.InlineText(grants.FrozenReason(self)) + "\n\n")
	}
	b.WriteString("<gno-columns>\n<gno-frame>\n<gno-icon name=\"user\" /> Owner\n\n" + sanitize.InlineCode(a.owner.String()) +
		"\n</gno-frame>\n<gno-columns-sep />\n<gno-frame>\n<gno-icon name=\"wallet\" /> Balance\n\n### " + gnot(balance(name)) +
		"\n</gno-frame>\n<gno-columns-sep />\n<gno-frame>\n<gno-icon name=\"coins\" /> " + from + "\n\n### " + gnot(received) +
		"\n</gno-frame>\n<gno-columns-sep />\n<gno-frame>\n<gno-icon name=\"send\" /> Spent\n\n### " + gnot(a.spent) +
		"\n</gno-frame>\n</gno-columns>\n\n")
	if a.offered != "" {
		b.WriteString("Handover offered to " + sanitize.InlineCode(a.offered.String()) +
			": it becomes the owner when it accepts with AcceptOwnership. Until then the owner keeps the account.\n\n")
	}

	b.WriteString("## <gno-icon name=\"coins\" /> Grants\n\n")
	if len(list) == 0 {
		b.WriteString("None yet. The owner asks for one with the form below.\n\n")
	} else {
		b.WriteString("| # | Grant | Paid | Status |\n| --- | --- | --- | --- |\n")
		for _, g := range list {
			b.WriteString("| " + strconv.Itoa(g.ID) + " | [" + sanitize.TableCell(g.Title) + "](/r/gov/grants/v0:grant/" +
				strconv.Itoa(g.ID) + ") | " + gnot(g.Paid()) + " of " + gnot(g.Total()) + " | " + grants.Label(g) + " |\n")
		}
		b.WriteString("\nWhat each grant waits for, and when it pays, is on its page at the desk.\n\n")
	}

	b.WriteString("## <gno-icon name=\"receipt\" /> Journal\n\n| When | To | Amount | For |\n| --- | --- | --- | --- |\n")
	a.journal.ReverseIterateByOffset((page-1)*pageSize, pageSize, func(_ string, v any) bool {
		e := v.(entry)
		b.WriteString("| " + time.Unix(e.At, 0).UTC().Format("2006-01-02") + " | " + sanitize.InlineCode(e.To.String()) +
			" | " + gnot(e.Amount) + " | " + sanitize.TableCell(e.Memo) + " |\n")
		return false
	})
	b.WriteString("\n")
	if page*pageSize < a.journal.Size() {
		b.WriteString("<gno-button href=\"" + link + "/_" + strconv.Itoa(page+1) + "\" label=\"Older\" variant=\"outline\" />\n\n")
	}

	field := "<gno-input name=\"name\" type=\"text\" value=\"" + sanitize.HTMLEscape(name) + "\" readonly=\"true\" />\n"
	b.WriteString("## <gno-icon name=\"edit\" /> Act (owner)\n\n")
	closed := !frozen && grants.Closed()
	if frozen {
		b.WriteString("GovDAO froze this account: it cannot apply, spend or change hands until GovDAO unfreezes it.\n\n")
	} else if closed {
		b.WriteString("GovDAO closed [the grants desk](/r/gov/grants/v0): it takes no new request.\n\n")
	}
	if !frozen && !closed {
		b.WriteString("<gno-details summary=\"Ask GovDAO for a grant\">\n\n" +
			"Amounts are whole GNOT. A payment falls due once you report its deliverable done and GovDAO's review window passes; " +
			"anyone then releases it with Claim on the grant's page at the desk. A first payment of at most " +
			strconv.Itoa(grants.Rules().Advance) + "% of the grant is an advance, due without a report: claim it, then report on payment 2. " +
			"Payments are at least the period apart: 30 days or more, or 0 for a single payment.\n\n<gno-form exec=\"Apply\">\n" + field +
			"<gno-input name=\"title\" type=\"text\" placeholder=\"Gnogolf v3\" />\n" +
			"<gno-textarea name=\"pitch\" rows=\"6\" placeholder=\"What it is, why it grows gno.land, who does it, how you will measure it\" />\n" +
			"<gno-input name=\"tranches\" type=\"text\" placeholder=\"v3 design published: 1000; v3 live with 100 players: 4000\" />\n" +
			"<gno-input name=\"periodDays\" type=\"number\" placeholder=\"Days between payments: 30 or more, 0 for one payment\" />\n" +
			"</gno-form>\n\n</gno-details>\n\n")
	}
	b.WriteString("<gno-details summary=\"Report on a grant\">\n\nThe report opens " +
		strconv.FormatInt(int64(grants.Rules().Window/(24*time.Hour)), 10) + " days in which GovDAO can object before the payment; " +
		"reporting again restarts them.\n\n<gno-form exec=\"Report\">\n" + field +
		"<gno-input name=\"id\" type=\"number\" placeholder=\"Grant id\" />\n" +
		"<gno-textarea name=\"body\" rows=\"8\" placeholder=\"Done: what the deliverable promised. Numbers: users, transactions, reach. Spent: see the journal. Next: what the next payment pays for.\" />\n" +
		"<gno-input name=\"link\" type=\"text\" placeholder=\"What shows it: /r/gnogolf/game:v3 or https://…\" />\n" +
		"</gno-form>\n\n</gno-details>\n\n")
	if !frozen {
		b.WriteString("<gno-details summary=\"Spend\">\n\n<gno-form exec=\"Spend\">\n" + field +
			"<gno-input name=\"to\" type=\"text\" placeholder=\"g1… recipient\" />\n" +
			"<gno-input name=\"amount\" type=\"text\" placeholder=\"GNOT, 12 or 0.5\" />\n" +
			"<gno-input name=\"memo\" type=\"text\" placeholder=\"What it pays for\" />\n" +
			"</gno-form>\n\n</gno-details>\n\n")
	}
	b.WriteString("<gno-details summary=\"Withdraw a grant\">\n\nGives the grant up; what it was not paid goes back to GovDAO's pool.\n\n" +
		"<gno-form exec=\"Withdraw\">\n" + field + "<gno-input name=\"id\" type=\"number\" placeholder=\"Grant id\" />\n</gno-form>\n\n</gno-details>\n\n")
	if !frozen {
		b.WriteString("<gno-details summary=\"Hand the account over\">\n\n<gno-form exec=\"TransferOwnership\">\n" + field +
			"<gno-input name=\"newOwner\" type=\"text\" placeholder=\"g1… new owner, a multisig or a DAO realm's address\" />\n" +
			"</gno-form>\n\n</gno-details>\n\n")
	}
	if a.offered != "" && !frozen {
		b.WriteString("<gno-details summary=\"Accept the handover (offered owner)\">\n\n<gno-form exec=\"AcceptOwnership\">\n" + field +
			"</gno-form>\n\n</gno-details>\n")
	}
	return b.String()
}

func balance(name string) int64 {
	return banker.NewReadonlyBanker().GetCoin(Address(name), "ugnot")
}

func gnot(ugnot int64) string { return grants.GNOT(ugnot) }
