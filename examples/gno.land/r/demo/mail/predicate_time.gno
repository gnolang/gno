package mail

import "time"

type TimeBefore struct{ t time.Time } // <=
type TimeAfter struct{ t time.Time }  // >=
type TimeBetween struct {
	min time.Time
	max time.Time
}

func (o TimeBefore) Satisfy(mail *Mail) bool {
	return !mail.time.After(o.t)
}

func (o TimeAfter) Satisfy(mail *Mail) bool {
	return !mail.time.Before(o.t)
}

func (o TimeBetween) Satisfy(mail *Mail) bool {
	return !mail.time.Before(o.min) && !mail.time.After(o.max)
}
