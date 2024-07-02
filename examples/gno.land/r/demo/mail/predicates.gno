package mail

import (
	"std"
	"strings"
	"time"
)

type Predicate interface {
	Satisfy(mail *Mail) bool
}

type (
	SenderIs       struct{ sender std.Address }
	RecipientIs    struct{ recipient std.Address }
	AndPred        struct{ a []Predicate }
	Contains       struct{ s string }
	ContainsEither struct{ a []string }
	TopicContains  struct{ s string }
	Not            struct{ pred Predicate }
	TimeBefore     struct{ t time.Time } // <=
	TimeAfter      struct{ t time.Time } // >=
	TimeBetween    struct {
		min time.Time
		max time.Time
	}
)

// `And(Contains("virus"), TimeBefore(covid))`
func And(preds ...Predicate) Predicate { return AndPred{preds} }

func (o AndPred) Satisfy(mail *Mail) bool     { return mail.Satisfies(o.a) }
func (x SenderIs) Satisfy(mail *Mail) bool    { return mail.sender == x.sender }
func (x RecipientIs) Satisfy(mail *Mail) bool { return mail.recipient == x.recipient }

// searches are always case-insensitive, and within Topic+Body
// You may combine if neeeded: Not{Contains} or And{Not{Contains}, Not{Contains}}
func (o Contains) Satisfy(mail *Mail) bool { return mail.Contains(o.s) }
func (o Not) Satisfy(mail *Mail) bool      { return !mail.Satisfies([]Predicate{o.pred}) }

func (o TopicContains) Satisfy(mail *Mail) bool {
	topic := strings.ToLower(mail.topic)
	return strings.Contains(topic, strings.ToLower(o.s))
}

func (o ContainsEither) Satisfy(mail *Mail) bool {
	for _, substring := range o.a {
		if mail.Contains(substring) {
			return true
		}
	}
	return false
}

func (o TimeBefore) Satisfy(mail *Mail) bool { return !mail.time.After(o.t) }
func (o TimeAfter) Satisfy(mail *Mail) bool  { return !mail.time.Before(o.t) }
func (o TimeBetween) Satisfy(mail *Mail) bool {
	return !mail.time.Before(o.min) && !mail.time.After(o.max)
}
