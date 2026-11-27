package evil

// Runner calls its argument, like a lent-to callee, but is code the victim
// never imported: reached through an interface it runs view-only.
type Runner struct{}

func (Runner) Run(f func()) { f() }
