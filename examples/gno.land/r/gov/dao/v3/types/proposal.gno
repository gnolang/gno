package types

import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao/proxy"
)


func (p *proxy.Proposal) String() string {
	return ufmt.Sprintf(`
Title: %s

Proposed by: %s

%s

This proposal contains the following metadata:

%s`, p.Title, p.Author, p.Description, p.Executor.String())
}

func (e *proxy.Executor) String() string {
	if e.metadata == nil {
		return "No metadata available."
	}

	return e.metadata.String()
}
