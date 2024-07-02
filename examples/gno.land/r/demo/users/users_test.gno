package users

import (
	"testing"

	"gno.land/p/demo/entropy"
	"gno.land/p/demo/uassert"
)

func TestClaimDefaultName(t *testing.T) {
	ent = entropy.FromSeed(42) // fixed entropy
	profile := "hello world"
	name := ClaimDefaultName(profile)
	expected := "nifty_teal_engineer_51" // this can easily break if p/demo/wordlist is modified
	uassert.Equal(t, expected, name)
}
