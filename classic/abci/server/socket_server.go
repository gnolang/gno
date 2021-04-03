package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"

	abci "github.com/tendermint/classic/abci/types"
	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/go-amino-x"
)

// var maxNumberConnections = 2

type SocketServer struct {
	cmn.BaseService

	proto    string
	addr     string
	listener net.Listener

	connsMtx   sync.Mutex
	conns      map[int]net.Conn
	nextConnID int

	appMtx sync.Mutex
	app    abci.Application
}

func NewSocketServer(protoAddr string, app abci.Application) cmn.Service {
	proto, addr := cmn.ProtocolAndAddress(protoAddr)
	s := &SocketServer{
		proto:    proto,
		addr:     addr,
		listener: nil,
		app:      app,
		conns:    make(map[int]net.Conn),
	}
	s.BaseService = *cmn.NewBaseService(nil, "ABCIServer", s)
	return s
}

func (s *SocketServer) OnStart() error {
	if err := s.BaseService.OnStart(); err != nil {
		return err
	}
	ln, err := net.Listen(s.proto, s.addr)
	if err != nil {
		return err
	}
	s.listener = ln
	go s.acceptConnectionsRoutine()
	return nil
}

func (s *SocketServer) OnStop() {
	s.BaseService.OnStop()
	if err := s.listener.Close(); err != nil {
		s.Logger.Error("Error closing listener", "err", err)
	}

	s.connsMtx.Lock()
	defer s.connsMtx.Unlock()
	for id, conn := range s.conns {
		delete(s.conns, id)
		if err := conn.Close(); err != nil {
			s.Logger.Error("Error closing connection", "id", id, "conn", conn, "err", err)
		}
	}
}

func (s *SocketServer) addConn(conn net.Conn) int {
	s.connsMtx.Lock()
	defer s.connsMtx.Unlock()

	connID := s.nextConnID
	s.nextConnID++
	s.conns[connID] = conn

	return connID
}

// deletes conn even if close errs
func (s *SocketServer) rmConn(connID int) error {
	s.connsMtx.Lock()
	defer s.connsMtx.Unlock()

	conn, ok := s.conns[connID]
	if !ok {
		return fmt.Errorf("Connection %d does not exist", connID)
	}

	delete(s.conns, connID)
	return conn.Close()
}

func (s *SocketServer) acceptConnectionsRoutine() {
	for {
		// Accept a connection
		s.Logger.Info("Waiting for new connection...")
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.IsRunning() {
				return // Ignore error from listener closing.
			}
			s.Logger.Error("Failed to accept connection: " + err.Error())
			continue
		}

		s.Logger.Info("Accepted a new connection")

		connID := s.addConn(conn)

		closeConn := make(chan error, 2)            // Push to signal connection closed
		responses := make(chan abci.Response, 1000) // A channel to buffer responses

		// Read requests from conn and deal with them
		go s.handleRequests(closeConn, conn, responses)
		// Pull responses from 'responses' and write them to conn.
		go s.handleResponses(closeConn, conn, responses)

		// Wait until signal to close connection
		go s.waitForClose(closeConn, connID)
	}
}

func (s *SocketServer) waitForClose(closeConn chan error, connID int) {
	err := <-closeConn
	switch {
	case err == io.EOF:
		s.Logger.Error("Connection was closed by client")
	case err != nil:
		s.Logger.Error("Connection error", "error", err)
	default:
		// never happens
		s.Logger.Error("Connection was closed.")
	}

	// Close the connection
	if err := s.rmConn(connID); err != nil {
		s.Logger.Error("Error in closing connection", "error", err)
	}
}

// Read requests from conn and deal with them
func (s *SocketServer) handleRequests(closeConn chan error, conn net.Conn, responses chan<- abci.Response) {
	var count int
	var bufReader = bufio.NewReader(conn)

	defer func() {
		// make sure to recover from any app-related panics to allow proper socket cleanup
		r := recover()
		if r != nil {
			closeConn <- fmt.Errorf("recovered from panic: %v", r)
			s.appMtx.Unlock()
		}
	}()

	for {

		var req abci.Request
		_, err := amino.UnmarshalSizedReader(bufReader, &req, 0)
		if err != nil {
			if err == io.EOF {
				closeConn <- err
			} else {
				closeConn <- fmt.Errorf("error reading message: %v", err)
			}
			return
		}
		s.appMtx.Lock()
		count++
		s.handleRequest(req, responses)
		s.appMtx.Unlock()
	}
}

func (s *SocketServer) handleRequest(req abci.Request, responses chan<- abci.Response) {
	switch req := req.(type) {
	case abci.RequestEcho:
		responses <- abci.ResponseEcho{Message: req.Message}
	case abci.RequestFlush:
		responses <- abci.ResponseFlush{}
	case abci.RequestInfo:
		res := s.app.Info(req)
		responses <- res
	case abci.RequestSetOption:
		res := s.app.SetOption(req)
		responses <- res
	case abci.RequestDeliverTx:
		res := s.app.DeliverTx(req)
		responses <- res
	case abci.RequestCheckTx:
		res := s.app.CheckTx(req)
		responses <- res
	case abci.RequestCommit:
		res := s.app.Commit()
		responses <- res
	case abci.RequestQuery:
		res := s.app.Query(req)
		responses <- res
	case abci.RequestInitChain:
		res := s.app.InitChain(req)
		responses <- res
	case abci.RequestBeginBlock:
		res := s.app.BeginBlock(req)
		responses <- res
	case abci.RequestEndBlock:
		res := s.app.EndBlock(req)
		responses <- res
	default:
		responses <- abci.ResponseException{
			ResponseBase: abci.ResponseBase{
				Error: abci.StringError("Unknown request"),
			},
		}
	}
}

// Pull responses from 'responses' and write them to conn.
func (s *SocketServer) handleResponses(closeConn chan error, conn net.Conn, responses <-chan abci.Response) {
	var count int
	var bufWriter = bufio.NewWriter(conn)
	for {
		var res abci.Response = <-responses
		_, err := amino.MarshalAnySizedWriter(bufWriter, res)
		if err != nil {
			closeConn <- fmt.Errorf("Error writing message: %v", err.Error())
			return
		}
		if _, ok := res.(abci.ResponseFlush); ok {
			err = bufWriter.Flush()
			if err != nil {
				closeConn <- fmt.Errorf("Error flushing write buffer: %v", err.Error())
				return
			}
		}
		count++
	}
}
