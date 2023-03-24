package mail

import "strings"

// searches are always case-insensitive, and within Topic+Body
// You may combine if neeeded: Not{Contains} or And{Not{Contains}, Not{Contains}}

type Contains struct{ s string }
type ContainsEither struct{ a []string }
type TopicContains struct{ s string }

func (o Contains) Satisfy(mail *Mail) bool {
	return mail.Contains(o.s)
}

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
