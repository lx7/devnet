package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/lx7/devnet/internal/signaling"
	"github.com/lx7/devnet/internal/transport"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	*http.Server

	wg *sync.WaitGroup
	sw *signaling.Switch
}

func New(addr string) *Server {
	s := &Server{
		&http.Server{
			Addr: addr,
		},
		&sync.WaitGroup{},
		signaling.NewSwitch(),
	}
	return s
}

func (s *Server) Bind(wspath string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.sw.Run()
	}()

	http.HandleFunc(wspath, BasicAuth(s.serveWs))
	http.HandleFunc("/", s.serveHome)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if err := s.Server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("http server: ", err)
		}
	}()
}

func (s *Server) Serve(wspath string) {
	s.Bind(wspath)
	s.wg.Wait()
}

func (s *Server) Shutdown() {
	if err := s.Server.Shutdown(context.TODO()); err != nil {
		log.Error("server shutdown: ", err)
	}
	s.sw.Shutdown()
	s.wg.Wait()
	log.Info("server shutdown complete")
}

func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()

	socket, err := transport.Upgrade(w, r)
	if err != nil {
		log.Error("connection upgrade: ", err)
		return
	}

	client := s.sw.Attach(socket, user)
	go client.ReadPump()
	go client.WritePump()
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}
