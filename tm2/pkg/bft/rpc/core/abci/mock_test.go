package abci

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

type (
	errorDelegate     func() error
	echoSyncDelegate  func(string) (abci.ResponseEcho, error)
	infoSyncDelegate  func(abci.RequestInfo) (abci.ResponseInfo, error)
	querySyncDelegate func(abci.RequestQuery) (abci.ResponseQuery, error)
)

type mockQuery struct {
	errorFn     errorDelegate
	echoSyncFn  echoSyncDelegate
	infoSyncFn  infoSyncDelegate
	querySyncFn querySyncDelegate
}

func (m *mockQuery) Error() error {
	if m.errorFn != nil {
		return m.errorFn()
	}

	return nil
}

func (m *mockQuery) EchoSync(msg string) (abci.ResponseEcho, error) {
	if m.echoSyncFn != nil {
		return m.echoSyncFn(msg)
	}

	return abci.ResponseEcho{}, nil
}

func (m *mockQuery) InfoSync(info abci.RequestInfo) (abci.ResponseInfo, error) {
	if m.infoSyncFn != nil {
		return m.infoSyncFn(info)
	}

	return abci.ResponseInfo{}, nil
}

func (m *mockQuery) QuerySync(query abci.RequestQuery) (abci.ResponseQuery, error) {
	if m.querySyncFn != nil {
		return m.querySyncFn(query)
	}

	return abci.ResponseQuery{}, nil
}
