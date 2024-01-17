package message

import "strings"

type FuncType string

const (
	FuncTypeIngest FuncType = "ingest"
	FuncTypeCommit FuncType = "commit"
)

func ParseFunc(rawMsg string) (FuncType, string) {
	msgParts := strings.Split(rawMsg, ",")
	if len(msgParts) < 2 {
		return FuncType(msgParts[0]), ""
	}

	return FuncType(msgParts[0]), msgParts[1]
}
