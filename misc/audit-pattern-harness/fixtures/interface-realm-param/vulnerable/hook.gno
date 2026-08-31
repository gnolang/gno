package hook

type Notifier interface {
	// Vulnerable: interface implementers receive cur realm from this realm API.
	Notify(cur realm, message string)
}

var notifier Notifier

func SetNotifier(cur realm, next Notifier) {
	notifier = next
}
