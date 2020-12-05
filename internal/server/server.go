package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/lx7/devnet/internal/auth"
	"github.com/lx7/devnet/internal/signaling"
	"github.com/lx7/devnet/transport"

	log "github.com/sirupsen/logrus"
)

// Server represents the http signaling server.
type Server struct {
	*http.Server
	sw *signaling.Switch
}

// New returns a new Server instance.
func New(addr string) *Server {
	s := &Server{
		&http.Server{
			Addr: addr,
		},
		signaling.NewSwitch(),
	}
	return s
}

// Serve starts listening on Addr ad serves HTTP requests. Connections to
// wspath will be upgraded to the WebSocket protocol as defined per RFC 6455.
func (s *Server) Serve(wspath string) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.sw.Run()
	}()

	http.HandleFunc(wspath, auth.BasicAuth(s.serveWs))
	http.HandleFunc("/", s.serveHome)

	if err := s.Server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	wg.Wait()
	return nil
}

// Shutdown terminates the http server.
func (s *Server) Shutdown() {
	if err := s.Server.Shutdown(context.TODO()); err != nil {
		log.Error("server shutdown: ", err)
	}
	s.sw.Shutdown()
	log.Info("server shutdown complete")
}

func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()

	socket, err := transport.Upgrade(w, r)
	if err != nil {
		log.Error("connection upgrade: ", err)
		return
	}

	s.sw.Attach(socket, user)
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}
