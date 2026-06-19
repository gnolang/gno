package hook

type Notifier interface {
	Notify(caller address, message string)
}

var notifier Notifier

func SetNotifier(cur realm, next Notifier) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	notifier = next
}
