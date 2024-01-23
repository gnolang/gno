package message

type FuncType string

const (
	FuncTypeIngest  FuncType = "ingest"
	FuncTypeCommit  FuncType = "commit"
	FuncTypeRequest FuncType = "request"
)
