package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lx7/devnet/internal/auth"

	log "github.com/sirupsen/logrus"
)

// Server represents the http signaling server.
type Server struct {
	*http.Server
	upgrader websocket.Upgrader
	sw       Switch
}

// New returns a new Server instance.
func New(addr string) *Server {
	s := &Server{
		&http.Server{
			Addr: addr,
		},
		websocket.Upgrader{
			ReadBufferSize:  0,
			WriteBufferSize: 0,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		NewSwitch(),
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
	user, _, ok := r.BasicAuth()
	if !ok {
		httpError(w, http.StatusUnauthorized, nil)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	log.Info("upgraded connection from ", conn.RemoteAddr())
	c := NewClient(conn, user)
	s.sw.Register(c)
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func httpError(w http.ResponseWriter, code int, err error) {
	text := http.StatusText(code)
	log.Errorf("http error %v: %v: %v", code, text, err)
	http.Error(w, text, code)
}
