package mail

import (
	"strings"
	"testing"
	"time"
)

var (
	UTC         = time.UTC
	y1985       = time.Date(1985, 1, 1, 0, 0, 0, 0, UTC)
	y2015       = time.Date(2015, 1, 1, 0, 0, 0, 0, UTC)
	y2023       = time.Date(2023, 1, 1, 0, 0, 0, 0, UTC)
	y2024       = time.Date(2024, 1, 1, 0, 0, 0, 0, UTC)
	y2048       = time.Date(2048, 1, 1, 0, 0, 0, 0, UTC)
	timenarrow1 = time.Date(1914, 6, 28, 14, 11, 0, 0, UTC)
	timenarrow2 = time.Date(1914, 6, 28, 14, 13, 0, 0, UTC)
	timenarrow3 = time.Date(1914, 6, 28, 14, 16, 0, 0, UTC)

	oldmail = &Mail{1, "1985", "old", y1985}
	nowmail = &Mail{2, "2023", "now", y2023}
	farmail = &Mail{3, "2048", "far", y2048}
)

func assert(t *testing.T, cond bool) {
	t.Helper()
	if !cond {
		t.Errorf("condition failed")
	}
}

// this bad name actually means:
// assert array `a` has an `i`-th element that is a mail,
// with mail.body containing `s`
func assertContains(t *testing.T, a []*Mail, i int, s string) {
	t.Helper()
	if i > len(a)-1 {
		t.Errorf("Array does not have requested index\n")
	} else {
		var mail *Mail = a[i]
		if !strings.Contains(mail.body, s) {
			t.Errorf("Mail must contain " + s + ". Actual: " + mail.body)
		}
	}
}

func TestPredicates(t *testing.T) {
	ctr := newContainerWith([]*Mail{oldmail, nowmail, farmail})
	{
		a := ctr.filter(TimeAfter{y2015})
		if len(a) != 2 {
			t.Fail()
		} else {
			assertContains(t, a, 0, "now")
			assertContains(t, a, 1, "far")
		}
	}
	{
		a := ctr.filter(TimeBefore{y2015})
		assertContains(t, a, 0, "old")
		assert(t, len(a) == 1)
	}
	{
		a := ctr.filter(TimeBetween{y2015, y2024})
		assertContains(t, a, 0, "now")
		assert(t, len(a) == 1)
	}
	{
		a := ctr.filter(And(TimeAfter{y1985}, TimeBefore{y2024}, Contains{"now"}, Not{Contains{"far"}}))
		assert(t, len(a) == 1)
		assertContains(t, a, 0, "now")
	}
}
