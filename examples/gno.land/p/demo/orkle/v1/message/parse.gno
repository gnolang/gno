package message

import "strings"

func ParseFunc(rawMsg string) (FuncType, string) {
	msgParts := strings.SplitN(rawMsg, ",", 2)
	if len(msgParts) < 2 {
		return FuncType(msgParts[0]), ""
	}

	return FuncType(msgParts[0]), msgParts[1]
}

func ParseID(rawMsg string) (string, string) {
	msgParts := strings.SplitN(rawMsg, ",", 2)
	if len(msgParts) < 2 {
		return msgParts[0], ""
	}

	return msgParts[0], msgParts[1]
}
