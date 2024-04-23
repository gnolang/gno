package pagination

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestRenderHome(t *testing.T) {
	// Sign up 50 users
	for i := 0; i < 50; i++ {
		addrName := ufmt.Sprintf("test%d", i)
		caller := testutils.TestAddress(addrName)
		std.TestSetOrigCaller(caller)
		SignUp()
	}

	render := Render("")

	expected := `# Welcome to Gno.land!

## Signed up users (51):

#### User #0 - g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm - signed up at Block #123
#### User #1 - g1w3jhxapsta047h6lta047h6lta047h6l2jx7pq - signed up at Block #123
#### User #2 - g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 - signed up at Block #123
#### User #3 - g1w3jhxapjta047h6lta047h6lta047h6laqcyu4 - signed up at Block #123
#### User #4 - g1w3jhxapnta047h6lta047h6lta047h6lzfhfxt - signed up at Block #123
#### User #5 - g1w3jhxap5ta047h6lta047h6lta047h6ldlnrjr - signed up at Block #123
#### User #6 - g1w3jhxap4ta047h6lta047h6lta047h6ljkuwga - signed up at Block #123
#### User #7 - g1w3jhxapkta047h6lta047h6lta047h6l6dde0k - signed up at Block #123
#### User #8 - g1w3jhxaphta047h6lta047h6lta047h6l9yz54g - signed up at Block #123
#### User #9 - g1w3jhxapcta047h6lta047h6lta047h6lyg9dwx - signed up at Block #123
#### User #10 - g1w3jhxapeta047h6lta047h6lta047h6lmp2q5c - signed up at Block #123
#### User #11 - g1w3jhxap3xp047h6lta047h6lta047h6l0pgx4l - signed up at Block #123
#### User #12 - g1w3jhxap3x9047h6lta047h6lta047h6ls3mqfl - signed up at Block #123
#### User #13 - g1w3jhxap3xf047h6lta047h6lta047h6lcg82yl - signed up at Block #123
#### User #14 - g1w3jhxap3xd047h6lta047h6lta047h6l8c5vcl - signed up at Block #123
#### User #15 - g1w3jhxap3x3047h6lta047h6lta047h6lgnk77l - signed up at Block #123
#### User #16 - g1w3jhxap3x4047h6lta047h6lta047h6lhr9czl - signed up at Block #123
#### User #17 - g1w3jhxap3xe047h6lta047h6lta047h6ll6ej0l - signed up at Block #123
#### User #18 - g1w3jhxap3xa047h6lta047h6lta047h6lq225nl - signed up at Block #123
#### User #19 - g1w3jhxap38p047h6lta047h6lta047h6lac9qrr - signed up at Block #123
---

Latest signup: g1w3jhxap589047h6lta047h6lta047h6l6vvkk7 at Block #123

---

You're viewing page #1

[Next page](/r/demo/pagination:page/2)

`

	if render != expected {
		t.Fatalf("got %s, expected %s", render, expected)
	}
}
