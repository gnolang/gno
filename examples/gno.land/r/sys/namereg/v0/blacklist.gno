package namereg

// reservedNames lists role/system identifiers that must never be allocated
// as a registered name. The intent is to prevent Open Nym Tier
// auto-registrations like "nym-admin000" from impersonating system roles.
//
// Sources merged here:
//   - Common role names already covered by Handshake's valid.json (the 90k
//     curated trademark/gTLD-application list) — admin, help, support, etc.
//   - Common role names NOT covered by Handshake — administrator, root,
//     sysadmin, owner, staff, api, etc.
//   - RFC 2606 / RFC 6761 reserved labels — example, invalid, localhost,
//     local, test.
//
// Plural rule: every entry below is ALSO reserved with the literal "s"
// suffix appended. So "doc" reserves both "doc" and "docs"; "setting"
// reserves both "setting" and "settings"; "new" covers "news", and so on.
// Entries are stored in the singular here and the validator appends "s"
// at check time. This halves list maintenance and avoids the temptation
// to add `name+"s"` after every singular entry.
//
// Note on length: the Open Nym Tier regex restricts the [a-z]{5,13}
// middle to 5–13 chars, so entries shorter than 5 (mod, api, bot, god,
// gno, ...) and longer than 13 (administrator, jesuschrist,
// newtendermint, ...) cannot appear in Register() even without this list.
// They are kept anyway because:
//
//	(a) cross-reference value — a single canonical list is easier to audit
//	    than two scope-specific lists with overlapping intent; and
//	(b) future controllers (e.g. a hypothetical DAO-allocated path) can
//	    opt in to the same blacklist by querying IsReserved, providing
//	    defense-in-depth across the registration ecosystem.
//
// IMPORTANT: this list is NOT consulted by `ProposeNewName` — governance
// renames are gated by GovDAO vote alone, not by this blacklist. Voters
// reviewing a rename proposal are responsible for catching collisions
// with reserved names.
//
// Sortedness: entries must be sorted lexicographically (Go's < on strings,
// which is byte-wise ASCII). TestReservedNamesSorted enforces this.
var reservedNames = []string{
	"about",
	"abuse",
	"account",
	"admin",
	"administrator",
	"aib",
	"aibinc",
	"allinbits",
	"allinbitsinc",
	"anonymous",
	"api",
	"atom",
	"atomone",
	"atomonehub",
	"atone",
	"atonehub",
	"bitcoin",
	"blockchain",
	"blog",
	"bot",
	"chain",
	"coin",
	"community",
	"contact",
	"cosmos",
	"cosmoshub",
	"crypto",
	"daemon",
	"dashboard",
	"default",
	"demo",
	"doc",
	"domain",
	"email",
	"ether",
	"ethereum",
	"everyone",
	"example",
	"gno",
	"gnoland",
	"gnolang",
	"gnome",
	"gnot",
	"gnoworld",
	"god",
	"gov",
	"govdao",
	"governance",
	"guest",
	"help",
	"home",
	"host",
	"info",
	"invalid",
	"jaekwon",
	"jesus",
	"jesuschrist",
	"local",
	"localhost",
	"login",
	"mail",
	"mod",
	"moderator",
	"new",
	"newtendermint",
	"newtendermintllc",
	"nt",
	"ntllc",
	"null",
	"owner",
	"photon",
	"profile",
	"register",
	"registry",
	"resolution",
	"resolve",
	"resolver",
	"root",
	"security",
	"service",
	"setting",
	"shop",
	"signup",
	"staff",
	"store",
	"support",
	"sys",
	"sysadmin",
	"system",
	"team",
	"tendermint",
	"test",
	"user",
}
