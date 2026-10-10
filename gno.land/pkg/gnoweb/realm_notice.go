package gnoweb

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// RealmNoticeHeader is set to "community" on markdown responses of pages that
// show the realm notice, whose body is served verbatim.
const RealmNoticeHeader = "X-Gnoweb-Realm-Notice"

// showRealmNotice reports whether u is a community page: a package or user
// page outside the trusted paths, as pagePolicy classifies it. A user page
// renders that user's home realm. The bare "/r/", "/p/" and "/u/" listings
// belong to no package.
func (h *HTTPHandler) showRealmNotice(u *weburl.GnoURL) bool {
	return h.Static.RealmNotice.Enabled() && h.policy.kind(u) == pageCommunity
}
