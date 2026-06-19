package hook

type Notifier interface {
	Notify(cur realm, message string)
}

var notifier Notifier

func SetNotifier(cur realm, next Notifier) {
	notifier = next
}
