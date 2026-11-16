package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/testsession"

	"github.com/gnolang/gno/contribs/gnobro/pkg/browser"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type fakeSession struct {
	ssh.Session
	addr     net.Addr
	stderr   bytes.Buffer
	exit     int
	hangup   chan struct{} // closed when the client leaves
	returned chan struct{} // closed when the handler returns
}

func (s *fakeSession) RemoteAddr() net.Addr  { return s.addr }
func (s *fakeSession) Stderr() io.ReadWriter { return &s.stderr }
func (s *fakeSession) Exit(code int) error   { s.exit = code; return nil }

func TestLimitSessionsMiddleware(t *testing.T) {
	served := make(chan struct{})
	h := LimitSessionsMiddleware()(func(s ssh.Session) {
		served <- struct{}{}
		<-s.(*fakeSession).hangup
	})

	port := 0
	open := func(ip string) (*fakeSession, bool) {
		port++ // every session comes from its own port
		s := &fakeSession{
			addr:     &net.TCPAddr{IP: net.ParseIP(ip), Port: port},
			hangup:   make(chan struct{}),
			returned: make(chan struct{}),
		}
		go func() { h(s); close(s.returned) }()
		select {
		case <-served:
			return s, true
		case <-s.returned:
			if s.exit != 1 || s.stderr.Len() == 0 {
				t.Errorf("%s refused with exit %d, stderr %q", s.addr, s.exit, s.stderr.String())
			}
			return s, false
		}
	}
	leave := func(s *fakeSession) { close(s.hangup); <-s.returned }

	host := func(i int) string { return fmt.Sprintf("127.0.0.%d", 1+i) }

	var last *fakeSession
	for range maxSessionsPerHost {
		s, ok := open(host(0))
		if !ok {
			t.Fatal("a session from a host refused under the per-host limit")
		}
		last = s
	}
	if _, ok := open(host(0)); ok {
		t.Fatal("a session from a host served past the per-host limit")
	}
	leave(last)
	if _, ok := open(host(0)); !ok {
		t.Fatal("a session that left kept its slot")
	}
	for i := range maxSessionsPerHost { // the addresses differ past the /64 prefix
		if _, ok := open(fmt.Sprintf("2001:db8::%x:0:0:1", 1+i)); !ok {
			t.Fatal("a session from an IPv6 /64 refused under the per-host limit")
		}
	}
	if _, ok := open("2001:db8::ffff:0:0:1"); ok {
		t.Fatal("a session from an IPv6 /64 served past the per-host limit")
	}
	if _, ok := open("2001:db8:0:1::1"); !ok {
		t.Fatal("the next IPv6 /64 refused under the per-host limit")
	}
	for i := 2*maxSessionsPerHost + 1; i < maxSessions; i++ {
		if _, ok := open(host(i)); !ok {
			t.Fatal("another host refused under the total limit")
		}
	}
	if _, ok := open(host(maxSessions)); ok {
		t.Fatal("a session served past the total limit")
	}
}

// startServer runs runServer on a free local port until the test ends.
func startServer(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()

	rpc, err := client.NewHTTPClient("127.0.0.1:1") // nothing listens, so the page fails to load
	if err != nil {
		t.Fatal(err)
	}
	cfg := &broCfg{sshListener: addr, sshHostKeyPath: filepath.Join(t.TempDir(), "key")}
	cio := commands.NewTestIO()
	cio.SetOut(commands.WriteNopCloser(io.Discard))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() { done <- runServer(ctx, &gnoclient.Client{RPCClient: rpc}, cfg, browser.DefaultConfig(), cio) }()
	t.Cleanup(func() { cancel(); <-done })
	return addr
}

type terminal struct {
	stdin  io.Writer
	stdout io.Reader
	stderr *bytes.Buffer
	wait   chan error
	resize func(h, w int) error
}

// openTerminal opens a session declaring an 80x24 terminal and starts its shell.
func openTerminal(t *testing.T, addr string) terminal {
	t.Helper()
	sess, err := testsession.NewClientSession(t, addr, nil)
	for start := time.Now(); err != nil && time.Since(start) < 5*time.Second; {
		time.Sleep(10 * time.Millisecond) // the server is still starting
		sess, err = testsession.NewClientSession(t, addr, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	term := terminal{stderr: &bytes.Buffer{}, wait: make(chan error, 1), resize: sess.WindowChange}
	sess.Stderr = term.stderr
	if term.stdout, err = sess.StdoutPipe(); err != nil {
		t.Fatal(err)
	}
	if term.stdin, err = sess.StdinPipe(); err != nil {
		t.Fatal(err)
	}
	if err := sess.RequestPty("xterm", 24, 80, nil); err != nil {
		t.Fatal(err)
	}
	if err := sess.Shell(); err != nil {
		t.Fatal(err)
	}
	go func() { term.wait <- sess.Wait() }()
	return term
}

// The limit has to run before bubbletea.Middleware, so that a refused session
// never starts a browser.
func TestRunServerRefusesPastSessionLimit(t *testing.T) {
	addr := startServer(t)
	for range maxSessionsPerHost {
		term := openTerminal(t, addr)
		if _, err := term.stdout.Read(make([]byte, 1)); err != nil { // the renderer queried the terminal, so the slot is taken
			t.Fatal(err)
		}
	}
	term := openTerminal(t, addr)
	select {
	case <-term.wait:
		if !strings.Contains(term.stderr.String(), "too many sessions") {
			t.Fatalf("refused with stderr %q", term.stderr.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("a session served past the per-host limit")
	}
}

func TestRunServerReleasesSessions(t *testing.T) {
	addr := startServer(t)
	left := func() (requests, zones int) {
		buf := make([]byte, 1<<22)
		stacks := string(buf[:runtime.Stack(buf, true)])
		return strings.Count(stacks, "(*session).handleRequests("), strings.Count(stacks, ").zoneWorker(")
	}
	settled := func(what string) {
		t.Helper()
		for start := time.Now(); ; time.Sleep(50 * time.Millisecond) {
			requests, zones := left()
			if requests == 0 && zones == 0 {
				return
			}
			if time.Since(start) > 5*time.Second {
				t.Fatalf("%s, %d request loops and %d zone workers are left", what, requests, zones)
			}
		}
	}
	ended := func(term terminal) {
		t.Helper()
		select {
		case <-term.wait:
		case <-time.After(10 * time.Second):
			t.Fatal("a session did not end")
		}
	}
	settled("before any session opened") // sessions of earlier tests end first

	var served []terminal
	for range maxSessionsPerHost {
		term := openTerminal(t, addr)
		if _, err := term.stdout.Read(make([]byte, 1)); err != nil {
			t.Fatal(err)
		}
		go io.Copy(io.Discard, term.stdout) // a browser blocked on its output never reads the ctrl+c
		served = append(served, term)
	}
	for range 3 {
		term := openTerminal(t, addr)
		term.resize(30, 100) // two window changes the refused session never reads, one past the slot the pty size fills
		term.resize(31, 101)
		ended(term)
	}
	for _, term := range served {
		term.resize(30, 100)
		for start := time.Now(); ; {
			term.stdin.Write([]byte{0x03}) // ctrl+c, sent again in case a terminal query read the last one
			select {
			case <-term.wait:
			case <-time.After(100 * time.Millisecond):
				if time.Since(start) < 10*time.Second {
					continue
				}
				t.Fatal("a browser did not quit on ctrl+c")
			}
			break
		}
	}
	settled("after every session ended")
}
