package mail

type Not struct {
	pred Predicate
}

func (o Not) Satisfy(mail *Mail) bool {
	return !mail.Satisfies([]Predicate{o.pred})
}
