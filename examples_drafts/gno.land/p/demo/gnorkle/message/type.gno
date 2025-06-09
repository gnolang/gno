package message

// FuncType is the type of function that is being called by the agent.
type FuncType string

const (
	// FuncTypeIngest means the agent is sending data for ingestion.
	FuncTypeIngest FuncType = "ingest"
	// FuncTypeCommit means the agent is requesting a feed commit the transitive data
	// being held by its ingester.
	FuncTypeCommit FuncType = "commit"
	// FuncTypeRequest means the agent is requesting feed definitions for all those
	// that it is whitelisted to provide data for.
	FuncTypeRequest FuncType = "request"
)
