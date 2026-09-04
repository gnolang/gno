package hook

type Notifier interface {
	// Fixed: pass inert identity data, not a forgeable realm capability.
	Notify(caller address, message string)
}

var notifier Notifier

func SetNotifier(cur realm, next Notifier) {
	notifier = next
}
