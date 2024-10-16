package mail

// A private structure to hold mails
// current impl. is an array
// RFC Would a ringbuffer with a certain capacity, and a "keep" feature
//     be better, I wonder.

type container struct {
	mails []*Mail
}

func newContainer() *container              { return &container{[]*Mail{}} }
func newContainerWith(a []*Mail) *container { return &container{a} }

func (ctr *container) push(mail *Mail) {
	ctr.mails = append(ctr.mails, mail)
}

func (ctr *container) filter(preds ...Predicate) []*Mail {
	a := []*Mail{}
	for _, mail := range ctr.mails {
		if mail.Satisfies(preds) {
			a = append(a, mail)
		}
	}
	return a
}
