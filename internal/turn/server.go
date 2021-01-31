package turn

import (
	"fmt"
	"net"
	"strconv"

	"github.com/lx7/devnet/internal/auth"
	"github.com/pion/turn/v2"
	"github.com/rs/zerolog/log"
)

type Server struct {
	*turn.Server

	ip    net.IP
	port  int
	realm string
}

func NewServer(ip net.IP, port int, realm string) (*Server, error) {
	log.Info().
		Stringer("ip", ip).
		Int("port", port).
		Msg("starting turn server")

	listener, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("failed to create udp listener: %v", err)
	}

	s := &Server{
		ip:    ip,
		port:  port,
		realm: realm,
	}

	s.Server, err = turn.NewServer(turn.ServerConfig{
		Realm:       realm,
		AuthHandler: s.auth,
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: listener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: ip,
					Address:      "0.0.0.0",
				},
			},
		},
		LoggerFactory: LoggerFactory{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start turn server: %v", err)
	}

	return s, nil
}

func (s *Server) Close() error {
	err := s.Server.Close()
	if err != nil {
		log.Error().Err(err).Msg("turn server shutdown")
		return err
	}
	log.Info().Msg("turn server shutdown complete")
	return nil
}

func (s *Server) auth(u string, realm string, src net.Addr) ([]byte, bool) {
	key, err := auth.UserAuthKey(u, realm)
	if err != nil {
		log.Warn().Err(err).Msg("authentication failed")
		return nil, false
	}
	return key, true
}
