package mock

import abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"

type (
	EchoSyncDelegate  func(string) (abci.ResponseEcho, error)
	InfoSyncDelegate  func(abci.RequestInfo) (abci.ResponseInfo, error)
	QuerySyncDelegate func(abci.RequestQuery) (abci.ResponseQuery, error)
	ErrorDelegate     func() error
)

type AppConn struct {
	EchoSyncFn  EchoSyncDelegate
	InfoSyncFn  InfoSyncDelegate
	QuerySyncFn QuerySyncDelegate
	ErrorFn     ErrorDelegate
}

func (m *AppConn) EchoSync(msg string) (abci.ResponseEcho, error) {
	if m.EchoSyncFn != nil {
		return m.EchoSyncFn(msg)
	}

	return abci.ResponseEcho{}, nil
}

func (m *AppConn) InfoSync(info abci.RequestInfo) (abci.ResponseInfo, error) {
	if m.InfoSyncFn != nil {
		return m.InfoSyncFn(info)
	}

	return abci.ResponseInfo{}, nil
}

func (m *AppConn) QuerySync(query abci.RequestQuery) (abci.ResponseQuery, error) {
	if m.QuerySyncFn != nil {
		return m.QuerySyncFn(query)
	}

	return abci.ResponseQuery{}, nil
}

func (m *AppConn) Error() error {
	if m.ErrorFn != nil {
		return m.ErrorFn()
	}

	return nil
}
