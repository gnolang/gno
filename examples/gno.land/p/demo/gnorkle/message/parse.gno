package message

import "strings"

// ParseFunc parses a raw message and returns the message function
// type extracted from the remainder of the message.
func ParseFunc(rawMsg string) (FuncType, string) {
	funcType, remainder := parseFirstToken(rawMsg)
	return FuncType(funcType), remainder
}

// ParseID parses a raw message and returns the ID extracted from
// the remainder of the message.
func ParseID(rawMsg string) (string, string) {
	return parseFirstToken(rawMsg)
}

func parseFirstToken(rawMsg string) (string, string) {
	msgParts := strings.SplitN(rawMsg, ",", 2)
	if len(msgParts) < 2 {
		return msgParts[0], ""
	}

	return msgParts[0], msgParts[1]
}
