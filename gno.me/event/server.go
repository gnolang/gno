package event

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	eventCreator Creator
	eventApplier Applier
	server       *http.Server
	done         chan struct{}
}

func NewServer(creator Creator, applier Applier, done chan struct{}) *Server {
	s := Server{
		eventCreator: creator,
		eventApplier: applier,
		done:         done,
	}

	server = s
	return &s
}

// TODO: need some type of monitoring for if this fails
func (s *Server) Start(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", handleEvents)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	s.server = srv

	go func() {
		err := srv.ListenAndServe()
		fmt.Println("event server stopped:", err)
		close(s.done)
	}()
}

func (s Server) Stop() {
	// TODO: should finish all current requests before shutting down.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	s.server.Shutdown(ctx)
}
