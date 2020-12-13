package proto

import "github.com/pion/webrtc/v2"

func ToPion(p *Frame_Sdp) webrtc.SessionDescription {
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(p.Sdp.Type),
		SDP:  p.Sdp.Desc,
	}
}

func WithPion(s webrtc.SessionDescription) *Frame_Sdp {
	return &Frame_Sdp{&SDP{
		Type: SDP_Type(s.Type),
		Desc: s.SDP,
	}}
}
