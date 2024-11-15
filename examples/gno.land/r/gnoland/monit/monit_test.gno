package monit

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestPackage(t *testing.T) {
	// initial state, watchdog is KO.
	{
		expected := `counter=0
last update=0001-01-01 00:00:00 +0000 UTC
last caller=
status=KO`
		got := Render("")
		uassert.Equal(t, expected, got)
	}

	// call Incr(), watchdog is OK.
	Incr()
	Incr()
	Incr()
	{
		expected := `counter=3
last update=2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
last caller=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
status=OK`
		got := Render("")
		uassert.Equal(t, expected, got)
	}

	/* XXX: improve tests once we've the missing std.TestSkipTime feature
		// wait 1h, watchdog is KO.
		use std.TestSkipTime(time.Hour)
		{
			expected := `counter=3
	last update=2009-02-13 22:31:30 +0000 UTC m=+1234564290.000000001
	last caller=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	status=KO`
			got := Render("")
			uassert.Equal(t, expected, got)
		}

		// call Incr(), watchdog is OK.
		Incr()
		{
			expected := `counter=4
	last update=2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
	last caller=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	status=OK`
			got := Render("")
			uassert.Equal(t, expected, got)
		}
	*/
}
