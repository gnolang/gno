package commondao

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = (*textPropDefinition)(nil)
	_ commondao.DefaultTallier     = (*textPropDefinition)(nil)
)

func TesttextPropDefinitionNew(cur realm, t *testing.T) {
	urequire.PanicsWithMessage(t, cur, "proposal title is empty", func() {
		newTextPropDefinition("  ", "body", time.Hour*24)
	})

	urequire.PanicsWithMessage(t, cur, "proposal title is too long, max length is 255 chars", func() {
		newTextPropDefinition(strings.Repeat("a", maxTextTitle+1), "body", time.Hour*24)
	})

	urequire.PanicsWithMessage(t, cur, "proposal body is empty", func() {
		newTextPropDefinition("title", "  ", time.Hour*24)
	})

	urequire.PanicsWithMessage(t, cur, "proposal body is too long, max length is 15000 chars", func() {
		newTextPropDefinition("title", strings.Repeat("a", maxTextBody+1), time.Hour*24)
	})

	urequire.PanicsWithMessage(t, cur, "minimum proposal voting period is one day", func() {
		newTextPropDefinition("title", "body", time.Hour)
	})

	def := newTextPropDefinition(" title ", " body ", time.Hour*48)
	uassert.Equal(t, "title", def.Title())
	uassert.Equal(t, "body", def.Body())
	uassert.Equal(t, int64(time.Hour*48), int64(def.VotingPeriod()))
	uassert.Equal(t, int(commondao.ThresholdSupermajority), int(def.Threshold()))
}
