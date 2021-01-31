package proto

import "github.com/pion/webrtc/v3"

func (c *Config_WebRTC) ICEServers() []webrtc.ICEServer {
	servers := []webrtc.ICEServer{}

	for _, s := range c.Iceservers {
		ic := webrtc.ICEServer{
			URLs: []string{s.Url},
		}
		servers = append(servers, ic)
	}

	return servers
}
