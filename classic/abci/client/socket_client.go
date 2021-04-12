package abcicli

import (
	"bufio"
	"container/list"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/tendermint/classic/abci/types"
	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/go-amino-x"
)

const reqQueueSize = 256 // TODO make configurable
// const maxResponseSize = 1048576 // 1MB TODO make configurable
const flushThrottleMS = 20 // Don't wait longer than...

var _ Client = (*socketClient)(nil)

// This is goroutine-safe, but users should beware that
// the application in general is not meant to be interfaced
// with concurrent callers.
type socketClient struct {
	cmn.BaseService

	addr        string
	mustConnect bool
	conn        net.Conn

	reqQueue   chan *ReqRes
	flushTimer *cmn.ThrottleTimer

	mtx     sync.Mutex
	err     error
	reqSent *list.List                        // list of requests sent, waiting for response
	resCb   func(abci.Request, abci.Response) // called on all requests, if set.

}

func NewSocketClient(addr string, mustConnect bool) *socketClient {
	cli := &socketClient{
		reqQueue:    make(chan *ReqRes, reqQueueSize),
		flushTimer:  cmn.NewThrottleTimer("socketClient", flushThrottleMS),
		mustConnect: mustConnect,

		addr:    addr,
		reqSent: list.New(),
		resCb:   nil,
	}
	cli.BaseService = *cmn.NewBaseService(nil, "socketClient", cli)
	return cli
}

func (cli *socketClient) OnStart() error {
	var err error
	var conn net.Conn
RETRY_LOOP:
	for {
		conn, err = cmn.Connect(cli.addr)
		if err != nil {
			if cli.mustConnect {
				return err
			}
			cli.Logger.Error(fmt.Sprintf("abci.socketClient failed to connect to %v.  Retrying...", cli.addr), "err", err)
			time.Sleep(time.Second * dialRetryIntervalSeconds)
			continue RETRY_LOOP
		}
		cli.conn = conn

		go cli.sendRequestsRoutine(conn)
		go cli.recvResponseRoutine(conn)

		return nil
	}
}

func (cli *socketClient) OnStop() {
	if cli.conn != nil {
		cli.conn.Close()
	}

	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	cli.flushQueue()
}

// Stop the client and set the error
func (cli *socketClient) StopForError(err error) {
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	if cli.err == nil {
		cli.err = err
	}
	cli.mtx.Unlock()

	cli.Logger.Error(fmt.Sprintf("Stopping abci.socketClient for error: %v", err.Error()))
	cli.Stop()
}

func (cli *socketClient) Error() error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	return cli.err
}

// Set listener for all responses
// NOTE: callback may get internally generated flush responses.
func (cli *socketClient) SetResponseCallback(resCb Callback) {
	cli.mtx.Lock()
	cli.resCb = resCb
	cli.mtx.Unlock()
}

//----------------------------------------

func (cli *socketClient) sendRequestsRoutine(conn net.Conn) {

	w := bufio.NewWriter(conn)
	for {
		select {
		case <-cli.flushTimer.Ch:
			select {
			case cli.reqQueue <- NewReqRes(abci.RequestFlush{}):
			default:
				// Probably will fill the buffer, or retry later.
			}
		case <-cli.Quit():
			return
		case reqres := <-cli.reqQueue:
			cli.willSendReq(reqres)
			var req abci.Request = reqres.Request
			_, err := amino.MarshalAnySizedWriter(w, req)
			if err != nil {
				cli.StopForError(fmt.Errorf("Error writing msg: %v", err))
				return
			}
			// cli.Logger.Debug("Sent request", "requestType", reflect.TypeOf(reqres.Request), "request", reqres.Request)
			if _, ok := reqres.Request.(abci.RequestFlush); ok {
				err = w.Flush()
				if err != nil {
					cli.StopForError(fmt.Errorf("Error flushing writer: %v", err))
					return
				}
			}
		}
	}
}

func (cli *socketClient) recvResponseRoutine(conn net.Conn) {

	r := bufio.NewReader(conn) // Buffer reads
	for {
		var res abci.Response
		_, err := amino.UnmarshalSizedReader(r, &res, 0)
		if err != nil {
			cli.StopForError(err)
			return
		}
		switch res := res.(type) {
		case abci.ResponseException:
			// XXX After setting cli.err, release waiters (e.g. reqres.Done())
			cli.StopForError(res.Error)
			return
		default:
			// cli.Logger.Debug("Received response", "responseType", reflect.TypeOf(res), "response", res)
			err := cli.didRecvResponse(res)
			if err != nil {
				cli.StopForError(err)
				return
			}
		}
	}
}

func (cli *socketClient) willSendReq(reqres *ReqRes) {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	cli.reqSent.PushBack(reqres)
}

func (cli *socketClient) didRecvResponse(res abci.Response) error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// Get the first ReqRes
	next := cli.reqSent.Front()
	if next == nil {
		return fmt.Errorf("Unexpected result type %v when nothing expected", reflect.TypeOf(res))
	}
	reqres := next.Value.(*ReqRes)
	if !resMatchesReq(reqres.Request, res) {
		return fmt.Errorf("Unexpected result type %v when response to %v expected",
			reflect.TypeOf(res), reflect.TypeOf(reqres.Request))
	}

	reqres.Response = res    // Set response
	reqres.Done()            // Release waiters
	cli.reqSent.Remove(next) // Pop first item from linked list

	// Notify client listener if set (global callback).
	if cli.resCb != nil {
		cli.resCb(reqres.Request, res)
	}

	// Notify reqRes listener if set (request specific callback).
	// NOTE: it is possible this callback isn't set on the reqres object.
	// at this point, in which case it will be called after, when it is set.
	if cb := reqres.GetCallback(); cb != nil {
		cb(res)
	}

	return nil
}

//----------------------------------------

func (cli *socketClient) EchoAsync(msg string) *ReqRes {
	return cli.queueRequest(abci.RequestEcho{Message: msg})
}

func (cli *socketClient) FlushAsync() *ReqRes {
	return cli.queueRequest(abci.RequestFlush{})
}

func (cli *socketClient) InfoAsync(req abci.RequestInfo) *ReqRes {
	return cli.queueRequest(req)
}

func (cli *socketClient) SetOptionAsync(req abci.RequestSetOption) *ReqRes {
	return cli.queueRequest(req)
}

func (cli *socketClient) DeliverTxAsync(req abci.RequestDeliverTx) *ReqRes {
	return cli.queueRequest(req)
}

func (cli *socketClient) CheckTxAsync(req abci.RequestCheckTx) *ReqRes {
	return cli.queueRequest(req)
}

func (cli *socketClient) QueryAsync(req abci.RequestQuery) *ReqRes {
	return cli.queueRequest(req)
}

func (cli *socketClient) CommitAsync() *ReqRes {
	return cli.queueRequest(abci.RequestCommit{})
}

func (cli *socketClient) InitChainAsync(req abci.RequestInitChain) *ReqRes {
	return cli.queueRequest(req)
}

func (cli *socketClient) BeginBlockAsync(req abci.RequestBeginBlock) *ReqRes {
	return cli.queueRequest(req)
}

func (cli *socketClient) EndBlockAsync(req abci.RequestEndBlock) *ReqRes {
	return cli.queueRequest(req)
}

//----------------------------------------

func (cli *socketClient) FlushSync() error {
	reqRes := cli.queueRequest(abci.RequestFlush{})
	if err := cli.Error(); err != nil {
		return err
	}
	reqRes.Wait() // NOTE: if we don't flush the queue, its possible to get stuck here
	return cli.Error()
}

func (cli *socketClient) EchoSync(msg string) (abci.ResponseEcho, error) {
	reqres := cli.queueRequest(abci.RequestEcho{Message: msg})
	cli.FlushSync()
	return reqres.Response.(abci.ResponseEcho), cli.Error()
}

func (cli *socketClient) InfoSync(req abci.RequestInfo) (abci.ResponseInfo, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseInfo), cli.Error()
}

func (cli *socketClient) SetOptionSync(req abci.RequestSetOption) (abci.ResponseSetOption, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseSetOption), cli.Error()
}

func (cli *socketClient) DeliverTxSync(req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseDeliverTx), cli.Error()
}

func (cli *socketClient) CheckTxSync(req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseCheckTx), cli.Error()
}

func (cli *socketClient) QuerySync(req abci.RequestQuery) (abci.ResponseQuery, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseQuery), cli.Error()
}

func (cli *socketClient) CommitSync() (abci.ResponseCommit, error) {
	reqres := cli.queueRequest(abci.RequestCommit{})
	cli.FlushSync()
	return reqres.Response.(abci.ResponseCommit), cli.Error()
}

func (cli *socketClient) InitChainSync(req abci.RequestInitChain) (abci.ResponseInitChain, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseInitChain), cli.Error()
}

func (cli *socketClient) BeginBlockSync(req abci.RequestBeginBlock) (abci.ResponseBeginBlock, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseBeginBlock), cli.Error()
}

func (cli *socketClient) EndBlockSync(req abci.RequestEndBlock) (abci.ResponseEndBlock, error) {
	reqres := cli.queueRequest(req)
	cli.FlushSync()
	return reqres.Response.(abci.ResponseEndBlock), cli.Error()
}

//----------------------------------------

func (cli *socketClient) queueRequest(req abci.Request) *ReqRes {
	reqres := NewReqRes(req)

	// TODO: set cli.err if reqQueue times out
	cli.reqQueue <- reqres

	// Maybe auto-flush, or unset auto-flush
	switch req.(type) {
	case abci.RequestFlush:
		cli.flushTimer.Unset()
	default:
		cli.flushTimer.Set()
	}

	return reqres
}

func (cli *socketClient) flushQueue() {
	// mark all in-flight messages as resolved (they will get cli.Error())
	for req := cli.reqSent.Front(); req != nil; req = req.Next() {
		reqres := req.Value.(*ReqRes)
		reqres.Done()
	}

	// mark all queued messages as resolved
LOOP:
	for {
		select {
		case reqres := <-cli.reqQueue:
			reqres.Done()
		default:
			break LOOP
		}
	}
}

//----------------------------------------

func resMatchesReq(req abci.Request, res abci.Response) (ok bool) {
	switch req.(type) {
	case abci.RequestEcho:
		_, ok = res.(abci.ResponseEcho)
	case abci.RequestFlush:
		_, ok = res.(abci.ResponseFlush)
	case abci.RequestInfo:
		_, ok = res.(abci.ResponseInfo)
	case abci.RequestSetOption:
		_, ok = res.(abci.ResponseSetOption)
	case abci.RequestDeliverTx:
		_, ok = res.(abci.ResponseDeliverTx)
	case abci.RequestCheckTx:
		_, ok = res.(abci.ResponseCheckTx)
	case abci.RequestCommit:
		_, ok = res.(abci.ResponseCommit)
	case abci.RequestQuery:
		_, ok = res.(abci.ResponseQuery)
	case abci.RequestInitChain:
		_, ok = res.(abci.ResponseInitChain)
	case abci.RequestBeginBlock:
		_, ok = res.(abci.ResponseBeginBlock)
	case abci.RequestEndBlock:
		_, ok = res.(abci.ResponseEndBlock)
	}
	return ok
}
