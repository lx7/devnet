package signaling

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/justinas/alice"
	"github.com/lx7/devnet/internal/auth"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Server represents the http signaling server.
type Server struct {
	*http.Server
	conf     *viper.Viper
	upgrader websocket.Upgrader
	sw       Switch
}

// NewServer returns a new Server instance.
func NewServer(conf *viper.Viper) *Server {
	auth.Configure(conf.Sub("auth"))

	s := &Server{
		Server: &http.Server{
			Addr: conf.GetString("signaling.addr"),
		},
		conf: conf,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  0,
			WriteBufferSize: 0,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		sw: NewSwitch(),
	}
	return s
}

// Serve starts listening and serves HTTP requests. Connections will be
// upgraded to the WebSocket protocol as defined per RFC 6455.
func (s *Server) Serve() error {
	wspath := s.conf.GetString("signaling.wspath")
	log.Info().Msgf("starting server on %v", s.Addr)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.sw.Run()
	}()

	c := alice.New()
	c = c.Append(hlog.NewHandler(log.Logger))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, st, si int, d time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", st).
			Int("size", si).
			Dur("duration", d).
			Msg("REQ")
	}))
	c = c.Append(hlog.RemoteAddrHandler("src"))
	c = c.Append(auth.BasicAuth)
	http.Handle("/", c.Then(http.HandlerFunc(s.serveOK)))
	http.Handle(wspath, c.Then(http.HandlerFunc(s.serveWS)))

	if s.conf.GetBool("signaling.tls") {
		crt := s.conf.GetString("signaling.tls_crt")
		key := s.conf.GetString("signaling.tls_key")
		if err := s.Server.ListenAndServeTLS(crt, key); err != http.ErrServerClosed {
			return err
		}
	} else {
		if err := s.Server.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
	}
	wg.Wait()
	return nil
}

// Shutdown terminates the http server.
func (s *Server) Shutdown() {
	if err := s.Server.Shutdown(context.TODO()); err != nil {
		log.Error().Err(err).Msg("server shutdown")
	}
	s.sw.Shutdown()
	log.Info().Msg("server shutdown complete")
}

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	user, _, ok := r.BasicAuth()
	if !ok {
		log.Error().Msg("request without user")
		code := http.StatusUnauthorized
		http.Error(w, http.StatusText(code), code)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("upgrade")
		return
	}
	c := NewClient(conn, user)

	err = c.Configure(s.conf.Sub("client"))
	if err != nil {
		log.Error().Err(err).Msg("configure client")
		code := http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
		return
	}

	s.sw.Register(c)
}

func (s *Server) serveOK(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}
