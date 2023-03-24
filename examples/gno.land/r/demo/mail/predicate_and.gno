package mail

// `And(Contains("virus"), TimeBefore(covid))`
func And(preds ...Predicate) Predicate {
	return andPred{preds}
}

type andPred struct{ a []Predicate }

func (o andPred) Satisfy(mail *Mail) bool {
	return mail.Satisfies(o.a)
}
