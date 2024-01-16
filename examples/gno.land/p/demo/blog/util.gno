package blog

import (
	"strconv"
	"strings"
	"time"
)

func breadcrumb(parts []string) string {
	return "# " + strings.Join(parts, " / ") + "\n\n"
}

func makePubDate(t time.Time, id int64) string {
	// We use the timestamp to sort by publish order.
	// In case of multiple msgs to publish in the same transaction
	// there will be a timestamp conflict (time.Now() is per transaction, not per msg).
	// This functions assures that timestamp IDs are unique in this case.
	return strconv.Itoa(int(t.Unix())) + "-" + strconv.Itoa(int(id))
}
