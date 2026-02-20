package antispam

import (
	"testing"
)

// hitsTotal sums a []RuleHit slice and returns the total score and highest-scoring rule.
func hitsTotal(hits []RuleHit) (total int, topRule string) {
	topScore := 0
	for _, h := range hits {
		total += h.Score
		if h.Score > topScore {
			topScore = h.Score
			topRule = h.Rule
		}
	}
	return
}

func TestScoreContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantMin  int
		wantMax  int
		wantRule string
		wantZero bool
	}{
		// ---- Success: legitimate content scores zero ----

		{
			name:     "success: normal english text",
			content:  "I'm new here and would like to understand the proposal process. Can someone explain?",
			wantZero: true,
		},
		{
			name:     "success: normal french text",
			content:  "Voici une proposition en français avec des accents: éàü. C'est très important pour la communauté.",
			wantZero: true,
		},
		{
			name:     "success: normal chinese text",
			content:  "你好世界，这是一个关于治理的讨论。我们需要考虑很多因素。",
			wantZero: true,
		},
		{
			name:     "success: empty content",
			content:  "",
			wantZero: true,
		},
		{
			name:     "success: short content no link",
			content:  "Thanks!",
			wantZero: true,
		},
		{
			name:     "success: single URL is fine",
			content:  "Here is a useful link: https://gno.land/r/demo/boards for reference.",
			wantZero: true,
		},
		{
			name:     "success: two URLs is fine",
			content:  "See https://gno.land and also https://docs.gno.land for more information about the project.",
			wantZero: true,
		},
		{
			name:     "success: three URLs is fine",
			content:  "See https://gno.land and https://docs.gno.land and https://github.com/gnolang/gno for sources.",
			wantZero: true,
		},
		{
			name:     "success: some caps emphasis in normal text",
			content:  "This is a REALLY important proposal that we should ALL consider carefully before voting on it.",
			wantZero: true,
		},
		{
			name:     "success: acronyms in normal text",
			content:  "The DAO voted on the GNO proposal. NASA and UNESCO sent their regards to the UN committee.",
			wantZero: true,
		},
		{
			name:     "success: code snippet with punctuation",
			content:  "The fix is: if err != nil { return fmt.Errorf(\"failed: %w\", err) } and that should work.",
			wantZero: true,
		},
		{
			name:     "success: normal text with some exclamation",
			content:  "The proposal passed! Thanks everyone for voting. This is a great step forward!",
			wantZero: true,
		},
		{
			name:     "success: markdown formatted text",
			content:  "## Step 1\nRun `gnokey query` to check.\n## Step 2\nCall the function with **bold** parameters.",
			wantZero: true,
		},
		{
			name:     "success: long legitimate post",
			content:  "This proposal aims to adjust the token distribution model. Currently, validators receive 80% of block rewards while delegators receive 20%. I propose changing this to 70/30 to incentivize broader participation. The rationale is based on analysis of other chains where higher delegator rewards led to better decentralization. We should discuss this at the next governance call and put it to a formal vote. Please share your thoughts below.",
			wantZero: true,
		},
		{
			name:     "success: numbered list",
			content:  "1. Review the PR\n2. Run the tests\n3. Deploy to staging\n4. Verify the fix\n5. Merge to main",
			wantZero: true,
		},

		// ---- ALL_CAPS detection ----

		{
			name:     "caps: 100% uppercase content",
			content:  "THIS IS ALL CAPS AND IT IS VERY LOUD AND ANNOYING TO READ RIGHT NOW",
			wantMin:  3,
			wantRule: "ALL_CAPS",
		},
		{
			name:     "caps: mostly caps (>50%)",
			content:  "THIS IS MOSTLY CAPS AND VERY LOUD WITH JUST A FEW lower words",
			wantMin:  3,
			wantRule: "ALL_CAPS",
		},
		{
			name:     "caps: all caps short message",
			content:  "BUY NOW FREE TOKENS",
			wantMin:  3,
			wantRule: "ALL_CAPS",
		},
		{
			name:     "caps: caps with numbers and symbols",
			content:  "SEND 1000 GNOT TO THIS ADDRESS NOW!!! GET 5000 BACK GUARANTEED!!!",
			wantMin:  3,
			wantRule: "ALL_CAPS",
		},
		{
			name:     "caps: boundary just under 50% should not trigger",
			content:  "SOME CAPS here but actually most of this text is in lowercase so it should be fine overall",
			wantZero: true,
		},

		// ---- Excessive punctuation detection ----

		{
			name:     "punct: exclamation spam",
			content:  "This is amazing!!! Check it out!!! You won't believe it!!! Click now!!! Hurry!!!",
			wantMin:  1,
			wantRule: "EXCESSIVE_PUNCT",
		},
		{
			name:     "punct: question mark spam",
			content:  "What??? Are you serious??? How is this possible??? Really??? Why???",
			wantMin:  1,
			wantRule: "EXCESSIVE_PUNCT",
		},
		{
			name:     "punct: ellipsis heavy",
			content:  "Well... I think... maybe... we should... probably... consider... this... option...",
			wantMin:  1,
			wantRule: "EXCESSIVE_PUNCT",
		},
		{
			name:     "punct: mixed excessive punctuation",
			content:  "Oh my god!!! Wow, really!?! How is this even possible?!?! No way!!! This is crazy!?!",
			wantMin:  1,
			wantRule: "EXCESSIVE_PUNCT",
		},
		{
			name:     "punct: all punctuation",
			content:  "!?!?!?...!!!???...!!!",
			wantMin:  1,
			wantRule: "EXCESSIVE_PUNCT",
		},
		{
			name:     "punct: normal text with reasonable punctuation",
			content:  "Is this the right approach? I think so. Let me know what you think!",
			wantZero: true,
		},

		// ---- Link-heavy detection ----

		{
			name:     "links: four https links",
			content:  "Visit https://spam1.com and https://spam2.com and https://spam3.com and https://spam4.com please",
			wantMin:  2,
			wantRule: "LINK_HEAVY",
		},
		{
			name:     "links: five http links",
			content:  "http://a.com http://b.com http://c.com http://d.com http://e.com",
			wantMin:  2,
			wantRule: "LINK_HEAVY",
		},
		{
			name:     "links: mixed http and https",
			content:  "Check http://one.com and https://two.com and http://three.com and https://four.com now",
			wantMin:  2,
			wantRule: "LINK_HEAVY",
		},
		{
			name:     "links: many links in spam format",
			content:  "https://buy.com\nhttps://cheap.com\nhttps://free.com\nhttps://win.com\nhttps://prize.com\nhttps://bonus.com",
			wantMin:  2,
			wantRule: "LINK_HEAVY",
		},
		{
			name:     "links: exactly three is ok",
			content:  "Main site https://gno.land, docs at https://docs.gno.land, and code at https://github.com/gnolang/gno for reference.",
			wantZero: true,
		},

		// ---- Short content with link ----

		{
			name:     "short+link: very short with link",
			content:  "Click https://scam.com",
			wantMin:  3,
			wantRule: "SHORT_WITH_LINK",
		},
		{
			name:     "short+link: just a URL",
			content:  "https://spam.com",
			wantMin:  3,
			wantRule: "SHORT_WITH_LINK",
		},
		{
			name:     "short+link: short text with http",
			content:  "see http://x.com",
			wantMin:  3,
			wantRule: "SHORT_WITH_LINK",
		},
		{
			name:     "short+link: emoji and link",
			content:  "🔥 https://scam.io",
			wantMin:  3,
			wantRule: "SHORT_WITH_LINK",
		},
		{
			name:     "short+link: longer content with link is fine",
			content:  "I found this very helpful article about Gno development at https://example.com which covers the basics.",
			wantZero: true,
		},

		// ---- Combined rules (multiple triggers) ----

		{
			name:    "combined: caps + many links",
			content: "CLICK HERE NOW https://scam1.com https://scam2.com https://scam3.com https://scam4.com FREE",
			wantMin: 5,
		},
		{
			name:    "combined: caps + punctuation",
			content: "BUY NOW!!! LIMITED TIME!!! DON'T MISS OUT!!! ACT FAST!!! AMAZING DEAL!!!",
			wantMin: 4,
		},
		{
			name:     "combined: short + link",
			content:  "GO https://scam.com",
			wantMin:  3,
			wantRule: "SHORT_WITH_LINK",
		},

		// ---- Edge cases ----

		{
			name:     "edge: only whitespace",
			content:  "   \t\n  ",
			wantZero: true,
		},
		{
			name:     "edge: single character",
			content:  "x",
			wantZero: true,
		},
		{
			name:     "edge: single uppercase character",
			content:  "A",
			wantZero: true,
		},
		{
			name:     "edge: only numbers",
			content:  "1234567890 9876543210",
			wantZero: true,
		},
		{
			name:     "edge: only spaces",
			content:  "          ",
			wantZero: true,
		},
		{
			name:     "edge: newlines only",
			content:  "\n\n\n\n\n",
			wantZero: true,
		},
		{
			name:     "edge: url-like but not real URL",
			content:  "the ratio is http/https and also ftp/sftp protocols",
			wantZero: true,
		},
		{
			name:     "edge: normal content with some caps words",
			content:  "But the content is perfectly normal and reasonable text that should not trigger any spam rules.",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := ScoreContent(tt.content)
			score, rule := hitsTotal(hits)

			if tt.wantZero {
				if score != 0 {
					t.Errorf("expected score 0, got %d (rule: %s)", score, rule)
				}
				return
			}

			if score < tt.wantMin {
				t.Errorf("expected score >= %d, got %d (rule: %s)", tt.wantMin, score, rule)
			}
			if tt.wantMax > 0 && score > tt.wantMax {
				t.Errorf("expected score <= %d, got %d (rule: %s)", tt.wantMax, score, rule)
			}
			if tt.wantRule != "" && rule != tt.wantRule {
				t.Errorf("expected top rule %q, got %q (score: %d)", tt.wantRule, rule, score)
			}
		})
	}
}
