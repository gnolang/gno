package mail

type Predicate interface {
	Satisfy(mail *Mail) bool
}
