package turn

import (
	"fmt"
	"net"
	"strconv"

	"github.com/lx7/devnet/internal/auth"
	"github.com/pion/turn/v2"
)

type Server struct {
	*turn.Server

	ip    net.IP
	port  int
	realm string
}

func NewServer(ip net.IP, port int, realm string) (*Server, error) {
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

func (s *Server) auth(user string, realm string, srcAddr net.Addr) ([]byte, bool) {
	key, err := auth.UserAuthKey(user, realm)
	if err != nil || key == nil {
		return nil, false
	}
	return key, true
}
