package message

import "strings"

// ParseFunc parses a raw message and returns the message function
// type extracted from the remainder of the message.
func ParseFunc(rawMsg string) (FuncType, string) {
	msgParts := strings.SplitN(rawMsg, ",", 2)
	if len(msgParts) < 2 {
		return FuncType(msgParts[0]), ""
	}

	return FuncType(msgParts[0]), msgParts[1]
}

// ParseID parses a raw message and returns the ID extracted from
// the remainder of the message.
func ParseID(rawMsg string) (string, string) {
	msgParts := strings.SplitN(rawMsg, ",", 2)
	if len(msgParts) < 2 {
		return msgParts[0], ""
	}

	return msgParts[0], msgParts[1]
}
