package std

type Event interface {
	AddAttribute(key, value string)
}

type EventAttribute interface{}
